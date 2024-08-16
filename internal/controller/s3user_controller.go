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
	gopassword "github.com/sethvargo/go-password/password"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// S3UserReconciler reconciles a S3User object
type S3UserReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	logger          *zap.SugaredLogger
	S3ClientFactory s3client.S3ClientFactory
}

//+kubebuilder:rbac:groups=s3.odit.services,resources=s3users,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=s3.odit.services,resources=s3users/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=s3.odit.services,resources=s3users/finalizers,verbs=update

func (r *S3UserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	r.logger.Infow("Reconciling S3User", "name", req.Name, "namespace", req.Namespace)

	s3User := &s3oditservicesv1alpha1.S3User{}
	err := r.Get(ctx, req.NamespacedName, s3User)
	if err != nil {
		r.logger.Errorw("Failed to get S3User resource", "name", req.Name, "namespace", req.Namespace, "error", err)
		s3User.Status.Conditions = append(s3User.Status.Conditions, metav1.Condition{
			Type:               s3oditservicesv1alpha1.ConditionFailed,
			Status:             metav1.ConditionFalse,
			Reason:             s3oditservicesv1alpha1.ReasonNotFound,
			Message:            "S3User resource not found",
			LastTransitionTime: metav1.Now(),
		})
		r.Status().Update(ctx, s3User)
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
	}

	s3User.Status.Conditions = append(s3User.Status.Conditions, metav1.Condition{
		Type:               s3oditservicesv1alpha1.ConditionReconciling,
		Status:             metav1.ConditionUnknown,
		Reason:             s3oditservicesv1alpha1.ConditionReconciling,
		Message:            "Reconciling S3User resource",
		LastTransitionTime: metav1.Now(),
	})
	err = r.Status().Update(ctx, s3User)
	if err != nil {
		r.logger.Errorw("Failed to update S3User resource status", "name", req.Name, "namespace", req.Namespace, "error", err)
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
	}

	if !controllerutil.ContainsFinalizer(s3User, "s3.odit.services/user") {
		controllerutil.AddFinalizer(s3User, "s3.odit.services/user")
		err := r.Update(ctx, s3User)
		if err != nil {
			r.logger.Errorw("Failed to add finalizer to s3User resource", "name", req.Name, "namespace", req.Namespace, "error", err)
			s3User.Status.Conditions = append(s3User.Status.Conditions, metav1.Condition{
				Type:               s3oditservicesv1alpha1.ConditionFailed,
				Status:             metav1.ConditionFalse,
				Reason:             s3oditservicesv1alpha1.ReasonFinalizerFailedToApply,
				Message:            "Failed to add finalizer to s3User resource",
				LastTransitionTime: metav1.Now(),
			})
			r.Status().Update(ctx, s3User)
			return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
		}
	}

	s3AdminClient, condition, err := s3client.GetS3AdminClientFromS3Server(s3User.Spec.ServerRef, r.S3ClientFactory, r.Client)
	if err != nil {
		r.logger.Errorw("Failed to create S3 admin client for S3User", "name", req.Name, "namespace", req.Namespace, "error", err)
		s3User.Status.Conditions = append(s3User.Status.Conditions, condition)
		r.Status().Update(ctx, s3User)
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
	}

	var secret *corev1.Secret
	if s3User.Status.SecretRef == "" {
		nanoID, err := gonanoid.Generate("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890", 20)
		if err != nil {
			r.logger.Errorw("Failed to generate nanoID", "error", err)
			s3User.Status.Conditions = append(s3User.Status.Conditions, metav1.Condition{
				Type:               s3oditservicesv1alpha1.ConditionFailed,
				Status:             metav1.ConditionFalse,
				Reason:             s3oditservicesv1alpha1.ReasonCreateFailed,
				Message:            "Failed to generate nanoID",
				LastTransitionTime: metav1.Now(),
			})
			r.Status().Update(ctx, s3User)
			return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
		}

		secretKey, err := gopassword.Generate(64, 20, 0, true, true)
		if err != nil {
			r.logger.Errorw("Failed to generate secretKey", "error", err)
			s3User.Status.Conditions = append(s3User.Status.Conditions, metav1.Condition{
				Type:               s3oditservicesv1alpha1.ConditionFailed,
				Status:             metav1.ConditionFalse,
				Reason:             s3oditservicesv1alpha1.ReasonCreateFailed,
				Message:            "Failed to generate secretKey",
				LastTransitionTime: metav1.Now(),
			})
			r.Status().Update(ctx, s3User)
			return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
		}

		secret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-s3creds", s3User.Name),
				Namespace: s3User.Namespace,
			},
			StringData: map[string]string{
				"accessKey": nanoID,
				"secretKey": secretKey,
			},
		}
		condition, err = createSecret(ctx, r.Client, secret)
		if err != nil {
			r.logger.Errorw("Failed to create secret for S3User", "name", req.Name, "namespace", req.Namespace, "error", err)
			s3User.Status.Conditions = append(s3User.Status.Conditions, condition)
			r.Status().Update(ctx, s3User)
			return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
		}
		s3User.Status.SecretRef = secret.Name
		r.Status().Update(ctx, s3User)
	} else {
		secret, condition, err = getSecret(ctx, r.Client, s3User.Namespace, s3User.Status.SecretRef)
		if err != nil {
			r.logger.Errorw("Failed to get secret for S3User", "name", req.Name, "namespace", req.Namespace, "error", err)
			s3User.Status.Conditions = append(s3User.Status.Conditions, condition)
			r.Status().Update(ctx, s3User)
			return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
		}
	}

	userCreds := struct {
		AccessKey string
		SecretKey string
	}{
		AccessKey: string(secret.Data["accessKey"]),
		SecretKey: string(secret.Data["secretKey"]),
	}

	userExists, err := s3AdminClient.UserExists(ctx, userCreds.AccessKey)
	if err != nil {
		r.logger.Errorw("Failed to check if user exists", "name", req.Name, "namespace", req.Namespace, "error", err)
		s3User.Status.Conditions = append(s3User.Status.Conditions, metav1.Condition{
			Type:               s3oditservicesv1alpha1.ConditionFailed,
			Status:             metav1.ConditionFalse,
			Reason:             s3oditservicesv1alpha1.ReasonRequestFailed,
			Message:            fmt.Sprintf("Failed to check if user exists: %v", err),
			LastTransitionTime: metav1.Now(),
		})
		r.Status().Update(ctx, s3User)
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
	}

	if s3User.DeletionTimestamp != nil {
		r.logger.Infow("Deleting S3User", "name", req.Name, "namespace", req.Namespace)
		if !userExists {
			r.logger.Debugw("User does not exist", "name", req.Name, "namespace", req.Namespace)
		} else {
			err := s3AdminClient.RemoveUser(ctx, userCreds.AccessKey)
			if err != nil {
				r.logger.Errorw("Failed to remove user", "name", req.Name, "namespace", req.Namespace, "error", err)
				s3User.Status.Conditions = append(s3User.Status.Conditions, metav1.Condition{
					Type:               s3oditservicesv1alpha1.ConditionFailed,
					Status:             metav1.ConditionFalse,
					Reason:             s3oditservicesv1alpha1.ReasonRequestFailed,
					Message:            fmt.Sprintf("Failed to remove user: %v", err),
					LastTransitionTime: metav1.Now(),
				})
				r.Status().Update(ctx, s3User)
				return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
			}
		}
		controllerutil.RemoveFinalizer(s3User, "s3.odit.services/user")
		err := r.Update(ctx, s3User)
		if err != nil {
			r.logger.Errorw("Failed to remove finalizer from S3User resource", "name", req.Name, "namespace", req.Namespace, "error", err)
			return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
		}
		return ctrl.Result{}, nil
	}

	if !userExists {
		r.logger.Infow("Creating user", "name", req.Name, "namespace", req.Namespace)
		err = s3AdminClient.MakeUser(ctx, userCreds.AccessKey, userCreds.SecretKey)
		if err != nil {
			r.logger.Errorw("Failed to create user", "name", req.Name, "namespace", req.Namespace, "error", err)
			s3User.Status.Conditions = append(s3User.Status.Conditions, metav1.Condition{
				Type:               s3oditservicesv1alpha1.ConditionFailed,
				Status:             metav1.ConditionFalse,
				Reason:             s3oditservicesv1alpha1.ReasonRequestFailed,
				Message:            fmt.Sprintf("Failed to create user: %v", err),
				LastTransitionTime: metav1.Now(),
			})
			r.Status().Update(ctx, s3User)
			return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
		}
		s3User.Status.Created = true
	}

	for _, policyRef := range s3User.Spec.PolicyRefs {
		r.logger.Debugw("Applying policy to user", "name", req.Name, "namespace", req.Namespace, "policy", policyRef)
		policyExists, err := s3AdminClient.PolicyExists(ctx, policyRef)
		if err != nil {
			r.logger.Errorw("Failed to check if policy exists", "name", req.Name, "namespace", req.Namespace, "error", err)
			s3User.Status.Conditions = append(s3User.Status.Conditions, metav1.Condition{
				Type:               s3oditservicesv1alpha1.ConditionFailed,
				Status:             metav1.ConditionFalse,
				Reason:             s3oditservicesv1alpha1.ReasonRequestFailed,
				Message:            fmt.Sprintf("Failed to check if policy exists: %v", err),
				LastTransitionTime: metav1.Now(),
			})
			r.Status().Update(ctx, s3User)
			return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
		}

		if !policyExists {
			r.logger.Errorw("Policy does not exist", "name", req.Name, "namespace", req.Namespace, "policy", policyRef)
			s3User.Status.Conditions = append(s3User.Status.Conditions, metav1.Condition{
				Type:               s3oditservicesv1alpha1.ConditionFailed,
				Status:             metav1.ConditionFalse,
				Reason:             s3oditservicesv1alpha1.ReasonRequestFailed,
				Message:            fmt.Sprintf("Policy %s does not exist", policyRef),
				LastTransitionTime: metav1.Now(),
			})
			r.Status().Update(ctx, s3User)
			return ctrl.Result{}, fmt.Errorf("policy %s does not exist", policyRef)
		}

		err = s3AdminClient.ApplyPolicyToUser(ctx, policyRef, userCreds.AccessKey)
		if err != nil {
			r.logger.Errorw("Failed to apply policy to user", "name", req.Name, "namespace", req.Namespace, "error", err)
			s3User.Status.Conditions = append(s3User.Status.Conditions, metav1.Condition{
				Type:               s3oditservicesv1alpha1.ConditionFailed,
				Status:             metav1.ConditionFalse,
				Reason:             s3oditservicesv1alpha1.ReasonRequestFailed,
				Message:            fmt.Sprintf("Failed to apply policy to user: %v", err),
				LastTransitionTime: metav1.Now(),
			})
			r.Status().Update(ctx, s3User)
			return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
		}
		r.logger.Debugw("Policy applied to user", "name", req.Name, "namespace", req.Namespace, "policy", policyRef)
	}

	r.logger.Infow("Finished reconciling S3User", "name", req.Name, "namespace", req.Namespace)
	s3User.Status.Conditions = append(s3User.Status.Conditions, metav1.Condition{
		Type:               s3oditservicesv1alpha1.ConditionReady,
		Status:             metav1.ConditionTrue,
		Reason:             metav1.StatusSuccess,
		Message:            "S3User reconciled",
		LastTransitionTime: metav1.Now(),
	})
	err = r.Status().Update(ctx, s3User)
	if err != nil {
		r.logger.Errorw("Failed to update S3User resource status", "name", req.Name, "namespace", req.Namespace, "error", err)
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
	}

	return ctrl.Result{
		RequeueAfter: 5 * time.Minute,
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *S3UserReconciler) SetupWithManager(mgr ctrl.Manager) error {
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
		For(&s3oditservicesv1alpha1.S3User{}).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{})).
		Complete(r)
}
