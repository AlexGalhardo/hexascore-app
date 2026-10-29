// Package api holds the public /api/v1 services: API keys and their guard (SDD 05).
package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/uuid"
)

const (
	// KeyPrefix starts every live key (legacy lib/api/keys.ts).
	KeyPrefix = "hx_live_"
	// keyBodyLen is the number of base62 characters after KeyPrefix.
	keyBodyLen = 32
	// DisplayPrefixLen is how much of a key is stored and shown in clear: KeyPrefix plus 6 characters.
	DisplayPrefixLen = len(KeyPrefix) + 6
	// MaxLiveKeys is the cap of unrevoked keys per user (SDD 05 FR-05-2).
	MaxLiveKeys = 10
	// MaxKeyName bounds a key name in characters.
	MaxKeyName = 40

	alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

var (
	// ErrTooManyKeys is returned by Create when the user already holds MaxLiveKeys live keys.
	ErrTooManyKeys = errors.New("api: too many live keys")
	// ErrKeyNotFound is returned by Revoke for a key that is not the user's or is already revoked.
	ErrKeyNotFound = errors.New("api: key not found")
	// ErrUnknownKey is returned by Resolve for a malformed, unknown or revoked key.
	ErrUnknownKey = errors.New("api: unknown key")
)

// Key is a stored API key; the secret itself is never stored.
type Key struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Prefix     string     `json:"prefix"`
	LastUsedAt *time.Time `json:"lastUsedAt"`
	RevokedAt  *time.Time `json:"revokedAt"`
	CreatedAt  time.Time  `json:"createdAt"`
}

// Keys stores and resolves API keys in api_keys.
type Keys struct {
	DB  *database.DB
	Now func() time.Time
}

// NewKey returns a fresh hx_live_ key of 32 uniformly random base62 characters (crypto/rand). The format matches
// the legacy generator; the legacy byte-modulo bias is not kept.
func NewKey() (string, error) {
	body := make([]byte, keyBodyLen)
	limit := big.NewInt(int64(len(alphabet)))

	for i := range body {
		n, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return "", fmt.Errorf("generate api key: %w", err)
		}

		body[i] = alphabet[n.Int64()]
	}

	return KeyPrefix + string(body), nil
}

// HashKey is the stored form of a key: lowercase SHA-256 hex of the full key, the legacy format (A-04).
func HashKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}

// WellFormed reports whether key has the hx_live_ shape, so malformed input never reaches the database.
func WellFormed(key string) bool {
	if len(key) != len(KeyPrefix)+keyBodyLen || key[:len(KeyPrefix)] != KeyPrefix {
		return false
	}

	for i := len(KeyPrefix); i < len(key); i++ {
		c := key[i]
		if (c < '0' || c > '9') && (c < 'A' || c > 'Z') && (c < 'a' || c > 'z') {
			return false
		}
	}

	return true
}

// Create stores a new key named name for userID and returns the plain key, which is never retrievable again.
// The user row is locked so concurrent creates cannot pass the MaxLiveKeys cap together.
func (s *Keys) Create(ctx context.Context, userID int64, name string) (string, Key, error) {
	plain, err := NewKey()
	if err != nil {
		return "", Key{}, err
	}

	k := Key{ID: uuid.NewV7().String(), Name: name, Prefix: plain[:DisplayPrefixLen], CreatedAt: s.Now().UTC()}

	err = s.DB.Tx(ctx, func(tx *database.Tx) error {
		var live int
		if err := tx.QueryRowContext(ctx, `SELECT 1 FROM users WHERE id = $1 FOR UPDATE`, userID).
			Scan(&live); err != nil {
			return fmt.Errorf("lock user %d: %w", userID, err)
		}

		if err := tx.QueryRowContext(ctx,
			`SELECT count(*) FROM api_keys WHERE user_id = $1 AND revoked_at IS NULL`, userID).Scan(&live); err != nil {
			return err
		}

		if live >= MaxLiveKeys {
			return ErrTooManyKeys
		}

		_, err := tx.ExecContext(ctx,
			`INSERT INTO api_keys (id, user_id, name, prefix, hash, created_at) VALUES ($1, $2, $3, $4, $5, $6)`,
			k.ID, userID, k.Name, k.Prefix, HashKey(plain), k.CreatedAt)

		return err
	})
	if err != nil {
		return "", Key{}, err
	}

	return plain, k, nil
}

// List returns every key of userID, revoked ones included, newest first.
func (s *Keys) List(ctx context.Context, userID int64) ([]Key, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT id, name, prefix, last_used_at, revoked_at, created_at FROM api_keys
WHERE user_id = $1 ORDER BY created_at DESC, id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Key{}

	for rows.Next() {
		var k Key
		if err := rows.Scan(&k.ID, &k.Name, &k.Prefix, &k.LastUsedAt, &k.RevokedAt, &k.CreatedAt); err != nil {
			return nil, err
		}

		out = append(out, k)
	}

	return out, rows.Err()
}

// Revoke revokes userID's live key id; a key of another user, an already revoked or a malformed id is
// ErrKeyNotFound. Resolve reads the database on every call, so the next request with the key gets 401.
func (s *Keys) Revoke(ctx context.Context, userID int64, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return ErrKeyNotFound
	}

	res, err := s.DB.ExecContext(ctx,
		`UPDATE api_keys SET revoked_at = $3 WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL`,
		id, userID, s.Now().UTC())
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n == 0 {
		return ErrKeyNotFound
	}

	return nil
}

// Resolve returns the key id and owner of a live key, ErrUnknownKey for a malformed, unknown or revoked one, and
// any other error when the lookup could not run (the guard fails closed on it).
// shortcut: one indexed query per request, no in-process LRU (SDD 05 §7.1); add it with plan resolution in P1.H.1.2.
func (s *Keys) Resolve(ctx context.Context, key string) (keyID string, userID int64, err error) {
	if !WellFormed(key) {
		return "", 0, ErrUnknownKey
	}

	err = s.DB.QueryRowContext(ctx, `SELECT id, user_id FROM api_keys WHERE hash = $1 AND revoked_at IS NULL`,
		HashKey(key)).Scan(&keyID, &userID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", 0, ErrUnknownKey
	}

	return keyID, userID, err
}
