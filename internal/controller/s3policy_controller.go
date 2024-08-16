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
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	s3oditservicesv1alpha1 "github.com/odit-services/s3ops/api/v1alpha1"
	s3client "github.com/odit-services/s3ops/internal/services/s3client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// S3PolicyReconciler reconciles a S3Policy object
type S3PolicyReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	logger          *zap.SugaredLogger
	S3ClientFactory s3client.S3ClientFactory
}

//+kubebuilder:rbac:groups=s3.odit.services,resources=s3policies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=s3.odit.services,resources=s3policies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=s3.odit.services,resources=s3policies/finalizers,verbs=update

func (r *S3PolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	r.logger.Infow("Reconciling S3Policy", "name", req.Name, "namespace", req.Namespace)

	s3Policy := &s3oditservicesv1alpha1.S3Policy{}
	err := r.Get(ctx, req.NamespacedName, s3Policy)
	if err != nil {
		r.logger.Errorw("Failed to get S3Policy resource", "name", req.Name, "namespace", req.Namespace, "error", err)
		s3Policy.Status.Conditions = append(s3Policy.Status.Conditions, metav1.Condition{
			Type:               s3oditservicesv1alpha1.ConditionFailed,
			Status:             metav1.ConditionFalse,
			Reason:             s3oditservicesv1alpha1.ReasonNotFound,
			Message:            err.Error(),
			LastTransitionTime: metav1.Now(),
		})

		return ctrl.Result{}, err
	}

	s3Policy.Status.Conditions = append(s3Policy.Status.Conditions, metav1.Condition{
		Type:               s3oditservicesv1alpha1.ConditionReconciling,
		Status:             metav1.ConditionTrue,
		Reason:             s3oditservicesv1alpha1.ConditionReconciling,
		Message:            "S3Policy reconciling",
		LastTransitionTime: metav1.Now(),
	})
	err = r.Status().Update(ctx, s3Policy)
	if err != nil {
		r.logger.Errorw("Failed to update S3Policy status", "name", req.Name, "namespace", req.Namespace, "error", err)
		return ctrl.Result{}, err
	}

	if !controllerutil.ContainsFinalizer(s3Policy, "s3.odit.services/policy") {
		controllerutil.AddFinalizer(s3Policy, "s3.odit.services/policy")
		err = r.Update(ctx, s3Policy)
		if err != nil {
			r.logger.Errorw("Failed to add finalizer to S3Policy", "name", req.Name, "namespace", req.Namespace, "error", err)
			s3Policy.Status.Conditions = append(s3Policy.Status.Conditions, metav1.Condition{
				Type:               s3oditservicesv1alpha1.ConditionFailed,
				Status:             metav1.ConditionFalse,
				Reason:             s3oditservicesv1alpha1.ReasonFinalizerFailed,
				Message:            "Failed to add finalizer to S3Policy",
				LastTransitionTime: metav1.Now(),
			})
			r.Status().Update(ctx, s3Policy)
			return ctrl.Result{}, err
		}
	}

	s3AdminClient, condition, err := s3client.GetS3AdminClientFromS3Server(s3Policy.Spec.ServerRef, r.S3ClientFactory, r.Client)
	if err != nil {
		r.logger.Errorw("Failed to get S3AdminClient", "name", req.Name, "namespace", req.Namespace, "error", err)
		s3Policy.Status.Conditions = append(s3Policy.Status.Conditions, condition)
		r.Status().Update(ctx, s3Policy)
		return ctrl.Result{}, err
	}

	//TODO: Handle policy creation

	r.logger.Infow("S3Policy reconciled", "name", req.Name, "namespace", req.Namespace)
	s3Policy.Status.Conditions = append(s3Policy.Status.Conditions, metav1.Condition{
		Type:               s3oditservicesv1alpha1.ConditionReady,
		Status:             metav1.ConditionTrue,
		Reason:             s3oditservicesv1alpha1.ConditionReady,
		Message:            "S3Policy reconciled",
		LastTransitionTime: metav1.Now(),
	})
	err = r.Status().Update(ctx, s3Policy)
	if err != nil {
		r.logger.Errorw("Failed to update S3Policy status", "name", req.Name, "namespace", req.Namespace, "error", err)
		return ctrl.Result{}, err
	}

	return ctrl.Result{
		RequeueAfter: 5 * time.Minute,
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *S3PolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
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
		For(&s3oditservicesv1alpha1.S3Policy{}).
		Complete(r)
}
