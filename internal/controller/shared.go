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

// upsertSecret creates the secret if it does not exist yet, or updates it in
// place if it does. This is needed when a user is recreated after being
// deleted from the backing S3 provider: the Kubernetes secret already exists
// (from the previous incarnation) and must be overwritten with the new
// credentials rather than failing with "already exists".
func upsertSecret(ctx context.Context, r client.Client, secret *corev1.Secret) error {
	existing := &corev1.Secret{}
	err := r.Get(ctx, client.ObjectKey{Namespace: secret.Namespace, Name: secret.Name}, existing)
	if err != nil {
		// Secret doesn't exist yet – create it.
		return r.Create(ctx, secret)
	}
	// Secret exists – overwrite its data with the new credentials.
	existing.StringData = secret.StringData
	existing.Data = secret.Data
	return r.Update(ctx, existing)
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
					"Action": ["s3:*"],
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
					"Action": ["s3:GetBucketLocation"],
					"Resource": [
						"arn:aws:s3:::*"
					]
				},
				{
					"Sid": "Allowall",
					"Effect": "Allow",
					"Principal": {
						"AWS": [
							"*"
						]
					},
					"Action": ["s3:*"],
					"Resource": ["arn:aws:s3:::*"]
				}
			]
		}
	`
)
