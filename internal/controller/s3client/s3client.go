package s3client

import (
	"context"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/odit-services/s3ops/api/v1alpha1"
)

type S3ClientFactory interface {
	NewClient(s3Server v1alpha1.S3Server) (S3Client, error)
}
type S3ClientFactoryDefault struct{}

func (f *S3ClientFactoryDefault) NewClient(s3Server v1alpha1.S3Server) (S3Client, error) {
	return minio.New(s3Server.Spec.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s3Server.Spec.Auth.AccessKey, s3Server.Spec.Auth.SecretKey, ""),
		Secure: s3Server.Spec.TLS,
	})
}

type S3Client interface {
	HealthCheck(time.Duration) (context.CancelFunc, error)
	IsOnline() bool
	BucketExists(context.Context, string) (bool, error)
	MakeBucket(context.Context, string, minio.MakeBucketOptions) error
	RemoveBucket(context.Context, string) error
}
