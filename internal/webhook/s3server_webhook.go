package webhook

import (
	"context"
	"fmt"

	s3oditservicesv1alpha1 "github.com/odit-services/s3ops/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// S3Server implements webhook for S3Server
type S3Server struct{}

// +kubebuilder:webhook:path=/validate-s3-odit-services-v1alpha1-s3server,mutating=false,failurePolicy=fail,sideEffects=None,groups=s3.odit.services,resources=s3servers,verbs=create;update,versions=v1alpha1,name=vs3server.s3.odit.services,admissionReviewVersions=v1

func (r *S3Server) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &s3oditservicesv1alpha1.S3Server{}).
		WithValidator(&s3ServerWebhookAdapter{}).
		Complete()
}

type s3ServerWebhookAdapter struct{}

func (a *s3ServerWebhookAdapter) ValidateCreate(ctx context.Context, obj *s3oditservicesv1alpha1.S3Server) (admission.Warnings, error) {
	return nil, validateAuth(&obj.Spec.Auth)
}

func (a *s3ServerWebhookAdapter) ValidateUpdate(ctx context.Context, oldObj, newObj *s3oditservicesv1alpha1.S3Server) (admission.Warnings, error) {
	return nil, validateAuth(&newObj.Spec.Auth)
}

func (a *s3ServerWebhookAdapter) ValidateDelete(ctx context.Context, obj *s3oditservicesv1alpha1.S3Server) (admission.Warnings, error) {
	return nil, nil
}

func validateAuth(auth *s3oditservicesv1alpha1.S3ServerAuthSpec) error {
	hasAccessKey := auth.AccessKey != "" && auth.SecretKey != ""
	hasExistingSecret := auth.ExistingSecretRef != ""
	hasApiToken := auth.ApiToken != ""

	if !hasAccessKey && !hasExistingSecret && !hasApiToken {
		return fmt.Errorf("at least one authentication method must be configured: accessKey+secretKey, existingSecretRef, or apiToken")
	}
	return nil
}