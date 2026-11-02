package covers

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/AlexGalhardo/vukem/database"
)

// HashRow is one cover_hashes row.
type HashRow struct {
	Key           string
	SHA256        [sha256.Size]byte
	DHash         uint64
	Width, Height int
	Bytes         int
	Source        string
	Licence       string
}

// Hashes is the cover_hashes store the job consults to skip a picture it already holds.
type Hashes interface {
	// Same reports whether key holds a picture within NearDuplicate bits of dhash.
	Same(ctx context.Context, key string, dhash uint64) (bool, error)
	// Put records the picture stored under row.Key.
	Put(ctx context.Context, row HashRow) error
}

// PGHashes is Hashes over Postgres. Give it a handle outside the step transaction (like Reserver): a rolled-back
// step may leave a row behind, which is harmless because Same compares against the key an item actually points at.
type PGHashes struct{ DB database.Executor }

// Same implements Hashes.
func (p PGHashes) Same(ctx context.Context, key string, dhash uint64) (bool, error) {
	var d int64

	// The unsigned hash is stored as the bigint with the same bits.
	err := p.DB.QueryRowContext(ctx, `SELECT dhash FROM cover_hashes WHERE key = $1`, key).Scan(&d)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("covers: read hash of %s: %w", key, err)
	}

	return Distance(uint64(d), dhash) <= NearDuplicate, nil //nolint:gosec // same 64 bits
}

// Put implements Hashes.
func (p PGHashes) Put(ctx context.Context, r HashRow) error {
	_, err := p.DB.ExecContext(
		ctx,
		`INSERT INTO cover_hashes (key, sha256, dhash, width, height, bytes, source, licence)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (key) DO UPDATE SET sha256 = EXCLUDED.sha256, dhash = EXCLUDED.dhash, width = EXCLUDED.width,
  height = EXCLUDED.height, bytes = EXCLUDED.bytes, source = EXCLUDED.source, licence = EXCLUDED.licence,
  created_at = now()`,
		r.Key,
		r.SHA256[:],
		int64(r.DHash), //nolint:gosec // same 64 bits
		r.Width,
		r.Height,
		r.Bytes,
		r.Source,
		r.Licence,
	)
	if err != nil {
		return fmt.Errorf("covers: write hash of %s: %w", r.Key, err)
	}

	return nil
}

// Rights answers whether a data source may have its images displayed, with the licence it recorded.
type Rights interface {
	Allowed(ctx context.Context, source string) (licence string, ok bool, err error)
}

// noImageLicences are the data_sources.licence values that grant no right to display a source's images. A source
// without a data_sources row is refused too (fail closed).
var noImageLicences = map[string]bool{"": true, "pending": true, "none": true, "unknown": true, "no-images": true}

// PGRights is Rights over data_sources.
type PGRights struct{ DB database.Executor }

// Allowed implements Rights.
func (p PGRights) Allowed(ctx context.Context, source string) (string, bool, error) {
	var licence string

	err := p.DB.QueryRowContext(ctx, `SELECT licence FROM data_sources WHERE code = $1`, source).Scan(&licence)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}

	if err != nil {
		return "", false, fmt.Errorf("covers: licence of %s: %w", source, err)
	}

	return licence, !noImageLicences[strings.ToLower(strings.TrimSpace(licence))], nil
}
