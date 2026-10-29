package api

import (
	"testing"
)

func TestKeysNewKeyShape(t *testing.T) {
	t.Parallel()

	seen := map[string]bool{}

	for range 50 {
		k, err := NewKey()
		if err != nil {
			t.Fatal(err)
		}

		if !WellFormed(k) {
			t.Fatalf("NewKey() = %q, not well formed", k)
		}

		if seen[k] {
			t.Fatalf("NewKey() repeated %q", k)
		}

		seen[k] = true
	}
}

func TestKeysHashIsLegacySHA256Hex(t *testing.T) {
	t.Parallel()

	// Expected value from `printf %s <key> | sha256sum`, the digest legacy createHash("sha256").digest("hex") stores.
	const key = "hx_live_00000000000000000000000000000000"
	if got, want := HashKey(key), "6841696ec1eb9a1bd7c648d309e2395c729e2fcd01def7055d0b193822d26da6"; got != want {
		t.Fatalf("HashKey = %q, want %q", got, want)
	}
}

func TestKeysWellFormed(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		key  string
		want bool
	}{
		{"hx_live_" + "aZ09aZ09aZ09aZ09aZ09aZ09aZ09aZ09", true},
		{"hx_app_" + "aZ09aZ09aZ09aZ09aZ09aZ09aZ09aZ09", false},
		{"hx_live_" + "aZ09aZ09aZ09aZ09aZ09aZ09aZ09aZ0", false},
		{"hx_live_" + "aZ09aZ09aZ09aZ09aZ09aZ09aZ09aZ09a", false},
		{"hx_live_" + "aZ09aZ09aZ09aZ09aZ09aZ09aZ09aZ0-", false},
		{"", false},
	} {
		if got := WellFormed(tc.key); got != tc.want {
			t.Errorf("WellFormed(%q) = %v, want %v", tc.key, got, tc.want)
		}
	}
}
