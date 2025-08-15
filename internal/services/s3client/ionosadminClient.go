package s3client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	ionoscloud "github.com/ionos-cloud/sdk-go-object-storage"
)

type IonosAdminClient struct {
	Client     *ionoscloud.APIClient
	HttpClient *http.Client
	ApiUrl     string
	ApiToken   string
}

type IonosUserListResponse struct {
	Id    string `json:"id"`
	Items []struct {
		Id   string `json:"id"`
		Type string `json:"type"`
		HRef string `json:"href"`
	} `json:"items"`
}

func NewIonosAdminClient(endpoint string, accessKey string, secretKey string, tls bool) (*IonosAdminClient, error) {
	client, err := GenerateIonosClient(endpoint, accessKey, secretKey)
	httpClient := &http.Client{}
	return &IonosAdminClient{
		Client:     client,
		HttpClient: httpClient,
		ApiUrl:     "https://api.ionos.com/cloudapi/v6",
	}, err
}

func (c *IonosAdminClient) ListUsers() (IonosUserListResponse, error) {
	req, err := http.NewRequest("GET", c.ApiUrl+"/um/users", nil)
	if err != nil {
		return IonosUserListResponse{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.ApiToken)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return IonosUserListResponse{}, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return IonosUserListResponse{}, errors.New("failed to fetch users: " + resp.Status)
	}

	var userList IonosUserListResponse
	if err := json.NewDecoder(resp.Body).Decode(&userList); err != nil {
		return IonosUserListResponse{}, err
	}
	return userList, nil
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
