package s3client

import (
	"context"

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

func (c *MinioAdminClient) UserExists(ctx context.Context, accessKey string) (bool, error) {
	users, err := c.Client.ListUsers(context.Background())
	if err != nil {
		return false, err
	}
	for k := range users {
		if k == accessKey {
			return true, nil
		}
	}
	return false, nil
}

func (c *MinioAdminClient) MakeUser(ctx context.Context, accessKey string, secretKey string) error {
	return c.Client.AddUser(context.Background(), accessKey, secretKey)
}

func (c *MinioAdminClient) RemoveUser(ctx context.Context, accessKey string) error {
	return c.Client.RemoveUser(context.Background(), accessKey)
}
