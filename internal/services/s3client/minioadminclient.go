package s3client

import (
	"context"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/minio/madmin-go"
	gopassword "github.com/sethvargo/go-password/password"
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

func (c *MinioAdminClient) GetType() string {
	return "minio"
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

func (c *MinioAdminClient) MakeUser(ctx context.Context, name string) (string, string, string, error) {
	nanoid, err := gonanoid.Generate("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890", 5)
	if err != nil {
		return "", "", "", err
	}
	accessKey := name + "-" + nanoid

	secretKey, err := gopassword.Generate(64, 20, 0, true, true)
	if err != nil {
		return "", "", "", err
	}

	err = c.Client.AddUser(context.Background(), accessKey, secretKey)
	return accessKey, accessKey, secretKey, err
}

func (c *MinioAdminClient) RemoveUser(ctx context.Context, accessKey string) error {
	return c.Client.RemoveUser(context.Background(), accessKey)
}

func (c *MinioAdminClient) PolicyExists(ctx context.Context, policyName string) (bool, error) {
	policies, err := c.Client.ListCannedPolicies(context.Background())
	if err != nil {
		return false, err
	}

	for k := range policies {
		if k == policyName {
			return true, nil
		}
	}
	return false, nil
}

func (c *MinioAdminClient) MakePolicy(ctx context.Context, policyName string, policy string) error {
	return c.Client.AddCannedPolicy(ctx, policyName, []byte(policy))
}

func (c *MinioAdminClient) UpdatePolicy(ctx context.Context, policyName string, policy string) error {
	return c.Client.AddCannedPolicy(ctx, policyName, []byte(policy))
}

func (c *MinioAdminClient) RemovePolicy(ctx context.Context, policyName string) error {
	return c.Client.RemoveCannedPolicy(ctx, policyName)
}

func (c *MinioAdminClient) ApplyPolicyToUser(ctx context.Context, policyName string, accessKey string) error {
	return c.Client.SetPolicy(context.Background(), policyName, accessKey, false)
}
