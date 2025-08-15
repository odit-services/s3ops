package s3client

import (
	"context"
	"fmt"

	"github.com/odit-services/s3ops/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetS3ServerObject(serverRef v1alpha1.ServerReference, r client.Client) (*v1alpha1.S3Server, error) {
	ctx := context.Background()

	s3Server := &v1alpha1.S3Server{}
	err := r.Get(ctx, client.ObjectKey{
		Namespace: serverRef.Namespace,
		Name:      serverRef.Name,
	}, s3Server)
	if err != nil {
		return nil, err
	}

	if s3Server.Spec.Auth.ExistingSecretRef != "" {
		secret := &corev1.Secret{}
		err := r.Get(ctx, client.ObjectKey{Namespace: s3Server.Namespace, Name: s3Server.Spec.Auth.ExistingSecretRef}, secret)
		if err != nil {
			return &v1alpha1.S3Server{}, err
		}

		if s3Server.Spec.Auth.AccessKeySecretKey == "" {
			s3Server.Spec.Auth.AccessKeySecretKey = "accessKey"
		}
		if s3Server.Spec.Auth.SecretKeySecretKey == "" {
			s3Server.Spec.Auth.SecretKeySecretKey = "secretKey"
		}
		if s3Server.Spec.Auth.ApiTokenSecretKey == "" {
			s3Server.Spec.Auth.ApiTokenSecretKey = "apiToken"
		}

		s3Server.Spec.Auth.AccessKey = string(secret.Data[s3Server.Spec.Auth.AccessKeySecretKey])
		s3Server.Spec.Auth.SecretKey = string(secret.Data[s3Server.Spec.Auth.SecretKeySecretKey])
		s3Server.Spec.Auth.ApiToken = string(secret.Data[s3Server.Spec.Auth.ApiTokenSecretKey])
	}

	return s3Server, nil
}

func GetS3ClientFromS3Server(serverRef v1alpha1.ServerReference, factory S3ClientFactory, r client.Client) (S3Client, error) {
	s3Server, err := GetS3ServerObject(serverRef, r)
	if err != nil {
		return nil, err
	}

	if !s3Server.Status.Online {
		return nil, fmt.Errorf("S3Server is offline")
	}

	s3Client, err := factory.NewClient(*s3Server)
	if err != nil {
		return nil, err
	}

	return s3Client, nil
}

func GetS3AdminClientFromS3Server(serverRef v1alpha1.ServerReference, factory S3ClientFactory, r client.Client) (S3AdminClient, error) {
	s3Server, err := GetS3ServerObject(serverRef, r)
	if err != nil {
		return nil, err
	}

	if !s3Server.Status.Online {
		return nil, fmt.Errorf("S3Server is offline")
	}

	s3AdminClient, err := factory.NewAdminClient(*s3Server)
	if err != nil {
		return nil, err
	}

	return s3AdminClient, nil
}
