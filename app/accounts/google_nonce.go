package accounts

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// GoogleNonceTTL is how long an issued nonce stays valid: enough for the browser round trip, short enough that a
// leaked one is useless.
const GoogleNonceTTL = 5 * time.Minute

func nonceHash(nonce string) string {
	sum := sha256.Sum256([]byte(nonce))

	return hex.EncodeToString(sum[:])
}

// IssueGoogleNonce creates a 128-bit nonce for one Google sign-in and stores its hash. Expired rows are pruned here,
// so the table stays as small as the sign-in rate times the TTL.
func (a *Accounts) IssueGoogleNonce(ctx context.Context) (string, time.Time, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", time.Time{}, fmt.Errorf("google nonce: %w", err)
	}

	nonce := hex.EncodeToString(raw)
	expires := a.Clock().Add(GoogleNonceTTL)
	ph := a.DB.Dialect().Placeholder

	if _, err := a.DB.ExecContext(ctx, "DELETE FROM google_nonces WHERE expires_at < "+ph(1), a.Clock()); err != nil {
		return "", time.Time{}, fmt.Errorf("google nonce prune: %w", err)
	}

	if _, err := a.DB.ExecContext(ctx,
		"INSERT INTO google_nonces (nonce_hash, expires_at) VALUES ("+ph(1)+", "+ph(2)+")",
		nonceHash(nonce), expires); err != nil {
		return "", time.Time{}, fmt.Errorf("google nonce store: %w", err)
	}

	return nonce, expires, nil
}

// GoogleNonceValid reports whether nonce was issued, is unexpired and is unused. It does not consume it: a sign-in
// that still needs its second factor must be able to retry with the same ID token.
func (a *Accounts) GoogleNonceValid(ctx context.Context, nonce string) (bool, error) {
	if nonce == "" {
		return false, nil
	}

	ph := a.DB.Dialect().Placeholder

	var one int

	rows, err := a.DB.QueryContext(ctx,
		"SELECT 1 FROM google_nonces WHERE nonce_hash = "+ph(1)+" AND expires_at > "+ph(2),
		nonceHash(nonce), a.Clock())
	if err != nil {
		return false, fmt.Errorf("google nonce check: %w", err)
	}
	defer func() { _ = rows.Close() }()

	if !rows.Next() {
		return false, rows.Err()
	}

	return true, rows.Scan(&one)
}

// ConsumeGoogleNonce deletes the nonce and reports whether this call was the one that did: of two concurrent
// exchanges with the same ID token exactly one wins.
func (a *Accounts) ConsumeGoogleNonce(ctx context.Context, nonce string) (bool, error) {
	if nonce == "" {
		return false, nil
	}

	ph := a.DB.Dialect().Placeholder

	res, err := a.DB.ExecContext(ctx,
		"DELETE FROM google_nonces WHERE nonce_hash = "+ph(1)+" AND expires_at > "+ph(2),
		nonceHash(nonce), a.Clock())
	if err != nil {
		return false, fmt.Errorf("google nonce consume: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("google nonce consume: %w", err)
	}

	return n == 1, nil
}
