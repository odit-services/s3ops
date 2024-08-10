package mocks

import (
	"context"
	"time"

	"github.com/odit-services/s3ops/api/v1alpha1"
	s3client "github.com/odit-services/s3ops/internal/controller/shared"
)

type S3ClientFactoryMocked struct{}

func (f *S3ClientFactoryMocked) NewClient(s3Server v1alpha1.S3Server) (s3client.S3Client, error) {
	return &S3ClientMocked{}, nil
}

type S3ClientMocked struct {
}

func (c *S3ClientMocked) HealthCheck(time.Duration) (context.CancelFunc, error) {
	return nil, nil
}

func (c *S3ClientMocked) IsOnline() bool {
	return true
}
