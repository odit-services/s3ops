package s3client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestShortenRustfsServiceAccountName(t *testing.T) {
	t.Run("keeps short names unchanged", func(t *testing.T) {
		name := "short-name"
		if got := shortenRustfsServiceAccountName(name); got != name {
			t.Fatalf("expected %q, got %q", name, got)
		}
	})

	t.Run("shortens long names to 31 chars deterministically", func(t *testing.T) {
		name := "choice-car-bonus-staging-file-storage"
		got := shortenRustfsServiceAccountName(name)

		if len(got) != rustfsNameMaxLen {
			t.Fatalf("expected length %d, got %d (%q)", rustfsNameMaxLen, len(got), got)
		}
		if got != "choice-car-bonus-staging-aad27b" {
			t.Fatalf("unexpected shortened name: %q", got)
		}
	})
}

func TestMakeUserWithCredentialsUsesSuppliedCredentials(t *testing.T) {
	const (
		accessKey = "persisted-access-key"
		secretKey = "persisted-secret-key"
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/rustfs/admin/v3/add-service-accounts" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		var request rustfsAddServiceAccountReq
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		if request.AccessKey != accessKey || request.SecretKey != secretKey {
			t.Fatalf("expected persisted credentials, got accessKey=%q secretKey=%q", request.AccessKey, request.SecretKey)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"credentials":{"accessKey":"persisted-access-key","secretKey":"persisted-secret-key"}}`))
	}))
	defer server.Close()

	client, err := NewRustfsAdminClient("unused", "admin-access-key", "admin-secret-key", false, map[string]string{"consoleEndpoint": server.URL})
	if err != nil {
		t.Fatal(err)
	}

	identifier, gotAccessKey, gotSecretKey, err := client.MakeUserWithCredentials(t.Context(), "test-user", accessKey, secretKey)
	if err != nil {
		t.Fatal(err)
	}
	if identifier != accessKey || gotAccessKey != accessKey || gotSecretKey != secretKey {
		t.Fatalf("unexpected credentials: identifier=%q accessKey=%q secretKey=%q", identifier, gotAccessKey, gotSecretKey)
	}
}

func TestCheckRustfsHealth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/health" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	if err := CheckRustfsHealth(t.Context(), server.URL, false); err != nil {
		t.Fatal(err)
	}
}

func TestCheckRustfsHealthRejectsNonSuccessStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	if err := CheckRustfsHealth(t.Context(), server.URL, false); err == nil {
		t.Fatal("expected health check to fail")
	}
}
