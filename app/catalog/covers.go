package catalog

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// Cover holds the two renditions the UI needs: the 96-wide thumbnail and the 600-wide image (SDD 06 §7.5).
type Cover struct {
	Thumb string `json:"thumb"`
	Full  string `json:"full"`
}

// CoverKey is the bucket key of an item cover, which is also its URL path after "/":
// covers/{section}/{slug}.webp, and covers/{section}/{slug}-t.webp for the thumbnail (owner decision 2026-11-01,
// item 18). Covers written before it keep covers/<id>-<hash>[-t].webp until covers:rekey moves them.
func CoverKey(typ, slug string, thumb bool) string {
	suffix := ""
	if thumb {
		suffix = "-t"
	}

	return "covers/" + Sections[typ] + "/" + slug + suffix + ".webp"
}

// ThumbCollides reports whether the thumbnail key of the item (typ, slug) is the full key of another item of the
// same type, slugged slug+"-t". That item wins the key; this thumbnail keeps a content-hashed key.
func ThumbCollides(ctx context.Context, db Querier, typ, slug string) (bool, error) {
	var taken bool

	err := db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM items WHERE type = $1 AND slug = $2)`,
		typ, slug+"-t").Scan(&taken)
	if err != nil {
		return false, fmt.Errorf("cover key collision: %w", err)
	}

	return taken, nil
}

// ResolveCover maps a slug cover path ("{section}/{name}.webp" under /covers) to the stored key of the variant it
// names: the full cover of the item slugged name, else the thumbnail of the item slugged name minus "-t". It
// returns "" when no item of the section has that cover.
func ResolveCover(ctx context.Context, db Querier, typ, name string) (string, error) {
	var key string

	err := db.QueryRowContext(ctx, `SELECT coalesce(cover_key, '') FROM items WHERE type = $1 AND slug = $2`,
		typ, name).Scan(&key)
	if errors.Is(err, sql.ErrNoRows) {
		base, ok := strings.CutSuffix(name, "-t")
		if !ok {
			return "", nil
		}

		err = db.QueryRowContext(ctx, `SELECT coalesce(cover_thumb_key, '') FROM items WHERE type = $1 AND slug = $2`,
			typ, base).Scan(&key)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("resolve cover: %w", err)
	}

	return key, nil
}

// CoverPathFor is the slug URL path of the cover variant of item id ("" when id is no item, or when the thumbnail
// path would name another item's cover): where an old /covers/<id>-<hash>.webp URL moves to.
func CoverPathFor(ctx context.Context, db Querier, id string, thumb bool) (string, error) {
	var typ, slug string

	err := db.QueryRowContext(ctx, `SELECT type::text, slug FROM items WHERE id = $1`, id).Scan(&typ, &slug)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("cover path: %w", err)
	}

	if thumb {
		if taken, err := ThumbCollides(ctx, db, typ, slug); err != nil || taken {
			return "", err
		}
	}

	return "/" + CoverKey(typ, slug, thumb), nil
}

// CoverURL maps a bucket key such as "covers/games/zelda.webp" (or an older "covers/<id>-<hash>.webp") to its same-origin path; the client prepends the shared coversBase prop (COVERS_BASE_URL); "" when the key is empty or unsafe.
func CoverURL(key string) string {
	if key == "" || strings.HasPrefix(key, "/") || strings.Contains(key, "..") {
		return ""
	}

	if !strings.HasPrefix(key, "covers/") {
		key = "covers/" + key
	}

	return "/" + key
}

// CoversBase normalises the COVERS_BASE_URL value: trimmed, no trailing slash.
func CoversBase(raw string) string {
	return strings.TrimRight(strings.TrimSpace(raw), "/")
}

// Absolute prefixes an app-relative cover path with base. JSON clients (the Android app) cannot resolve
// "/covers/..." against the API host when covers live in a bucket, so the JSON handlers send absolute URLs.
func Absolute(base, path string) string {
	if path == "" {
		return ""
	}

	return base + path
}

// WithBase returns a copy of c whose renditions are absolute; nil stays nil and c is never mutated.
func (c *Cover) WithBase(base string) *Cover {
	if c == nil || base == "" {
		return c
	}

	return &Cover{Thumb: Absolute(base, c.Thumb), Full: Absolute(base, c.Full)}
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
