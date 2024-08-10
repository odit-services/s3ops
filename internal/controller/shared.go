package controller

import (
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/odit-services/s3ops/api/v1alpha1"
)

func GenerateMinioClientFromS3Server(s3Server v1alpha1.S3Server) (*minio.Client, error) {
	return minio.New(s3Server.Spec.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s3Server.Spec.Auth.AccessKey, s3Server.Spec.Auth.SecretKey, ""),
		Secure: s3Server.Spec.TLS,
	})
}
