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
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	s3oditservicesv1alpha1 "github.com/odit-services/s3ops/api/v1alpha1"
	s3client "github.com/odit-services/s3ops/internal/services/s3client"
	corev1 "k8s.io/api/core/v1"
)

// S3ServerReconciler reconciles a S3Server object
type S3ServerReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	logger          *zap.SugaredLogger
	S3ClientFactory s3client.S3ClientFactory
}

func (r *S3ServerReconciler) HandleError(s3Server *s3oditservicesv1alpha1.S3Server, err error) (ctrl.Result, error) {
	r.logger.Errorw("Failed to reconcile s3Server", "name", s3Server.Name, "namespace", s3Server.Namespace, "error", err)
	s3Server.Status = s3oditservicesv1alpha1.S3ServerStatus{
		CrStatus: s3oditservicesv1alpha1.CrStatus{
			State:             s3oditservicesv1alpha1.StateFailed,
			LastAction:        s3Server.Status.LastAction,
			LastMessage:       fmt.Sprintf("Failed to reconcile s3Server: %v", err),
			LastReconcileTime: time.Now().Format(time.RFC3339),
			CurrentRetries:    s3Server.Status.CurrentRetries + 1,
		},
		Online: s3Server.Status.Online,
	}
	updateErr := r.Status().Update(context.Background(), s3Server)
	if updateErr != nil {
		r.logger.Errorw("Failed to update s3Server status", "name", s3Server.Name, "namespace", s3Server.Namespace, "error", updateErr)
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
	}
	r.logger.Infow("Requeue s3Server", "name", s3Server.Name, "namespace", s3Server.Namespace)
	return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
}

// +kubebuilder:rbac:groups=s3.odit.services,resources=s3servers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=s3.odit.services,resources=s3servers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=s3.odit.services,resources=s3servers/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch

func (r *S3ServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger.Infow("Reconciling S3Server", "name", req.Name, "namespace", req.Namespace)

	s3Server := &s3oditservicesv1alpha1.S3Server{}
	err := r.Get(ctx, req.NamespacedName, s3Server)
	if err != nil {
		r.logger.Errorw("Failed to get S3Server resource", "name", req.Name, "namespace", req.Namespace, "error", err)
		return r.HandleError(s3Server, err)
	}

	s3Server.Status = s3oditservicesv1alpha1.S3ServerStatus{
		CrStatus: s3oditservicesv1alpha1.CrStatus{
			State:             s3oditservicesv1alpha1.StateReconciling,
			LastAction:        s3Server.Status.LastAction,
			LastMessage:       "Reconciling S3Server",
			LastReconcileTime: time.Now().Format(time.RFC3339),
			CurrentRetries:    s3Server.Status.CurrentRetries,
		},
		Online: s3Server.Status.Online,
	}

	err = r.Status().Update(ctx, s3Server)
	if err != nil {
		return r.HandleError(s3Server, err)
	}

	if s3Server.Spec.Auth.ExistingSecretRef != "" {
		r.logger.Infow("Using existing secret for S3Server", "name", req.Name, "namespace", req.Namespace)
		secret := &corev1.Secret{}
		err := r.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: s3Server.Spec.Auth.ExistingSecretRef}, secret)
		if err != nil {
			r.logger.Errorw("Failed to get secret for S3Server", "name", req.Name, "namespace", req.Namespace, "error", err)
			return r.HandleError(s3Server, err)
		}

		if s3Server.Spec.Auth.AccessKeySecretKey == "" {
			s3Server.Spec.Auth.AccessKeySecretKey = "accessKey"
		}
		if s3Server.Spec.Auth.SecretKeySecretKey == "" {
			s3Server.Spec.Auth.SecretKeySecretKey = "secretKey"
		}
		s3Server.Spec.Auth.AccessKey = string(secret.Data[s3Server.Spec.Auth.AccessKeySecretKey])
		s3Server.Spec.Auth.SecretKey = string(secret.Data[s3Server.Spec.Auth.SecretKeySecretKey])
	}

	minioClient, err := r.S3ClientFactory.NewClient(*s3Server)
	if err != nil {
		r.logger.Errorw("Failed to create Minio client for S3Server", "name", req.Name, "namespace", req.Namespace, "error", err)
		return r.HandleError(s3Server, err)
	}

	minioClient.HealthCheck(1 * time.Second)
	time.Sleep(1 * time.Second)

	minioOnline := minioClient.IsOnline()
	if !minioOnline {
		r.logger.Errorw("Minio server is offline for S3Server", "name", req.Name, "namespace", req.Namespace)
		s3Server.Status.Online = false
		return r.HandleError(s3Server, fmt.Errorf("minio server is offline"))
	}

	r.logger.Infow("Finished reconciling S3Server", "name", req.Name, "namespace", req.Namespace)
	s3Server.Status = s3oditservicesv1alpha1.S3ServerStatus{
		CrStatus: s3oditservicesv1alpha1.CrStatus{
			State:             s3oditservicesv1alpha1.StateSuccess,
			LastAction:        s3Server.Status.LastAction,
			LastMessage:       "S3Server reconciled",
			LastReconcileTime: time.Now().Format(time.RFC3339),
			CurrentRetries:    0,
		},
		Online: true,
	}
	err = r.Status().Update(ctx, s3Server)
	if err != nil {
		r.logger.Errorw("Failed to update S3Server status", "name", req.Name, "namespace", req.Namespace, "error", err)
		return r.HandleError(s3Server, err)
	}

	return ctrl.Result{
		RequeueAfter: 5 * time.Minute,
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *S3ServerReconciler) SetupWithManager(mgr ctrl.Manager) error {

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
		For(&s3oditservicesv1alpha1.S3Server{}).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{})).
		Complete(r)
}
