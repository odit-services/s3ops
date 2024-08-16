/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/minio/minio-go/v7"
	s3oditservicesv1alpha1 "github.com/odit-services/s3ops/api/v1alpha1"
	s3client "github.com/odit-services/s3ops/internal/services/s3client"
)

// S3BucketReconciler reconciles a S3Bucket object
type S3BucketReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	logger          *zap.SugaredLogger
	S3ClientFactory s3client.S3ClientFactory
}

// +kubebuilder:rbac:groups=s3.odit.services,resources=s3buckets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=s3.odit.services,resources=s3buckets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=s3.odit.services,resources=s3buckets/finalizers,verbs=update

func (r *S3BucketReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	r.logger.Infow("Reconciling S3Bucket", "name", req.Name, "namespace", req.Namespace)

	s3Bucket := &s3oditservicesv1alpha1.S3Bucket{}
	err := r.Get(ctx, req.NamespacedName, s3Bucket)
	if err != nil {
		r.logger.Errorw("Failed to get S3Bucket resource", "name", req.Name, "namespace", req.Namespace, "error", err)
		s3Bucket.Status.Conditions = append(s3Bucket.Status.Conditions, metav1.Condition{
			Type:               s3oditservicesv1alpha1.ConditionFailed,
			Status:             metav1.ConditionFalse,
			Reason:             s3oditservicesv1alpha1.ReasonNotFound,
			Message:            "S3Bucket resource not found",
			LastTransitionTime: metav1.Now(),
		})
		r.Status().Update(ctx, s3Bucket)
		return ctrl.Result{}, err
	}

	s3Bucket.Status.Conditions = append(s3Bucket.Status.Conditions, metav1.Condition{
		Type:               s3oditservicesv1alpha1.ConditionReconciling,
		Status:             metav1.ConditionUnknown,
		Reason:             s3oditservicesv1alpha1.ConditionReconciling,
		Message:            "Reconciling S3Bucket resource",
		LastTransitionTime: metav1.Now(),
	})
	err = r.Status().Update(ctx, s3Bucket)
	if err != nil {
		r.logger.Errorw("Failed to update S3Bucket resource status", "name", req.Name, "namespace", req.Namespace, "error", err)
		return ctrl.Result{}, err
	}

	if !controllerutil.ContainsFinalizer(s3Bucket, "s3.odit.services/bucket") {
		controllerutil.AddFinalizer(s3Bucket, "s3.odit.services/bucket")
		err := r.Update(ctx, s3Bucket)
		if err != nil {
			r.logger.Errorw("Failed to add finalizer to S3Bucket resource", "name", req.Name, "namespace", req.Namespace, "error", err)
			s3Bucket.Status.Conditions = append(s3Bucket.Status.Conditions, metav1.Condition{
				Type:               s3oditservicesv1alpha1.ConditionFailed,
				Status:             metav1.ConditionFalse,
				Reason:             s3oditservicesv1alpha1.ReasonFinalizerFailedToApply,
				Message:            "Failed to add finalizer to S3Bucket resource",
				LastTransitionTime: metav1.Now(),
			})
			r.Status().Update(ctx, s3Bucket)
			return ctrl.Result{}, err
		}
	}

	s3Client, condition, err := s3client.GetS3ClientFromS3Server(s3Bucket.Spec.ServerRef, r.S3ClientFactory, r.Client)
	if err != nil {
		r.logger.Errorw("Failed to get S3Client from S3Server", "name", req.Name, "namespace", req.Namespace, "error", err)
		s3Bucket.Status.Conditions = append(s3Bucket.Status.Conditions, condition)
		r.Status().Update(ctx, s3Bucket)
		return ctrl.Result{}, err
	}

	var bucketName string
	if s3Bucket.Status.Name != "" {
		bucketName = s3Bucket.Status.Name
	} else if !s3Bucket.Spec.DisableNameGeneration {
		r.logger.Debugw("Generating bucket name", "name", s3Bucket.Name, "namespace", s3Bucket.Namespace)
		nanoID, err := gonanoid.New(21)
		if err != nil {
			r.logger.Errorw("Failed to generate bucket name", "name", s3Bucket.Name, "error", err)
			s3Bucket.Status.Conditions = append(s3Bucket.Status.Conditions, metav1.Condition{
				Type:               s3oditservicesv1alpha1.ConditionFailed,
				Status:             metav1.ConditionFalse,
				Reason:             s3oditservicesv1alpha1.ReasonRequestFailed,
				Message:            fmt.Sprintf("Failed to generate bucket name: %v", err),
				LastTransitionTime: metav1.Now(),
			})
			r.Status().Update(ctx, s3Bucket)
			return ctrl.Result{}, err
		}
		bucketPrefix := fmt.Sprintf("%s-%s", s3Bucket.Name, s3Bucket.Namespace)
		var truncateAt int
		if len(bucketPrefix) > 39 {
			truncateAt = 39
		} else {
			truncateAt = len(bucketPrefix) - 1
		}
		bucketName = fmt.Sprintf("%s-%s", bucketPrefix[0:truncateAt], strings.ToLower(nanoID))
	} else {
		r.logger.Debugw("Using bucket name from spec", "name", s3Bucket.Name, "namespace", s3Bucket.Namespace)
		bucketName = s3Bucket.Name
	}
	s3Bucket.Status.Name = bucketName

	bucketExists, err := s3Client.BucketExists(context.Background(), bucketName)
	if err != nil {
		r.logger.Errorw("Failed to check if bucket exists", "name", s3Bucket.Name, "bucketName", bucketName, "error", err)
		s3Bucket.Status.Conditions = append(s3Bucket.Status.Conditions, metav1.Condition{
			Type:               s3oditservicesv1alpha1.ConditionFailed,
			Status:             metav1.ConditionFalse,
			Reason:             s3oditservicesv1alpha1.ReasonRequestFailed,
			Message:            fmt.Sprintf("Failed to check if bucket exists: %v", err),
			LastTransitionTime: metav1.Now(),
		})
		r.Status().Update(ctx, s3Bucket)
		return ctrl.Result{}, err
	}

	if s3Bucket.DeletionTimestamp != nil {
		r.logger.Infow("Deleting s3Bucket", "name", req.Name, "namespace", req.Namespace)
		if !bucketExists {
			r.logger.Debugw("Bucket does not exist", "name", req.Name, "namespace", req.Namespace)
		} else if s3Bucket.Spec.SoftDelete {
			r.logger.Debugw("Soft delete is enabled", "name", req.Name, "namespace", req.Namespace)
		} else {
			r.logger.Debugw("Removing bucket", "name", req.Name, "namespace", req.Namespace)
			err := s3Client.RemoveBucket(context.Background(), bucketName)
			if err != nil {
				r.logger.Errorw("Failed to remove bucket", "name", req.Name, "namespace", req.Namespace, "error", err)
				s3Bucket.Status.Conditions = append(s3Bucket.Status.Conditions, metav1.Condition{
					Type:               s3oditservicesv1alpha1.ConditionFailed,
					Status:             metav1.ConditionFalse,
					Reason:             s3oditservicesv1alpha1.ReasonRequestFailed,
					Message:            fmt.Sprintf("Failed to remove bucket: %v", err),
					LastTransitionTime: metav1.Now(),
				})
				r.Status().Update(ctx, s3Bucket)
				return ctrl.Result{}, err
			}
		}
		controllerutil.RemoveFinalizer(s3Bucket, "s3.odit.services/bucket")
		err := r.Update(ctx, s3Bucket)
		if err != nil {
			r.logger.Errorw("Failed to remove finalizer from s3Bucket resource", "name", req.Name, "namespace", req.Namespace, "error", err)
			return ctrl.Result{}, err
		}
		r.logger.Infow("Finished reconciling s3Bucket", "name", req.Name, "namespace", req.Namespace, "bucketName", bucketName)
		return ctrl.Result{}, nil
	}

	if !bucketExists {
		err = s3Client.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{
			Region:        s3Bucket.Spec.Region,
			ObjectLocking: s3Bucket.Spec.ObjectLocking,
		})
		if err != nil {
			r.logger.Errorw("Failed to create bucket", "name", s3Bucket.Name, "bucketName", bucketName, "error", err)
			s3Bucket.Status.Conditions = append(s3Bucket.Status.Conditions, metav1.Condition{
				Type:               s3oditservicesv1alpha1.ConditionFailed,
				Status:             metav1.ConditionFalse,
				Reason:             s3oditservicesv1alpha1.ReasonRequestFailed,
				Message:            fmt.Sprintf("Failed to create bucket: %v", err),
				LastTransitionTime: metav1.Now(),
			})
			r.Status().Update(ctx, s3Bucket)
			return ctrl.Result{}, err
		}
		s3Bucket.Status.Created = true
	}

	r.logger.Infow("Finished reconciling s3Bucket", "name", req.Name, "namespace", req.Namespace, "bucketName", bucketName)
	s3Bucket.Status.Conditions = append(s3Bucket.Status.Conditions, metav1.Condition{
		Type:    s3oditservicesv1alpha1.ConditionReady,
		Status:  metav1.ConditionTrue,
		Reason:  metav1.StatusSuccess,
		Message: "s3Bucket is ready",
		LastTransitionTime: metav1.Time{
			Time: time.Now(),
		},
	})
	err = r.Status().Update(ctx, s3Bucket)
	if err != nil {
		r.logger.Errorw("Failed to update s3Bucket status", "name", req.Name, "namespace", req.Namespace, "error", err)
		return ctrl.Result{}, err
	}

	return ctrl.Result{
		RequeueAfter: 5 * time.Minute,
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *S3BucketReconciler) SetupWithManager(mgr ctrl.Manager) error {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "INFO"
	}

	var zapLogLevel zapcore.Level
	err := zapLogLevel.UnmarshalText([]byte(strings.ToLower(logLevel)))
	if err != nil {
		zapLogLevel = zapcore.InfoLevel
	}

	zapConfig := zap.NewProductionConfig()
	zapConfig.Level = zap.NewAtomicLevelAt(zapLogLevel)
	zapLogger, _ := zapConfig.Build()
	defer zapLogger.Sync()
	r.logger = zapLogger.Sugar()

	r.S3ClientFactory = &s3client.S3ClientFactoryDefault{}

	return ctrl.NewControllerManagedBy(mgr).
		For(&s3oditservicesv1alpha1.S3Bucket{}).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{})).
		Complete(r)
}
