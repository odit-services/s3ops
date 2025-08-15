package s3client

import (
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/odit-services/s3ops/api/v1alpha1"
)

type S3ClientFactory interface {
	NewClient(s3Server v1alpha1.S3Server) (S3Client, error)
	NewAdminClient(s3Server v1alpha1.S3Server) (S3AdminClient, error)
}
type S3ClientFactoryDefault struct{}

func (f *S3ClientFactoryDefault) NewClient(s3Server v1alpha1.S3Server) (S3Client, error) {
	switch s3Server.Spec.Type {
	case "minio":
		return minio.New(s3Server.Spec.Endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(s3Server.Spec.Auth.AccessKey, s3Server.Spec.Auth.SecretKey, ""),
			Secure: s3Server.Spec.TLS,
		})
	case "ionos":
		return NewIonosClient(s3Server.Spec.Endpoint, s3Server.Spec.Auth.AccessKey, s3Server.Spec.Auth.SecretKey, s3Server.Spec.TLS)
	default:
		return nil, fmt.Errorf("not implemented")
	}
}

func (f *S3ClientFactoryDefault) NewAdminClient(s3Server v1alpha1.S3Server) (S3AdminClient, error) {
	switch s3Server.Spec.Type {
	case "minio":
		return NewMinioAdminClient(s3Server.Spec.Endpoint, s3Server.Spec.Auth.AccessKey, s3Server.Spec.Auth.SecretKey, s3Server.Spec.TLS)
	case "ionos":
		return NewIonosAdminClient(s3Server.Spec.Endpoint, s3Server.Spec.Auth.AccessKey, s3Server.Spec.Auth.SecretKey, s3Server.Spec.Auth.ApiToken, s3Server.Spec.TLS)
	default:
		return nil, fmt.Errorf("not implemented")
	}
}

type S3Client interface {
	HealthCheck(time.Duration) (context.CancelFunc, error)
	IsOnline() bool
	BucketExists(context.Context, string) (bool, error)
	MakeBucket(context.Context, string, minio.MakeBucketOptions) error
	RemoveBucket(context.Context, string) error
}

type S3AdminClient interface {
	UserExists(context.Context, string) (bool, error)
	MakeUser(context.Context, string) (string, string, string, error)
	RemoveUser(context.Context, string) error
	PolicyExists(context.Context, string) (bool, error)
	MakePolicy(context.Context, string, string) error
	UpdatePolicy(context.Context, string, string) error
	RemovePolicy(context.Context, string) error
	ApplyPolicyToUser(context.Context, string, string) error
	GetType() string
}
