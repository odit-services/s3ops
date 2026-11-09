package controller

import (
	"encoding/json"
	"strings"
	"testing"
)

// Regression test for the IONOS contract-owned bucket policy: the principal
// must be arn:aws:iam:::user/<contractNumber>:<userUUID>. A plain user UUID
// matches no principal; this was previously masked by an AllowAll statement.
func TestBuildIonosBucketPolicyUsesContractUserID(t *testing.T) {
	policy := BuildIonosBucketPolicy("12345678", "abc-uuid", "my-bucket")

	const principal = "arn:aws:iam:::user/12345678:abc-uuid"
	if got := strings.Count(policy, principal); got != 2 {
		t.Fatalf("expected principal %q twice (both statements), found %d times in:\n%s", principal, got, policy)
	}

	for _, want := range []string{"arn:aws:s3:::my-bucket", "arn:aws:s3:::my-bucket/*"} {
		if !strings.Contains(policy, want) {
			t.Fatalf("expected policy to contain resource %q, got:\n%s", want, policy)
		}
	}

	// The bare UUID without contract prefix must not appear as a principal.
	if strings.Contains(policy, "arn:aws:iam:::user/abc-uuid\"") {
		t.Fatalf("policy must not contain the bare user UUID as principal:\n%s", policy)
	}

	// The rendered policy must be valid JSON.
	var decoded map[string]any
	if err := json.Unmarshal([]byte(policy), &decoded); err != nil {
		t.Fatalf("rendered policy is not valid JSON: %v\n%s", err, policy)
	}
}

func TestBuildIonosBucketPolicyIsIdempotentForPrefixedUserID(t *testing.T) {
	plain := BuildIonosBucketPolicy("12345678", "abc-uuid", "my-bucket")
	prefixed := BuildIonosBucketPolicy("12345678", "12345678:abc-uuid", "my-bucket")

	if plain != prefixed {
		t.Fatalf("expected identical policy for plain and prefixed user ID, got:\n%s\n---\n%s", plain, prefixed)
	}
}

func TestBuildIonosBucketPolicyTrimsWhitespace(t *testing.T) {
	policy := BuildIonosBucketPolicy("  12345678  ", "  abc-uuid  ", "my-bucket")

	if !strings.Contains(policy, "arn:aws:iam:::user/12345678:abc-uuid") {
		t.Fatalf("expected trimmed principal in policy, got:\n%s", policy)
	}
}
