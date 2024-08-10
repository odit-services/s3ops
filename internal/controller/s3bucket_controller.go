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

	"github.com/minio/minio-go/v7"
	s3oditservicesv1alpha1 "github.com/odit-services/s3ops/api/v1alpha1"
	s3client "github.com/odit-services/s3ops/internal/controller/shared"
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
		return ctrl.Result{}, nil
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
		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(s3Bucket, "s3.odit.services/finalizer") {
		controllerutil.AddFinalizer(s3Bucket, "s3.odit.services/finalizer")
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
			return ctrl.Result{}, nil
		}
	}

	s3Server := &s3oditservicesv1alpha1.S3Server{}
	err = r.Get(ctx, client.ObjectKey{
		Namespace: s3Bucket.Spec.ServerRef.Namespace,
		Name:      s3Bucket.Spec.ServerRef.Name,
	}, s3Server)
	if err != nil {
		r.logger.Errorw("Failed to get S3Server resource", "name", s3Bucket.Spec.ServerRef.Name, "namespace", s3Bucket.Spec.ServerRef.Namespace, "error", err)
		s3Bucket.Status.Conditions = append(s3Bucket.Status.Conditions, metav1.Condition{
			Type:               s3oditservicesv1alpha1.ConditionFailed,
			Status:             metav1.ConditionFalse,
			Reason:             s3oditservicesv1alpha1.ReasonNotFound,
			Message:            "S3Server resource not found",
			LastTransitionTime: metav1.Now(),
		})
		r.Status().Update(ctx, s3Bucket)
		return ctrl.Result{}, nil
	}

	minioClient, err := r.S3ClientFactory.NewClient(*s3Server)
	if err != nil {
		r.logger.Errorw("Failed to create Minio client for S3Server", "name", req.Name, "namespace", req.Namespace, "error", err)
		s3Bucket.Status.Conditions = append(s3Server.Status.Conditions, metav1.Condition{
			Type:    s3oditservicesv1alpha1.ConditionFailed,
			Status:  metav1.ConditionFalse,
			Reason:  err.Error(),
			Message: fmt.Sprintf("Failed to create Minio client: %v", err),
			LastTransitionTime: metav1.Time{
				Time: time.Now(),
			},
		})
		r.Status().Update(ctx, s3Server)
		return ctrl.Result{}, err
	}

	bucketExists, err := minioClient.BucketExists(context.Background(), s3Bucket.Name)
	if err != nil {
		r.logger.Errorw("Failed to check if bucket exists", "name", s3Bucket.Name, "error", err)
		s3Bucket.Status.Conditions = append(s3Bucket.Status.Conditions, metav1.Condition{
			Type:               s3oditservicesv1alpha1.ConditionFailed,
			Status:             metav1.ConditionFalse,
			Reason:             s3oditservicesv1alpha1.ReasonRequestFailed,
			Message:            fmt.Sprintf("Failed to check if bucket exists: %v", err),
			LastTransitionTime: metav1.Now(),
		})
		r.Status().Update(ctx, s3Bucket)
		return ctrl.Result{}, nil
	}

	if !bucketExists {
		err = minioClient.MakeBucket(context.Background(), s3Bucket.Name, minio.MakeBucketOptions{
			Region:        s3Bucket.Spec.Region,
			ObjectLocking: s3Bucket.Spec.ObjectLocking,
		})
		if err != nil {
			r.logger.Errorw("Failed to create bucket", "name", s3Bucket.Name, "error", err)
			s3Bucket.Status.Conditions = append(s3Bucket.Status.Conditions, metav1.Condition{
				Type:               s3oditservicesv1alpha1.ConditionFailed,
				Status:             metav1.ConditionFalse,
				Reason:             s3oditservicesv1alpha1.ReasonRequestFailed,
				Message:            fmt.Sprintf("Failed to create bucket: %v", err),
				LastTransitionTime: metav1.Now(),
			})
			r.Status().Update(ctx, s3Bucket)
			return ctrl.Result{}, nil
		}
	}

	r.logger.Infow("Finished reconciling s3Bucket", "name", req.Name, "namespace", req.Namespace)
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
		Complete(r)
}
