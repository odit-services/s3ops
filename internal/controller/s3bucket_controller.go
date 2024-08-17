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

func (r *S3BucketReconciler) HandleError(s3Bucket *s3oditservicesv1alpha1.S3Bucket, err error) (ctrl.Result, error) {
	r.logger.Errorw("Failed to reconcile S3Bucket", "name", s3Bucket.Name, "namespace", s3Bucket.Namespace, "error", err)
	s3Bucket.Status = s3oditservicesv1alpha1.S3BucketStatus{
		CrStatus: s3oditservicesv1alpha1.CrStatus{
			State:             s3oditservicesv1alpha1.StateFailed,
			LastAction:        s3Bucket.Status.LastAction,
			LastMessage:       fmt.Sprintf("Failed to reconcile S3Bucket: %v", err),
			LastReconcileTime: time.Now().Format(time.RFC3339),
			CurrentRetries:    s3Bucket.Status.CurrentRetries + 1,
		},
		Name:    s3Bucket.Status.Name,
		Created: s3Bucket.Status.Created,
	}
	updateErr := r.Status().Update(context.Background(), s3Bucket)
	if updateErr != nil {
		r.logger.Errorw("Failed to update S3Bucket status", "name", s3Bucket.Name, "namespace", s3Bucket.Namespace, "error", updateErr)
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
	}
	r.logger.Infow("Requeue S3Bucket", "name", s3Bucket.Name, "namespace", s3Bucket.Namespace)
	return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
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
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
	}

	s3Bucket.Status = s3oditservicesv1alpha1.S3BucketStatus{
		CrStatus: s3oditservicesv1alpha1.CrStatus{
			State:             s3oditservicesv1alpha1.StateReconciling,
			LastAction:        s3oditservicesv1alpha1.ActionUnknown,
			CurrentRetries:    s3Bucket.Status.CurrentRetries,
			LastReconcileTime: time.Now().Format(time.RFC3339),
		},
		Name:    s3Bucket.Status.Name,
		Created: s3Bucket.Status.Created,
	}
	err = r.Status().Update(ctx, s3Bucket)
	if err != nil {
		r.logger.Errorw("Failed to update S3Bucket resource status", "name", req.Name, "namespace", req.Namespace, "error", err)
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
	}

	if !controllerutil.ContainsFinalizer(s3Bucket, "s3.odit.services/bucket") {
		controllerutil.AddFinalizer(s3Bucket, "s3.odit.services/bucket")
		err := r.Update(ctx, s3Bucket)
		if err != nil {
			r.logger.Errorw("Failed to add finalizer to S3Bucket resource", "name", req.Name, "namespace", req.Namespace, "error", err)
			return r.HandleError(s3Bucket, err)
		}
	}

	s3Client, _, err := s3client.GetS3ClientFromS3Server(s3Bucket.Spec.ServerRef, r.S3ClientFactory, r.Client)
	if err != nil {
		r.logger.Errorw("Failed to get S3Client from S3Server", "name", req.Name, "namespace", req.Namespace, "error", err)
		return r.HandleError(s3Bucket, err)
	}

	var bucketName string
	if s3Bucket.Status.Name != "" {
		bucketName = s3Bucket.Status.Name
	} else if !s3Bucket.Spec.DisableNameGeneration {
		r.logger.Debugw("Generating bucket name", "name", s3Bucket.Name, "namespace", s3Bucket.Namespace)
		nanoID, err := gonanoid.Generate("abcdefghijklmnopqrstuvwxyz0123456789", 21)
		if err != nil {
			r.logger.Errorw("Failed to generate bucket name", "name", s3Bucket.Name, "error", err)
			return r.HandleError(s3Bucket, err)
		}
		bucketPrefix := fmt.Sprintf("%s-%s", s3Bucket.Name, s3Bucket.Namespace)
		var truncateAt int
		if len(bucketPrefix) > 39 {
			truncateAt = 39
		} else {
			truncateAt = len(bucketPrefix)
		}
		bucketName = fmt.Sprintf("%s-%s", bucketPrefix[0:truncateAt], nanoID)
	} else {
		r.logger.Debugw("Using bucket name from spec", "name", s3Bucket.Name, "namespace", s3Bucket.Namespace)
		bucketName = s3Bucket.Name
	}
	s3Bucket.Status.Name = bucketName

	bucketExists, err := s3Client.BucketExists(context.Background(), bucketName)
	if err != nil {
		r.logger.Errorw("Failed to check if bucket exists", "name", s3Bucket.Name, "bucketName", bucketName, "error", err)
		return r.HandleError(s3Bucket, err)
	}

	if s3Bucket.DeletionTimestamp != nil {
		r.logger.Infow("Deleting s3Bucket", "name", req.Name, "namespace", req.Namespace)
		s3Bucket.Status.LastAction = s3oditservicesv1alpha1.ActionDelete
		err = r.Status().Update(ctx, s3Bucket)
		if err != nil {
			r.logger.Errorw("Failed to update s3Bucket status", "name", req.Name, "namespace", req.Namespace, "error", err)
			return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
		}

		if !bucketExists {
			r.logger.Debugw("Bucket does not exist", "name", req.Name, "namespace", req.Namespace)
		} else if s3Bucket.Spec.SoftDelete {
			r.logger.Debugw("Soft delete is enabled", "name", req.Name, "namespace", req.Namespace)
		} else {
			r.logger.Debugw("Removing bucket", "name", req.Name, "namespace", req.Namespace)
			err := s3Client.RemoveBucket(context.Background(), bucketName)
			if err != nil {
				r.logger.Errorw("Failed to remove bucket", "name", req.Name, "namespace", req.Namespace, "error", err)
				return r.HandleError(s3Bucket, err)
			}
		}

		if s3Bucket.Spec.CreateUserFromTemplate != "" {
			r.logger.Debugw("Removing policy", "name", req.Name, "namespace", req.Namespace)
			err := r.Client.Delete(ctx, &s3oditservicesv1alpha1.S3Policy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      s3Bucket.Name,
					Namespace: req.Namespace,
				},
			})
			if err != nil {
				r.logger.Errorw("Failed to remove policy", "name", req.Name, "namespace", req.Namespace, "error", err)
				return r.HandleError(s3Bucket, err)
			}

			r.logger.Debugw("Removing user", "name", req.Name, "namespace", req.Namespace)
			err = r.Client.Delete(ctx, &s3oditservicesv1alpha1.S3User{
				ObjectMeta: metav1.ObjectMeta{
					Name:      s3Bucket.Name,
					Namespace: req.Namespace,
				},
			})
			if err != nil {
				r.logger.Errorw("Failed to remove user", "name", req.Name, "namespace", req.Namespace, "error", err)
				return r.HandleError(s3Bucket, err)
			}
		}

		controllerutil.RemoveFinalizer(s3Bucket, "s3.odit.services/bucket")
		err := r.Update(ctx, s3Bucket)
		if err != nil {
			r.logger.Errorw("Failed to remove finalizer from s3Bucket resource", "name", req.Name, "namespace", req.Namespace, "error", err)
			return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
		}
		r.logger.Infow("Finished reconciling s3Bucket", "name", req.Name, "namespace", req.Namespace, "bucketName", bucketName)
		return ctrl.Result{}, nil
	}

	if !bucketExists {
		s3Bucket.Status.LastAction = s3oditservicesv1alpha1.ActionCreate
		err = s3Client.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{
			Region:        s3Bucket.Spec.Region,
			ObjectLocking: s3Bucket.Spec.ObjectLocking,
		})
		if err != nil {
			r.logger.Errorw("Failed to create bucket", "name", s3Bucket.Name, "bucketName", bucketName, "error", err)
			return r.HandleError(s3Bucket, err)
		}
		s3Bucket.Status.Created = true
	} else {
		s3Bucket.Status.LastAction = s3oditservicesv1alpha1.ActionUpdate
		err = r.Status().Update(ctx, s3Bucket)
		if err != nil {
			r.logger.Errorw("Failed to update s3Bucket status", "name", req.Name, "namespace", req.Namespace, "error", err)
			return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
		}
	}

	if s3Bucket.Spec.CreateUserFromTemplate != "" {
		r.logger.Infow("Create policy for bucket", "name", req.Name, "namespace", req.Namespace, "template", s3Bucket.Spec.CreateUserFromTemplate)

		var policyContent string
		switch s3Bucket.Spec.CreateUserFromTemplate {
		case "readwrite":
			policyContent = PolicyReadWrite
		default:
			return r.HandleError(s3Bucket, fmt.Errorf("unknown template: %s", s3Bucket.Spec.CreateUserFromTemplate))
		}

		policyExists := false
		policy := &s3oditservicesv1alpha1.S3Policy{}
		err = r.Client.Get(ctx, client.ObjectKey{
			Namespace: req.Namespace,
			Name:      s3Bucket.Name,
		}, policy)
		if err == nil {
			r.logger.Debugw("Policy already exists", "name", req.Name, "namespace", req.Namespace)
			policyExists = true
		} else if client.IgnoreNotFound(err) != nil {
			r.logger.Errorw("Failed to get policy", "name", req.Name, "namespace", req.Namespace, "error", err)
			return r.HandleError(s3Bucket, err)
		}

		if !policyExists {
			err := r.Client.Create(ctx, &s3oditservicesv1alpha1.S3Policy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      s3Bucket.Name,
					Namespace: req.Namespace,
					Labels: map[string]string{
						"bucket":   s3Bucket.Name,
						"template": s3Bucket.Spec.CreateUserFromTemplate,
						"owner":    s3Bucket.Name,
					},
				},
				Spec: s3oditservicesv1alpha1.S3PolicySpec{
					ServerRef:     s3Bucket.Spec.ServerRef,
					PolicyContent: policyContent,
				},
			})
			if err != nil {
				r.logger.Errorw("Failed to create policy", "name", req.Name, "namespace", req.Namespace, "error", err)
				return r.HandleError(s3Bucket, err)
			}
			r.logger.Infow("Finished creating policy", "name", req.Name, "namespace", req.Namespace)
		} else {
			policy.Spec.PolicyContent = policyContent
			err := r.Client.Update(ctx, policy)
			if err != nil {
				r.logger.Errorw("Failed to update policy", "name", req.Name, "namespace", req.Namespace, "error", err)
				return r.HandleError(s3Bucket, err)
			}
			r.logger.Infow("Finished updating policy", "name", req.Name, "namespace", req.Namespace)
		}

		userExists := false
		user := &s3oditservicesv1alpha1.S3User{}
		err = r.Client.Get(ctx, client.ObjectKey{
			Namespace: req.Namespace,
			Name:      s3Bucket.Name,
		}, user)
		if err == nil {
			r.logger.Debugw("User already exists", "name", req.Name, "namespace", req.Namespace)
			userExists = true
		} else if client.IgnoreNotFound(err) != nil {
			r.logger.Errorw("Failed to get user", "name", req.Name, "namespace", req.Namespace, "error", err)
			return r.HandleError(s3Bucket, err)
		}

		if !userExists {
			r.logger.Infow("Create user for bucket", "name", req.Name, "namespace", req.Namespace, "template", s3Bucket.Spec.CreateUserFromTemplate)
			err = r.Client.Create(ctx, &s3oditservicesv1alpha1.S3User{
				ObjectMeta: metav1.ObjectMeta{
					Name:      s3Bucket.Name,
					Namespace: req.Namespace,
					Labels: map[string]string{
						"bucket":   s3Bucket.Name,
						"template": s3Bucket.Spec.CreateUserFromTemplate,
						"owner":    s3Bucket.Name,
					},
				},
				Spec: s3oditservicesv1alpha1.S3UserSpec{
					ServerRef:  s3Bucket.Spec.ServerRef,
					PolicyRefs: []string{s3Bucket.Name},
				},
			})
			if err != nil {
				r.logger.Errorw("Failed to create user", "name", req.Name, "namespace", req.Namespace, "error", err)
				return r.HandleError(s3Bucket, err)
			}
			r.logger.Infow("Finished creating user", "name", req.Name, "namespace", req.Namespace)
		} else {
			r.logger.Infow("User already exists", "name", req.Name, "namespace", req.Namespace)
		}
	}

	r.logger.Infow("Finished reconciling s3Bucket", "name", req.Name, "namespace", req.Namespace, "bucketName", bucketName)
	s3Bucket.Status = s3oditservicesv1alpha1.S3BucketStatus{
		CrStatus: s3oditservicesv1alpha1.CrStatus{
			State:             s3oditservicesv1alpha1.StateSuccess,
			LastAction:        s3Bucket.Status.LastAction,
			LastMessage:       "S3Bucket is ready",
			LastReconcileTime: time.Now().Format(time.RFC3339),
			CurrentRetries:    0,
		},
		Name:    s3Bucket.Status.Name,
		Created: true,
	}

	err = r.Status().Update(ctx, s3Bucket)
	if err != nil {
		r.logger.Errorw("Failed to update s3Bucket status", "name", req.Name, "namespace", req.Namespace, "error", err)
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
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
