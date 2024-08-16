package s3client

import (
	"context"
	"fmt"

	"github.com/minio/madmin-go"
)

type MinioAdminClient struct {
	Client *madmin.AdminClient
}

func NewMinioAdminClient(endpoint string, accessKey string, secretKey string, tls bool) (*MinioAdminClient, error) {
	client, err := madmin.New(endpoint, accessKey, secretKey, tls)
	if err != nil {
		return nil, err
	}

	return &MinioAdminClient{
		Client: client,
	}, nil
}

func (c *MinioAdminClient) UserExists(ctx context.Context, userName string) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (c *MinioAdminClient) MakeUser(ctx context.Context, userName string, password string) error {
	return fmt.Errorf("not implemented")
}

func (c *MinioAdminClient) RemoveUser(ctx context.Context, userName string) error {
	return fmt.Errorf("not implemented")
}
