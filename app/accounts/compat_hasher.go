package accounts

import (
	"crypto/subtle"
	"encoding/hex"
	"regexp"

	"github.com/AlexGalhardo/vukem/auth/hash"
	"golang.org/x/crypto/scrypt"
	"golang.org/x/text/unicode/norm"
)

// Better Auth 1.7.7 hashes (@better-auth/utils 0.4.2, password.node.mjs):
// "<16-byte salt as 32 hex chars>:<64-byte scrypt key as hex>", scrypt
// N=16384 r=16 p=1 over the NFKC-normalised password. The salt fed to
// scrypt is the hex TEXT, not the decoded bytes.
var betterAuthHash = regexp.MustCompile(`^[0-9a-f]{32}:[0-9a-f]{128}$`)

const (
	betterAuthN     = 16384
	betterAuthR     = 16
	betterAuthP     = 1
	betterAuthKeyLn = 64
)

// CompatHasher verifies legacy Better Auth scrypt hashes and reports them for
// rehash, so the guard replaces each with the inner hasher's argon2id hash on
// the user's next successful login (ADR 0008). New hashes come from inner.
type CompatHasher struct {
	inner hash.Hasher
}

// NewCompatHasher wraps inner (hash.Default() in production).
func NewCompatHasher(inner hash.Hasher) CompatHasher {
	return CompatHasher{inner: inner}
}

// Make hashes with the inner hasher.
func (h CompatHasher) Make(password string) (string, error) {
	return h.inner.Make(password)
}

// Verify checks Better Auth hashes itself and everything else with inner.
func (h CompatHasher) Verify(stored, password string) (bool, error) {
	if !betterAuthHash.MatchString(stored) {
		return h.inner.Verify(stored, password)
	}

	salt, want := stored[:32], stored[33:]
	key, err := scrypt.Key([]byte(norm.NFKC.String(password)), []byte(salt),
		betterAuthN, betterAuthR, betterAuthP, betterAuthKeyLn)
	if err != nil {
		return false, err
	}

	// The pattern guarantees valid hex.
	wantKey, _ := hex.DecodeString(want)

	return subtle.ConstantTimeCompare(key, wantKey) == 1, nil
}

// NeedsRehash is always true for Better Auth hashes.
func (h CompatHasher) NeedsRehash(stored string) bool {
	return betterAuthHash.MatchString(stored) || h.inner.NeedsRehash(stored)
}
