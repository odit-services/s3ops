package webhook

import (
	"context"

	s3oditservicesv1alpha1 "github.com/odit-services/s3ops/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// S3User implements webhook for S3User
type S3User struct{}

// +kubebuilder:webhook:path=/validate-s3-odit-services-v1alpha1-s3user,mutating=false,failurePolicy=fail,sideEffects=None,groups=s3.odit.services,resources=s3users,verbs=create;update,versions=v1alpha1,name=vs3user.s3.odit.services,admissionReviewVersions=v1

func (r *S3User) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &s3oditservicesv1alpha1.S3User{}).
		WithValidator(&s3UserWebhookAdapter{}).
		Complete()
}

type s3UserWebhookAdapter struct{}

func (a *s3UserWebhookAdapter) ValidateCreate(ctx context.Context, obj *s3oditservicesv1alpha1.S3User) (admission.Warnings, error) {
	return nil, validateServerRefCrossNamespace(obj.Spec.ServerRef, obj.Namespace)
}

func (a *s3UserWebhookAdapter) ValidateUpdate(ctx context.Context, oldObj, newObj *s3oditservicesv1alpha1.S3User) (admission.Warnings, error) {
	return nil, validateServerRefCrossNamespace(newObj.Spec.ServerRef, newObj.Namespace)
}

func (a *s3UserWebhookAdapter) ValidateDelete(ctx context.Context, obj *s3oditservicesv1alpha1.S3User) (admission.Warnings, error) {
	return nil, nil
}
