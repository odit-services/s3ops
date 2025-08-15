package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getSecret(ctx context.Context, r client.Client, namespace string, name string) (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	err := r.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}, secret)
	if err != nil {
		return secret, err
	}

	return secret, nil
}

func createSecret(ctx context.Context, r client.Client, secret *corev1.Secret) error {
	err := r.Create(ctx, secret)
	if err != nil {
		return err
	}

	return nil
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
	PolicyReadWriteIonos = `
		{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Sid": "Grant Full Control",
					"Effect": "Allow",
					"Principal": {
						"AWS": [
							"arn:aws:iam:::user/%s"
						]
					},
					"Action": "s3:*",
					"Resource": [
						"arn:aws:s3:::%s",
						"arn:aws:s3:::%s/*"
					]
				},
				{
					"Sid": "Grant Full Control",
					"Effect": "Allow",
					"Principal": {
						"AWS": [
							"arn:aws:iam:::user/%s"
						]
					},
					"Action": "s3:GetBucketLocation",
					"Resource": [
						"arn:aws:s3:::*"
					]
				},
				{
					"Sid": "AllowGetBucketLocation",
					"Effect": "Allow",
					"Principal": "*",
					"Action": "s3:GetBucketLocation",
					"Resource": "arn:aws:s3:::*"
				},
				{
					"Sid": "AllowGetBucketLocation",
					"Effect": "Allow",
					"Principal": "*",
					"Action": "s3:ListBucket",
					"Resource": "arn:aws:s3:::*"
				}
			]
		}
	`
)
