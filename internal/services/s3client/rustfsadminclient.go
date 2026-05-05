package s3client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
	gopassword "github.com/sethvargo/go-password/password"

	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
)

const (
	rustfsAdminPrefix  = "/rustfs/admin/v3"
	rustfsNeverExpires = "9999-01-01T00:00:00.000Z"
)

// RustfsAdminClient implements S3AdminClient for RustFS.
//
// User management is done via service accounts on RustFS's console admin API
// (port 9001, path prefix /rustfs/admin/v3/).  Each "user" the operator
// creates is a RustFS service account with an inline session policy that
// grants the requested permissions.
//
// Canned policy management uses the same admin API
// (add-canned-policy / remove-canned-policy / list-canned-policies).
// ApplyPolicyToUser fetches the named canned policy and re-embeds it as an
// inline policy on the service account via update-service-account, because
// RustFS does not support set-user-or-group-policy for service accounts.
type RustfsAdminClient struct {
	httpClient      *http.Client
	consoleEndpoint string // e.g. "http://localhost:9001"
	accessKey       string
	secretKey       string
	signer          *v4.Signer
}

// rustfsAddServiceAccountReq matches the JSON body for PUT /rustfs/admin/v3/add-service-accounts.
type rustfsAddServiceAccountReq struct {
	AccessKey   string  `json:"accessKey"`
	SecretKey   string  `json:"secretKey"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Policy      *string `json:"policy,omitempty"`
	Expiration  string  `json:"expiration"`
}

// rustfsUpdateServiceAccountReq matches the JSON body for POST /rustfs/admin/v3/update-service-account.
type rustfsUpdateServiceAccountReq struct {
	NewPolicy     *string `json:"newPolicy,omitempty"`
	NewExpiration string  `json:"newExpiration"`
}

// rustfsAddServiceAccountResp matches the JSON response from add-service-accounts.
type rustfsAddServiceAccountResp struct {
	Credentials struct {
		AccessKey string `json:"accessKey"`
		SecretKey string `json:"secretKey"`
	} `json:"credentials"`
}

// rustfsListPoliciesResp is a map of policyName -> raw policy object.
type rustfsListPoliciesResp map[string]json.RawMessage

// NewRustfsAdminClient creates a RustfsAdminClient.
//
// s3Endpoint is the S3 API endpoint (e.g. "localhost:9000").
// consoleEndpoint is the console/admin endpoint including scheme
// (e.g. "http://localhost:9001").  If empty, it is derived from s3Endpoint
// by replacing a trailing :9000 with :9001 (or appending :9001).
func NewRustfsAdminClient(s3Endpoint string, accessKey string, secretKey string, tls bool, providerOptions map[string]string) (*RustfsAdminClient, error) {
	consoleEndpoint := providerOptions["consoleEndpoint"]
	if consoleEndpoint == "" {
		scheme := "http"
		if tls {
			scheme = "https"
		}
		// Derive console port from S3 port: replace :9000 → :9001, else append :9001.
		host := s3Endpoint
		if strings.HasSuffix(host, ":9000") {
			host = strings.TrimSuffix(host, ":9000") + ":9001"
		} else {
			// Strip any existing port and use 9001.
			if idx := strings.LastIndex(host, ":"); idx != -1 {
				host = host[:idx] + ":9001"
			} else {
				host = host + ":9001"
			}
		}
		consoleEndpoint = scheme + "://" + host
	}

	creds := credentials.NewStaticCredentials(accessKey, secretKey, "")
	signer := v4.NewSigner(creds)

	return &RustfsAdminClient{
		httpClient:      &http.Client{},
		consoleEndpoint: consoleEndpoint,
		accessKey:       accessKey,
		secretKey:       secretKey,
		signer:          signer,
	}, nil
}

func (c *RustfsAdminClient) GetType() string {
	return "rustfs"
}

// doRequest signs and executes an admin API request.
func (c *RustfsAdminClient) doRequest(ctx context.Context, method, path string, body []byte) (*http.Response, error) {
	url := c.consoleEndpoint + rustfsAdminPrefix + path

	var bodyReader *bytes.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Reset reader for signing.
	_, _ = bodyReader.Seek(0, 0)
	_, err = c.signer.Sign(req, bodyReader, "s3", "us-east-1", time.Now())
	if err != nil {
		return nil, fmt.Errorf("rustfs: failed to sign request: %w", err)
	}

	return c.httpClient.Do(req)
}

// checkStatus returns an error if the response status is not 2xx.
func checkStatus(resp *http.Response, operation string) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("rustfs: %s returned HTTP %d and the error body could not be read: %w", operation, resp.StatusCode, err)
	}

	bodyText := strings.TrimSpace(string(body))
	if bodyText == "" {
		return fmt.Errorf("rustfs: %s returned HTTP %d", operation, resp.StatusCode)
	}
	if len(bodyText) > 1024 {
		bodyText = bodyText[:1024] + "...(truncated)"
	}

	return fmt.Errorf("rustfs: %s returned HTTP %d: %s", operation, resp.StatusCode, bodyText)
}

// UserExists checks whether a service account with the given access key exists.
// It uses info-service-account: 200 = exists, anything else = does not exist.
func (c *RustfsAdminClient) UserExists(ctx context.Context, accessKey string) (bool, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/info-service-account?accessKey="+accessKey, nil)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK, nil
}

// MakeUser creates a new RustFS service account and returns
// (providerID, accessKey, secretKey, error).
// providerID == accessKey for RustFS (there is no separate internal ID).
func (c *RustfsAdminClient) MakeUser(ctx context.Context, name string) (string, string, string, error) {
	nanoid, err := gonanoid.Generate("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890", 20)
	if err != nil {
		return "", "", "", err
	}
	accessKey := nanoid

	secretKey, err := gopassword.Generate(40, 10, 0, false, true)
	if err != nil {
		return "", "", "", err
	}

	reqBody := rustfsAddServiceAccountReq{
		AccessKey:   accessKey,
		SecretKey:   secretKey,
		Name:        name,
		Description: "",
		Expiration:  rustfsNeverExpires,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", "", "", err
	}

	resp, err := c.doRequest(ctx, http.MethodPut, "/add-service-accounts", bodyBytes)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()
	if err := checkStatus(resp, "add-service-accounts"); err != nil {
		return "", "", "", fmt.Errorf("%w (name=%q expiration=%q accessKeyLength=%d)", err, name, reqBody.Expiration, len(accessKey))
	}

	var result rustfsAddServiceAccountResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", "", fmt.Errorf("rustfs: failed to decode add-service-accounts response: %w", err)
	}

	return result.Credentials.AccessKey, result.Credentials.AccessKey, result.Credentials.SecretKey, nil
}

// RemoveUser deletes the service account identified by accessKey.
func (c *RustfsAdminClient) RemoveUser(ctx context.Context, accessKey string) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, "/delete-service-accounts?accessKey="+accessKey, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkStatus(resp, "delete-service-accounts")
}

// PolicyExists returns true if the named canned policy exists.
func (c *RustfsAdminClient) PolicyExists(ctx context.Context, policyName string) (bool, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/list-canned-policies", nil)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if err := checkStatus(resp, "list-canned-policies"); err != nil {
		return false, err
	}

	var policies rustfsListPoliciesResp
	if err := json.NewDecoder(resp.Body).Decode(&policies); err != nil {
		return false, fmt.Errorf("rustfs: failed to decode list-canned-policies response: %w", err)
	}
	_, exists := policies[policyName]
	return exists, nil
}

// MakePolicy creates a named canned policy.
func (c *RustfsAdminClient) MakePolicy(ctx context.Context, policyName string, policy string) error {
	resp, err := c.doRequest(ctx, http.MethodPut, "/add-canned-policy?name="+policyName, []byte(policy))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkStatus(resp, "add-canned-policy")
}

// UpdatePolicy replaces an existing canned policy by re-creating it.
func (c *RustfsAdminClient) UpdatePolicy(ctx context.Context, policyName string, policy string) error {
	return c.MakePolicy(ctx, policyName, policy)
}

// RemovePolicy deletes the named canned policy.
func (c *RustfsAdminClient) RemovePolicy(ctx context.Context, policyName string) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, "/remove-canned-policy?name="+policyName, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkStatus(resp, "remove-canned-policy")
}

// ApplyPolicyToUser fetches the named canned policy and embeds it as an
// inline session policy on the service account via update-service-account.
// RustFS does not support set-user-or-group-policy for service accounts, so
// this is the only way to scope a service account to a specific policy.
func (c *RustfsAdminClient) ApplyPolicyToUser(ctx context.Context, policyName string, accessKey string) error {
	// Fetch the canned policy document.
	resp, err := c.doRequest(ctx, http.MethodGet, "/list-canned-policies", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := checkStatus(resp, "list-canned-policies"); err != nil {
		return err
	}

	var policies rustfsListPoliciesResp
	if err := json.NewDecoder(resp.Body).Decode(&policies); err != nil {
		return fmt.Errorf("rustfs: failed to decode list-canned-policies response: %w", err)
	}

	raw, exists := policies[policyName]
	if !exists {
		return errors.New("rustfs: policy not found: " + policyName)
	}

	// The update endpoint expects the policy as a JSON string (not an object).
	policyStr := string(raw)

	updateReq := rustfsUpdateServiceAccountReq{
		NewPolicy:     &policyStr,
		NewExpiration: rustfsNeverExpires,
	}
	bodyBytes, err := json.Marshal(updateReq)
	if err != nil {
		return err
	}

	updateResp, err := c.doRequest(ctx, http.MethodPost, "/update-service-account?accessKey="+accessKey, bodyBytes)
	if err != nil {
		return err
	}
	defer func() { _ = updateResp.Body.Close() }()
	return checkStatus(updateResp, "update-service-account")
}
