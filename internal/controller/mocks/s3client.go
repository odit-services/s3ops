package mocks

import (
	"context"
	"fmt"
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
	S3ClientMockSpy *S3ClientMockSpy
}

type S3ClientMockEnv struct {
	ValidEndpoints   []string
	ValidCredentials []v1alpha1.S3ServerAuthSpec
	ExistingBuckets  []string
}

type S3ClientMockSpy struct {
	HealthCheckCalled  int
	IsOnlineCalled     int
	BucketExistsCalled int
	MakeBucketCalled   int
	RemoveBucketCalled int
}

type S3Credentials struct {
	AccessKey string
	SecretKey string
}

func (f *S3ClientFactoryMocked) NewClient(s3Server v1alpha1.S3Server) (s3client.S3Client, error) {
	return &S3ClientMocked{
		S3ClientMockEnv: f.S3ClientMockEnv,
		S3ClientMockSpy: f.S3ClientMockSpy,
		Endpoint:        s3Server.Spec.Endpoint,
		Options: &minio.Options{
			Creds:  credentials.NewStaticV4(s3Server.Spec.Auth.AccessKey, s3Server.Spec.Auth.SecretKey, ""),
			Secure: s3Server.Spec.TLS,
		},
	}, nil
}

type S3ClientMocked struct {
	S3ClientMockEnv *S3ClientMockEnv
	S3ClientMockSpy *S3ClientMockSpy
	Endpoint        string
	Options         *minio.Options
}

func (c *S3ClientMocked) CheckServerValid() bool {
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

func (c *S3ClientMocked) HealthCheck(time.Duration) (context.CancelFunc, error) {
	c.S3ClientMockSpy.HealthCheckCalled++
	return nil, nil
}

func (c *S3ClientMocked) IsOnline() bool {
	c.S3ClientMockSpy.IsOnlineCalled++
	return c.CheckServerValid()
}

func (c *S3ClientMocked) BucketExists(ctx context.Context, name string) (bool, error) {
	c.S3ClientMockSpy.BucketExistsCalled++
	if !c.CheckServerValid() {
		return false, fmt.Errorf("invalid server")
	}
	return slices.Contains(c.S3ClientMockEnv.ExistingBuckets, name), nil
}

func (c *S3ClientMocked) MakeBucket(context.Context, string, minio.MakeBucketOptions) error {
	c.S3ClientMockSpy.MakeBucketCalled++
	if !c.CheckServerValid() {
		return fmt.Errorf("invalid server")
	}
	return nil
}

func (c *S3ClientMocked) RemoveBucket(context.Context, string) error {
	c.S3ClientMockSpy.RemoveBucketCalled++
	return fmt.Errorf("not implemented")
}
