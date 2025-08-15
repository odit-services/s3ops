package s3client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	ionoscloud "github.com/ionos-cloud/sdk-go-object-storage"
	gonanoid "github.com/matoous/go-nanoid/v2"
	gopassword "github.com/sethvargo/go-password/password"
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

type IonosCreateUserRequest struct {
	Properties struct {
		Firstname     string `json:"firstname"`
		Lastname      string `json:"lastname"`
		Email         string `json:"email"`
		Administrator bool   `json:"administrator"`
		Password      string `json:"password"`
		Active        bool   `json:"active"`
	} `json:"properties"`
}

type IonosCreateUserResponse struct {
	Id string `json:"id"`
}

type IonosCreateUserS3KeyResponse struct {
	Id         string `json:"id"`
	Properties struct {
		SecretKey string `json:"secretKey"`
	} `json:"properties"`
}

func NewIonosAdminClient(endpoint string, accessKey string, secretKey string, apiToken string, tls bool) (*IonosAdminClient, error) {
	client, err := GenerateIonosClient(endpoint, accessKey, secretKey)
	httpClient := &http.Client{}
	return &IonosAdminClient{
		Client:     client,
		HttpClient: httpClient,
		ApiUrl:     "https://api.ionos.com/cloudapi/v6",
		ApiToken:   apiToken,
	}, err
}

func (c *IonosAdminClient) GetType() string {
	return "ionos"
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
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return IonosUserListResponse{}, errors.New("failed to fetch users: " + resp.Status)
	}

	var userList IonosUserListResponse
	if err := json.NewDecoder(resp.Body).Decode(&userList); err != nil {
		return IonosUserListResponse{}, err
	}
	return userList, nil
}

func (c *IonosAdminClient) UserExists(ctx context.Context, userId string) (bool, error) {
	userList, err := c.ListUsers()
	if err != nil {
		return false, err
	}
	for _, user := range userList.Items {
		if user.Id == userId {
			return true, nil
		}
	}
	return false, nil
}

func (c *IonosAdminClient) MakeUser(ctx context.Context, name string) (string, string, string, error) {
	nanoid, err := gonanoid.Generate("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890", 5)
	if err != nil {
		return "", "", "", err
	}
	name = name + "-" + nanoid

	password, err := gopassword.Generate(64, 20, 0, true, true)
	if err != nil {
		return "", "", "", err
	}

	createUserRequest := IonosCreateUserRequest{
		Properties: struct {
			Firstname     string `json:"firstname"`
			Lastname      string `json:"lastname"`
			Email         string `json:"email"`
			Administrator bool   `json:"administrator"`
			Password      string `json:"password"`
			Active        bool   `json:"active"`
		}{
			Firstname:     name,
			Lastname:      name,
			Email:         name + "@s3.insfx.com",
			Administrator: false,
			Password:      password,
			Active:        true,
		},
	}
	reqBody, err := json.Marshal(createUserRequest)
	if err != nil {
		return "", "", "", err
	}
	req, err := http.NewRequest("POST", c.ApiUrl+"/um/users", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.ApiToken)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return "", "", "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", "", errors.New("failed to create user: " + resp.Status)
	}

	var createUserResponse IonosCreateUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&createUserResponse); err != nil {
		return "", "", "", err
	}

	req, err = http.NewRequest("POST", fmt.Sprintf("%s/um/users/%s/s3keys", c.ApiUrl, createUserResponse.Id), nil)
	if err != nil {
		return "", "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.ApiToken)

	resp, err = c.HttpClient.Do(req)
	if err != nil {
		return "", "", "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", "", errors.New("failed to create s3key: " + resp.Status + " for user " + createUserResponse.Id)
	}

	var createUserS3KeyResponse IonosCreateUserS3KeyResponse
	if err := json.NewDecoder(resp.Body).Decode(&createUserS3KeyResponse); err != nil {
		return "", "", "", err
	}

	return createUserResponse.Id, createUserS3KeyResponse.Id, createUserS3KeyResponse.Properties.SecretKey, nil
}

func (c *IonosAdminClient) RemoveUser(ctx context.Context, identifier string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/um/users/%s", c.ApiUrl, identifier), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.ApiToken)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New("failed to create user: " + resp.Status)
	}

	return nil
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
