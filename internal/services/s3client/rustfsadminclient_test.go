package s3client

import "testing"

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
