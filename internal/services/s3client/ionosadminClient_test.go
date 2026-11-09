package s3client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIonosContractNumberFromProviderOptions(t *testing.T) {
	tests := []struct {
		name string
		opts map[string]string
		want string
	}{
		{name: "canonical key", opts: map[string]string{"contractNumber": "12345678"}, want: "12345678"},
		{name: "alias contractId", opts: map[string]string{"contractId": "87654321"}, want: "87654321"},
		{name: "alias contract", opts: map[string]string{"contract": "11111111"}, want: "11111111"},
		{name: "alias contract_number", opts: map[string]string{"contract_number": "22222222"}, want: "22222222"},
		{name: "trims whitespace", opts: map[string]string{"contractNumber": "  12345678  "}, want: "12345678"},
		{name: "canonical wins over aliases", opts: map[string]string{"contractNumber": "1111", "contractId": "2222"}, want: "1111"},
		{name: "skips empty values", opts: map[string]string{"contractNumber": "", "contractId": "3333"}, want: "3333"},
		{name: "nil map", opts: nil, want: ""},
		{name: "empty map", opts: map[string]string{}, want: ""},
		{name: "unrelated keys only", opts: map[string]string{"s3GroupId": "abc"}, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IonosContractNumberFromProviderOptions(tt.opts); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestIonosContractUserID(t *testing.T) {
	tests := []struct {
		name     string
		contract string
		userID   string
		want     string
	}{
		{name: "prefixes plain uuid", contract: "12345678", userID: "abc-uuid", want: "12345678:abc-uuid"},
		{name: "already prefixed is unchanged", contract: "12345678", userID: "12345678:abc-uuid", want: "12345678:abc-uuid"},
		{name: "foreign prefix is kept", contract: "12345678", userID: "99999999:abc-uuid", want: "99999999:abc-uuid"},
		{name: "trims whitespace", contract: "  12345678  ", userID: "  abc-uuid  ", want: "12345678:abc-uuid"},
		{name: "empty contract returns user id", contract: "", userID: "abc-uuid", want: "abc-uuid"},
		{name: "empty user id returns empty", contract: "12345678", userID: "", want: ""},
		{name: "both empty returns empty", contract: "", userID: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IonosContractUserID(tt.contract, tt.userID); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestNewIonosAdminClientPopulatesContractNumber(t *testing.T) {
	client, err := NewIonosAdminClient(
		"s3-eu-central-4.ionoscloud.com",
		"access-key",
		"secret-key",
		"api-token",
		true,
		map[string]string{"s3GroupId": "group-id", "contractNumber": "12345678"},
	)
	if err != nil {
		t.Fatal(err)
	}
	if client.ContractNumber != "12345678" {
		t.Fatalf("expected contract number %q, got %q", "12345678", client.ContractNumber)
	}
	if client.S3GroupId != "group-id" {
		t.Fatalf("expected s3 group id %q, got %q", "group-id", client.S3GroupId)
	}
}

func TestResolveContractNumberReturnsConfiguredWithoutHTTP(t *testing.T) {
	// Point at an address that would fail if ever called: the configured
	// value must win without any HTTP request.
	client := &IonosAdminClient{
		HttpClient:     http.DefaultClient,
		ApiUrl:         "http://127.0.0.1:1",
		ApiToken:       "token",
		ContractNumber: "12345678",
	}

	got, err := client.ResolveContractNumber(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got != "12345678" {
		t.Fatalf("expected %q, got %q", "12345678", got)
	}
}

func TestResolveContractNumberLookupSingleContract(t *testing.T) {
	var gotAuth, gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[{"properties":{"contractNumber":12345678}}]}`))
	}))
	defer server.Close()

	client := &IonosAdminClient{
		HttpClient: server.Client(),
		ApiUrl:     server.URL,
		ApiToken:   "test-token",
	}

	got, err := client.ResolveContractNumber(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got != "12345678" {
		t.Fatalf("expected %q, got %q", "12345678", got)
	}
	if gotAuth != "Bearer test-token" {
		t.Fatalf("expected Bearer auth header, got %q", gotAuth)
	}
	if gotPath != "/contracts" {
		t.Fatalf("expected path /contracts, got %q", gotPath)
	}
}

func TestResolveContractNumberLookupMultipleContracts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[{"properties":{"contractNumber":1111}},{"properties":{"contractNumber":2222}}]}`))
	}))
	defer server.Close()

	client := &IonosAdminClient{
		HttpClient: server.Client(),
		ApiUrl:     server.URL,
		ApiToken:   "test-token",
	}

	_, err := client.ResolveContractNumber(context.Background())
	if err == nil {
		t.Fatal("expected error for multiple contracts")
	}
	if !strings.Contains(err.Error(), "providerOptions.contractNumber") {
		t.Fatalf("expected error to mention providerOptions.contractNumber, got: %v", err)
	}
}

func TestResolveContractNumberLookupNoContracts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	defer server.Close()

	client := &IonosAdminClient{
		HttpClient: server.Client(),
		ApiUrl:     server.URL,
		ApiToken:   "test-token",
	}

	_, err := client.ResolveContractNumber(context.Background())
	if err == nil {
		t.Fatal("expected error for zero contracts")
	}
}

func TestResolveContractNumberLookupHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := &IonosAdminClient{
		HttpClient: server.Client(),
		ApiUrl:     server.URL,
		ApiToken:   "test-token",
	}

	_, err := client.ResolveContractNumber(context.Background())
	if err == nil {
		t.Fatal("expected error for HTTP 500")
	}
	if !strings.Contains(err.Error(), "failed to fetch contracts") {
		t.Fatalf("unexpected error: %v", err)
	}
}
