package s3client

import (
	"context"
	"errors"

	ionoscloud "github.com/ionos-cloud/sdk-go-object-storage"
)

type IonosAdminClient struct {
	Client *ionoscloud.APIClient
}

func NewIonosAdminClient(endpoint string, accessKey string, secretKey string, tls bool) (*IonosAdminClient, error) {
	client, err := GenerateIonosClient(endpoint, accessKey, secretKey)
	return &IonosAdminClient{
		Client: client,
	}, err
}

func (c *IonosAdminClient) UserExists(ctx context.Context, accessKey string) (bool, error) {
	//TODO:
	return false, nil
}

func (c *IonosAdminClient) MakeUser(ctx context.Context, accessKey string, secretKey string) error {
	//TODO:
	return errors.New("MakeUser not implemented")
}

func (c *IonosAdminClient) RemoveUser(ctx context.Context, accessKey string) error {
	//TODO:
	return errors.New("RemoveUser not implemented")
}

func (c *IonosAdminClient) PolicyExists(ctx context.Context, policyName string) (bool, error) {
	//TODO:
	return false, errors.New("PolicyExists not implemented")
}

func (c *IonosAdminClient) MakePolicy(ctx context.Context, policyName string, policy string) error {
	return errors.New("MakePolicy not implemented")
}

func (c *IonosAdminClient) UpdatePolicy(ctx context.Context, policyName string, policy string) error {
	return errors.New("UpdatePolicy not implemented")
}

func (c *IonosAdminClient) RemovePolicy(ctx context.Context, policyName string) error {
	return errors.New("RemovePolicy not implemented")
}

func (c *IonosAdminClient) ApplyPolicyToUser(ctx context.Context, policyName string, accessKey string) error {
	return errors.New("ApplyPolicyToUser not implemented")
}
