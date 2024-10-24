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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	gonanoid "github.com/matoous/go-nanoid/v2"
	s3oditservicesv1alpha1 "github.com/odit-services/s3ops/api/v1alpha1"
	s3client "github.com/odit-services/s3ops/internal/services/s3client"
)

// S3PolicyReconciler reconciles a S3Policy object
type S3PolicyReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	logger          *zap.SugaredLogger
	S3ClientFactory s3client.S3ClientFactory
}

func (r *S3PolicyReconciler) HandleError(s3Policy *s3oditservicesv1alpha1.S3Policy, err error) (ctrl.Result, error) {
	r.logger.Errorw("Failed to reconcile S3Policy", "name", s3Policy.Name, "namespace", s3Policy.Namespace, "error", err)
	s3Policy.Status = s3oditservicesv1alpha1.S3PolicyStatus{
		CrStatus: s3oditservicesv1alpha1.CrStatus{
			State:             s3oditservicesv1alpha1.StateFailed,
			LastAction:        s3Policy.Status.LastAction,
			LastMessage:       fmt.Sprintf("Failed to reconcile S3Policy: %v", err),
			LastReconcileTime: time.Now().Format(time.RFC3339),
			CurrentRetries:    s3Policy.Status.CurrentRetries + 1,
		},
		Created: s3Policy.Status.Created,
		Name:    s3Policy.Status.Name,
	}
	updateErr := r.Status().Update(context.Background(), s3Policy)
	if updateErr != nil {
		r.logger.Errorw("Failed to update S3Policy status", "name", s3Policy.Name, "namespace", s3Policy.Namespace, "error", updateErr)
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
	}
	r.logger.Infow("Requeue S3Policy", "name", s3Policy.Name, "namespace", s3Policy.Namespace)
	return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
}

// +kubebuilder:rbac:groups=s3.odit.services,resources=s3policies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=s3.odit.services,resources=s3policies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=s3.odit.services,resources=s3policies/finalizers,verbs=update

