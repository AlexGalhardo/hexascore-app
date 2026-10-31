package catalog

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync/atomic"
)

// coversBase prefixes every cover key. Empty serves keys from the app origin ("/covers/<key>"); production
// sets COVERS_BASE_URL to the public bucket or CDN origin (SDD 09 §7.8 allows that host in img-src).
var coversBase atomic.Pointer[string]

// SetCoversBase sets the origin that cover keys are served from; call it once at boot.
func SetCoversBase(base string) {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	coversBase.Store(&base)
}

// Cover holds the two renditions the UI needs: the 96-wide thumbnail and the 600-wide image (SDD 06 §7.5).
type Cover struct {
	Thumb string `json:"thumb"`
	Full  string `json:"full"`
}

// CoverURL maps a bucket key such as "covers/<id>-<hash>.webp" to its public URL; "" when the key is empty or unsafe.
func CoverURL(key string) string {
	if key == "" || strings.HasPrefix(key, "/") || strings.Contains(key, "..") {
		return ""
	}

	var base string
	if p := coversBase.Load(); p != nil {
		base = *p
	}
	if !strings.HasPrefix(key, "covers/") {
		key = "covers/" + key
	}

	return base + "/" + key
}

// NewCover builds the renditions from the nullable key columns; nil means "show the placeholder".
func NewCover(fullKey, thumbKey sql.NullString) *Cover {
	thumb, full := CoverURL(thumbKey.String), CoverURL(fullKey.String)
	switch {
	case thumb == "" && full == "":
		return nil
	case thumb == "":
		thumb = full
	case full == "":
		full = thumb
	}

	return &Cover{Thumb: thumb, Full: full}
}

// CoverFromThumbKey rebuilds both renditions from the thumbnail key alone (library, list and profile rows):
// the full image sits at the same key without the "-t" suffix before the extension (SDD 06 §7.5).
func CoverFromThumbKey(thumbKey *string) *Cover {
	if thumbKey == nil {
		return nil
	}

	full := strings.Replace(*thumbKey, "-t.", ".", 1)

	return NewCover(sql.NullString{String: full, Valid: full != ""}, sql.NullString{String: *thumbKey, Valid: true})
}

// CoversFor loads the covers of the given item ids; items without a cover are absent from the map.
func CoversFor(ctx context.Context, db Querier, ids []string) (map[string]*Cover, error) {
	out := map[string]*Cover{}
	if len(ids) == 0 {
		return out, nil
	}

	rows, err := db.QueryContext(ctx,
		`SELECT id::text, cover_key, cover_thumb_key FROM items WHERE id = ANY($1::uuid[])`,
		"{"+strings.Join(ids, ",")+"}")
	if err != nil {
		return nil, fmt.Errorf("load covers: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var full, thumb sql.NullString
		if err := rows.Scan(&id, &full, &thumb); err != nil {
			return nil, fmt.Errorf("scan cover: %w", err)
		}
		if c := NewCover(full, thumb); c != nil {
			out[id] = c
		}
	}

	return out, rows.Err()
}
