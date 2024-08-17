package controller

import (
	"context"

	"github.com/odit-services/s3ops/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getSecret(ctx context.Context, r client.Client, namespace string, name string) (*corev1.Secret, metav1.Condition, error) {
	secret := &corev1.Secret{}
	err := r.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}, secret)
	if err != nil {
		err = r.Get(ctx, client.ObjectKey{
			Namespace: namespace,
			Name:      name,
		}, secret)
		if err != nil {
			return secret, metav1.Condition{
				Type:               v1alpha1.ConditionFailed,
				Status:             metav1.ConditionFalse,
				Reason:             v1alpha1.ReasonNotFound,
				Message:            "Secret not found",
				LastTransitionTime: metav1.Now(),
			}, err
		}
	}

	return secret, metav1.Condition{}, nil
}

func createSecret(ctx context.Context, r client.Client, secret *corev1.Secret) (metav1.Condition, error) {
	err := r.Create(ctx, secret)
	if err != nil {
		return metav1.Condition{
			Type:               v1alpha1.ConditionFailed,
			Status:             metav1.ConditionFalse,
			Reason:             v1alpha1.ReasonCreateFailed,
			Message:            "Failed to create secret",
			LastTransitionTime: metav1.Now(),
		}, err
	}

	return metav1.Condition{}, nil
}

const (
	PolicyReadWrite = `
		{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Action": [
						"s3:*"
					],
					"Resource": [
						"arn:aws:s3:::%s",
						"arn:aws:s3:::%s/*"
					]
				}
			]
		}
	`
)
