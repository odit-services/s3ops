package mocks

import (
	"context"
	"log"
	"slices"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/odit-services/s3ops/api/v1alpha1"
	s3client "github.com/odit-services/s3ops/internal/controller/shared"
)

type S3ClientFactoryMocked struct {
	S3ClientMockEnv *S3ClientMockEnv
}

type S3ClientMockEnv struct {
	ValidEndpoints   []string
	ValidCredentials []v1alpha1.S3ServerAuthSpec
}

type S3Credentials struct {
	AccessKey string
	SecretKey string
}

func (f *S3ClientFactoryMocked) NewClient(s3Server v1alpha1.S3Server) (s3client.S3Client, error) {
	return &S3ClientMocked{
		S3ClientMockEnv: f.S3ClientMockEnv,
		Endpoint:        s3Server.Spec.Endpoint,
		Options: &minio.Options{
			Creds:  credentials.NewStaticV4(s3Server.Spec.Auth.AccessKey, s3Server.Spec.Auth.SecretKey, ""),
			Secure: s3Server.Spec.TLS,
		},
	}, nil
}

type S3ClientMocked struct {
	S3ClientMockEnv *S3ClientMockEnv
	Endpoint        string
	Options         *minio.Options
}

func (c *S3ClientMocked) HealthCheck(time.Duration) (context.CancelFunc, error) {
	return nil, nil
}

func (c *S3ClientMocked) IsOnline() bool {
	if !slices.Contains(c.S3ClientMockEnv.ValidEndpoints, c.Endpoint) {
		log.Printf("Invalid endpoint %s", c.Endpoint)
		return false
	}

	credValue, _ := c.Options.Creds.Get()
	if !slices.ContainsFunc(c.S3ClientMockEnv.ValidCredentials, func(cred v1alpha1.S3ServerAuthSpec) bool {
		return cred.AccessKey == credValue.AccessKeyID && cred.SecretKey == credValue.SecretAccessKey
	}) {
		log.Printf("Invalid credentials %s, %s for endpoint %s", credValue.AccessKeyID, credValue.SecretAccessKey, c.Endpoint)
		return false
	}
	return true
}
