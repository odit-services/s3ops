package s3client

import (
	"context"
	"time"

	ionoscloud "github.com/ionos-cloud/sdk-go-object-storage"
	"github.com/minio/minio-go/v7"
)

type IonosClient struct {
	Client *ionoscloud.APIClient
}

func NewIonosClient(endpoint string, accessKey string, secretKey string, tls bool) (*IonosClient, error) {
	client, err := GenerateIonosClient(endpoint, accessKey, secretKey)
	return &IonosClient{
		Client: client,
	}, err
}

func (c *IonosClient) HealthCheck(timeout time.Duration) (context.CancelFunc, error) {
	_, _, err := c.Client.BucketsApi.ListBuckets(context.Background()).Execute()
	return func() {}, err
}

func (c *IonosClient) IsOnline() bool {
	_, err := c.HealthCheck(5 * time.Second)
	return err == nil
}

func (c *IonosClient) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	resource, _, err := c.Client.BucketsApi.ListBuckets(ctx).Execute()
	if err != nil {
		return false, err
	}
	for _, bucket := range *resource.Buckets {
		if *bucket.Name == bucketName {
			return true, nil
		}
	}
	return false, nil
}

func (c *IonosClient) MakeBucket(ctx context.Context, bucketName string, options minio.MakeBucketOptions) error {
	_, err := c.Client.BucketsApi.CreateBucket(ctx, bucketName).CreateBucketConfiguration(*ionoscloud.NewCreateBucketConfiguration()).XAmzBucketObjectLockEnabled(options.ObjectLocking).Execute()
	if err != nil {
		return err
	}
	return nil
}

func (c *IonosClient) RemoveBucket(ctx context.Context, bucketName string) error {
	_, err := c.Client.BucketsApi.DeleteBucket(ctx, bucketName).Execute()
	if err != nil {
		return err
	}
	return nil
}
