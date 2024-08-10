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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	s3oditservicesv1alpha1 "github.com/odit-services/s3ops/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// S3ServerReconciler reconciles a S3Server object
type S3ServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger *zap.SugaredLogger
}

// +kubebuilder:rbac:groups=s3.odit.services,resources=s3servers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=s3.odit.services,resources=s3servers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=s3.odit.services,resources=s3servers/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

func (r *S3ServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	r.logger.Infow("Reconciling S3Server", "name", req.Name, "namespace", req.Namespace)

	s3Server := &s3oditservicesv1alpha1.S3Server{}
	err := r.Get(ctx, req.NamespacedName, s3Server)
	if err != nil {
		r.logger.Errorw("Failed to get S3Server resource", "name", req.Name, "namespace", req.Namespace, "error", err)
		s3Server.Status.Conditions = append(s3Server.Status.Conditions, metav1.Condition{
			Type:    s3oditservicesv1alpha1.ConditionFailed,
			Status:  metav1.ConditionFalse,
			Reason:  s3oditservicesv1alpha1.ReasonNotFound,
			Message: "S3Server resource not found",
		})
		r.Status().Update(ctx, s3Server)
		return ctrl.Result{}, err
	}

	s3Server.Status.Conditions = append(s3Server.Status.Conditions, metav1.Condition{
		Type:    s3oditservicesv1alpha1.ConditionReconciling,
		Status:  metav1.ConditionUnknown,
		Message: fmt.Sprintf("Reconciling S3Server %s", s3Server.Name),
	})
	err = r.Status().Update(ctx, s3Server)
	if err != nil {
		return ctrl.Result{}, err
	}

	minioClient, err := minio.New(s3Server.Spec.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s3Server.Spec.Auth.AccessKey, s3Server.Spec.Auth.SecretKey, ""),
		Secure: s3Server.Spec.TLS,
	})
	if err != nil {
		r.logger.Errorw("Failed to create Minio client for S3Server", "name", req.Name, "namespace", req.Namespace, "error", err)
		s3Server.Status.Conditions = append(s3Server.Status.Conditions, metav1.Condition{
			Type:    s3oditservicesv1alpha1.ConditionFailed,
			Status:  metav1.ConditionFalse,
			Message: fmt.Sprintf("Failed to create Minio client: %v", err),
		})
		r.Status().Update(ctx, s3Server)
		return ctrl.Result{}, err
	}

	minioOnline := minioClient.IsOnline()
	if !minioOnline {
		r.logger.Errorw("Minio server is offline for S3Server", "name", req.Name, "namespace", req.Namespace)
		s3Server.Status.Conditions = append(s3Server.Status.Conditions, metav1.Condition{
			Type:    s3oditservicesv1alpha1.ConditionFailed,
			Status:  metav1.ConditionFalse,
			Reason:  s3oditservicesv1alpha1.ReasonOffline,
			Message: "Minio server is offline",
		})
		r.Status().Update(ctx, s3Server)
		return ctrl.Result{}, fmt.Errorf("minio server is offline")
	}

	r.logger.Infow("Finished reconciling S3Server", "name", req.Name, "namespace", req.Namespace)
	s3Server.Status.Conditions = append(s3Server.Status.Conditions, metav1.Condition{
		Type:    s3oditservicesv1alpha1.ConditionReady,
		Status:  metav1.ConditionTrue,
		Message: "S3Server is ready",
	})
	err = r.Status().Update(ctx, s3Server)
	if err != nil {
		r.logger.Errorw("Failed to update S3Server status", "name", req.Name, "namespace", req.Namespace, "error", err)
		return ctrl.Result{}, err
	}

	return ctrl.Result{
		RequeueAfter: 5 * time.Minute,
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *S3ServerReconciler) SetupWithManager(mgr ctrl.Manager) error {

	lovLevel := os.Getenv("LOG_LEVEL")
	if lovLevel == "" {
		lovLevel = "INFO"
	}

	var zapLogLevel zapcore.Level
	err := zapLogLevel.UnmarshalText([]byte(strings.ToLower(lovLevel)))
	if err != nil {
		zapLogLevel = zapcore.InfoLevel
	}

	zapConfig := zap.NewProductionConfig()
	zapConfig.Level = zap.NewAtomicLevelAt(zapLogLevel)
	zapLogger, _ := zapConfig.Build()
	defer zapLogger.Sync()
	r.logger = zapLogger.Sugar()

	return ctrl.NewControllerManagedBy(mgr).
		For(&s3oditservicesv1alpha1.S3Server{}).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{})).
		Complete(r)
}