func (r *S3PolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	r.logger.Infow("Reconciling S3Policy", "name", req.Name, "namespace", req.Namespace)

	s3Policy := &s3oditservicesv1alpha1.S3Policy{}
	err := r.Get(ctx, req.NamespacedName, s3Policy)
	if err != nil {
		r.logger.Errorw("Failed to get S3Policy resource", "name", req.Name, "namespace", req.Namespace, "error", err)
		return r.HandleError(s3Policy, err)
	}

	s3Policy.Status = s3oditservicesv1alpha1.S3PolicyStatus{
		CrStatus: s3oditservicesv1alpha1.CrStatus{
			State:             s3oditservicesv1alpha1.StateReconciling,
			LastAction:        s3oditservicesv1alpha1.ActionUnknown,
			LastMessage:       fmt.Sprintf("Reconciling S3Policy %s", s3Policy.Name),
			LastReconcileTime: time.Now().Format(time.RFC3339),
			CurrentRetries:    s3Policy.Status.CurrentRetries,
		},
		Created: s3Policy.Status.Created,
		Name:    s3Policy.Status.Name,
	}
	err = r.Status().Update(ctx, s3Policy)
	if err != nil {
		r.logger.Errorw("Failed to update S3Policy status", "name", req.Name, "namespace", req.Namespace, "error", err)
		return r.HandleError(s3Policy, err)
	}

	if !controllerutil.ContainsFinalizer(s3Policy, "s3.odit.services/policy") {
		controllerutil.AddFinalizer(s3Policy, "s3.odit.services/policy")
		err = r.Update(ctx, s3Policy)
		if err != nil {
			r.logger.Errorw("Failed to add finalizer to S3Policy", "name", req.Name, "namespace", req.Namespace, "error", err)
			return r.HandleError(s3Policy, err)
		}
	}

	s3AdminClient, err := s3client.GetS3AdminClientFromS3Server(s3Policy.Spec.ServerRef, r.S3ClientFactory, r.Client)
	if err != nil {
		r.logger.Errorw("Failed to get S3AdminClient", "name", req.Name, "namespace", req.Namespace, "error", err)
		return r.HandleError(s3Policy, err)
	}

	var policyName string
	if s3Policy.Status.Name != "" {
		policyName = s3Policy.Status.Name
	} else {
		r.logger.Debugw("Generating policy name", "name", s3Policy.Name, "namespace", s3Policy.Namespace)
		nanoID, err := gonanoid.Generate("abcdefghijklmnopqrstuvwxyz0123456789", 21)
		if err != nil {
			r.logger.Errorw("Failed to generate bucket name", "name", s3Policy.Name, "error", err)
			return r.HandleError(s3Policy, err)
		}
		policyPrefix := fmt.Sprintf("%s-%s", s3Policy.Name, s3Policy.Namespace)
		var truncateAt int
		if len(policyPrefix) > 39 {
			truncateAt = 39
		} else {
			truncateAt = len(policyPrefix)
		}
		policyName = fmt.Sprintf("%s-%s", policyPrefix[0:truncateAt], nanoID)
	}
	s3Policy.Status.Name = policyName

	err = r.Status().Update(ctx, s3Policy)
	if err != nil {
		r.logger.Errorw("Failed to update S3Policy status", "name", req.Name, "namespace", req.Namespace, "error", err)
		return r.HandleError(s3Policy, err)
	}

	policyExists, err := s3AdminClient.PolicyExists(ctx, policyName)
	if err != nil {
		r.logger.Errorw("Failed to check if policy exists", "name", req.Name, "namespace", req.Namespace, "error", err)
		return r.HandleError(s3Policy, err)
	}

	if s3Policy.DeletionTimestamp != nil {
		r.logger.Infow("Deleting S3Policy", "name", req.Name, "namespace", req.Namespace)
		s3Policy.Status.LastAction = s3oditservicesv1alpha1.ActionDelete
		err = r.Status().Update(ctx, s3Policy)
		if err != nil {
			r.logger.Errorw("Failed to update S3Policy status", "name", req.Name, "namespace", req.Namespace, "error", err)
			return r.HandleError(s3Policy, err)
		}

		if !policyExists {
			r.logger.Debugw("User does not exist", "name", req.Name, "namespace", req.Namespace)
		} else {
			err := s3AdminClient.RemovePolicy(ctx, policyName)
			if err != nil {
				r.logger.Errorw("Failed to remove user", "name", req.Name, "namespace", req.Namespace, "error", err)
				return r.HandleError(s3Policy, err)
			}
		}
		controllerutil.RemoveFinalizer(s3Policy, "s3.odit.services/policy")
		err := r.Update(ctx, s3Policy)
		if err != nil {
			r.logger.Errorw("Failed to remove finalizer from S3Policy resource", "name", req.Name, "namespace", req.Namespace, "error", err)
			return r.HandleError(s3Policy, err)
		}
		return ctrl.Result{}, nil
	}

	if !policyExists {
		s3Policy.Status.LastAction = s3oditservicesv1alpha1.ActionCreate
		err = r.Status().Update(ctx, s3Policy)
		if err != nil {
			r.logger.Errorw("Failed to update S3Policy status", "name", req.Name, "namespace", req.Namespace, "error", err)
			return r.HandleError(s3Policy, err)
		}

		err = s3AdminClient.MakePolicy(ctx, policyName, s3Policy.Spec.PolicyContent)
		if err != nil {
			r.logger.Errorw("Failed to create policy", "name", req.Name, "namespace", req.Namespace, "error", err)
			return r.HandleError(s3Policy, err)
		}
	} else {
		s3Policy.Status.LastAction = s3oditservicesv1alpha1.ActionUpdate
		err = r.Status().Update(ctx, s3Policy)
		if err != nil {
			r.logger.Errorw("Failed to update S3Policy status", "name", req.Name, "namespace", req.Namespace, "error", err)
			return r.HandleError(s3Policy, err)
		}

		err = s3AdminClient.UpdatePolicy(ctx, policyName, s3Policy.Spec.PolicyContent)
		if err != nil {
			r.logger.Errorw("Failed to update policy", "name", req.Name, "namespace", req.Namespace, "error", err)
			return r.HandleError(s3Policy, err)
		}
	}

	r.logger.Infow("S3Policy reconciled", "name", req.Name, "namespace", req.Namespace)
	s3Policy.Status = s3oditservicesv1alpha1.S3PolicyStatus{
		CrStatus: s3oditservicesv1alpha1.CrStatus{
			State:             s3oditservicesv1alpha1.StateSuccess,
			LastAction:        s3Policy.Status.LastAction,
			LastMessage:       "S3Policy reconciled",
			LastReconcileTime: time.Now().Format(time.RFC3339),
		},
		Created: true,
		Name:    s3Policy.Status.Name,
	}
	err = r.Status().Update(ctx, s3Policy)
	if err != nil {
		r.logger.Errorw("Failed to update S3Policy status", "name", req.Name, "namespace", req.Namespace, "error", err)
		return r.HandleError(s3Policy, err)
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
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{})).
		Complete(r)
}
