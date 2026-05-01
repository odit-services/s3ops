package webhook

import (
	"context"

	s3oditservicesv1alpha1 "github.com/odit-services/s3ops/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

//nolint:dupl // Each webhook implements For(r) for different CR types

// S3Policy implements webhook for S3Policy
type S3Policy struct{}

// +kubebuilder:webhook:path=/validate-s3-odit-services-v1alpha1-s3policy,mutating=false,failurePolicy=fail,sideEffects=None,groups=s3.odit.services,resources=s3policies,verbs=create;update,versions=v1alpha1,name=vs3policy.s3.odit.services,admissionReviewVersions=v1

func (r *S3Policy) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &s3oditservicesv1alpha1.S3Policy{}).
		WithValidator(&s3PolicyWebhookAdapter{}).
		Complete()
}

type s3PolicyWebhookAdapter struct{}

func (a *s3PolicyWebhookAdapter) ValidateCreate(ctx context.Context, obj *s3oditservicesv1alpha1.S3Policy) (admission.Warnings, error) {
	return nil, validateServerRefCrossNamespace(obj.Spec.ServerRef, obj.Namespace)
}

func (a *s3PolicyWebhookAdapter) ValidateUpdate(ctx context.Context, oldObj, newObj *s3oditservicesv1alpha1.S3Policy) (admission.Warnings, error) {
	return nil, validateServerRefCrossNamespace(newObj.Spec.ServerRef, newObj.Namespace)
}

func (a *s3PolicyWebhookAdapter) ValidateDelete(ctx context.Context, obj *s3oditservicesv1alpha1.S3Policy) (admission.Warnings, error) {
	return nil, nil
}
