package webhook

import (
	"context"

	s3oditservicesv1alpha1 "github.com/odit-services/s3ops/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// S3Bucket implements webhook for S3Bucket
type S3Bucket struct{}

// +kubebuilder:webhook:path=/validate-s3-odit-services-v1alpha1-s3bucket,mutating=false,failurePolicy=fail,sideEffects=None,groups=s3.odit.services,resources=s3buckets,verbs=create;update,versions=v1alpha1,name=vs3bucket.s3.odit.services,admissionReviewVersions=v1

func (r *S3Bucket) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &s3oditservicesv1alpha1.S3Bucket{}).
		WithValidator(&s3BucketWebhookAdapter{}).
		Complete()
}

type s3BucketWebhookAdapter struct{}

func (a *s3BucketWebhookAdapter) ValidateCreate(ctx context.Context, obj *s3oditservicesv1alpha1.S3Bucket) (admission.Warnings, error) {
	return nil, validateServerRefCrossNamespace(obj.Spec.ServerRef, obj.Namespace)
}

func (a *s3BucketWebhookAdapter) ValidateUpdate(ctx context.Context, oldObj, newObj *s3oditservicesv1alpha1.S3Bucket) (admission.Warnings, error) {
	return nil, validateServerRefCrossNamespace(newObj.Spec.ServerRef, newObj.Namespace)
}

func (a *s3BucketWebhookAdapter) ValidateDelete(ctx context.Context, obj *s3oditservicesv1alpha1.S3Bucket) (admission.Warnings, error) {
	return nil, nil
}