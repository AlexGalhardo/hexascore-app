package seo

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif" // registered for image.Decode
	_ "image/jpeg"
	"image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/AlexGalhardo/vukem/storage"
	"github.com/AlexGalhardo/vukem/uuid"
)

// Open Graph card size (1.91:1, the size the big platforms crop to).
const (
	// CardWidth is the card width in pixels.
	CardWidth = 1200
	// CardHeight is the card height in pixels.
	CardHeight = 630
)

// Panel is token color.primary (light, oklch 0.48 0.196 286.1) in sRGB; the card carries no text
// (SDD 14 §11 item 12), so the title lives only in og:title and og:description.
var Panel = color.RGBA{R: 0x5b, G: 0x3f, B: 0xc4, A: 0xff}

// maxCover caps the cover bytes read from the bucket; covers are ≤ 600×600 WebP (SDD 06 §7.5).
const maxCover = 15 << 20

// Card draws the cover at its own size (never upscaled, clipped if larger) on a solid panel, with equal
// top, bottom and left margins; nil cover gives the panel alone.
func Card(cover image.Image) *image.RGBA {
	card := image.NewRGBA(image.Rect(0, 0, CardWidth, CardHeight))
	draw.Draw(card, card.Bounds(), image.NewUniform(Panel), image.Point{}, draw.Src)

	if cover != nil {
		b := cover.Bounds()
		margin := max((CardHeight-b.Dy())/2, 0)
		at := image.Rect(margin, margin, margin+b.Dx(), margin+b.Dy())
		draw.Draw(card, at, cover, b.Min, draw.Over)
	}

	return card
}

// Cards renders share cards and caches them on Disk under og/; a nil Disk renders every time without
// a cover (the covers live on the bucket).
type Cards struct {
	disk storage.Disk
	// slot keeps one render (and one dwebp process) at a time, like the WebP encodes (NFR-06-2).
	slot chan struct{}
}

// NewCards builds the renderer over disk (nil: no cache, no cover).
func NewCards(disk storage.Disk) *Cards {
	return &Cards{disk: disk, slot: make(chan struct{}, 1)}
}

// PNG returns the card of kind/id whose cover is coverKey ("" for none). The cache key hashes coverKey,
// so a new cover is a new object.
// shortcut: replaced cards stay on the bucket; add them to covers.gc when the og/ prefix grows.
func (c *Cards) PNG(ctx context.Context, kind, id, coverKey string) ([]byte, error) {
	sum := sha256.Sum256([]byte(coverKey))
	key := fmt.Sprintf("og/%s/%s-%s.png", kind, id, hex.EncodeToString(sum[:8]))

	if c.disk != nil {
		if out, err := readAll(ctx, c.disk, key, -1); err == nil {
			return out, nil
		} else if !errors.Is(err, storage.ErrNotFound) {
			return nil, err
		}
	}

	select {
	case c.slot <- struct{}{}:
		defer func() { <-c.slot }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	var cover image.Image
	if c.disk != nil && coverKey != "" {
		raw, err := readAll(ctx, c.disk, coverKey, maxCover)
		if err != nil && !errors.Is(err, storage.ErrNotFound) {
			return nil, err
		}
		cover = decodeCover(ctx, raw)
	}

	var b bytes.Buffer
	if err := png.Encode(&b, Card(cover)); err != nil {
		return nil, fmt.Errorf("encode og card: %w", err)
	}

	if c.disk != nil {
		if err := c.disk.Put(ctx, key, bytes.NewReader(b.Bytes())); err != nil {
			return nil, fmt.Errorf("cache og card %s: %w", key, err)
		}
	}

	return b.Bytes(), nil
}

func readAll(ctx context.Context, d storage.Disk, key string, limit int64) ([]byte, error) {
	rc, err := d.Open(ctx, key)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	var r io.Reader = rc
	if limit > 0 {
		r = io.LimitReader(rc, limit)
	}

	out, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", key, err)
	}

	return out, nil
}

// decodeCover decodes PNG, JPEG or GIF with the standard library and WebP through dwebp (libwebp-tools,
// in the image next to cwebp); anything else, such as the demo SVGs, gives nil and a panel-only card.
func decodeCover(ctx context.Context, raw []byte) image.Image {
	if len(raw) == 0 {
		return nil
	}

	if img, _, err := image.Decode(bytes.NewReader(raw)); err == nil {
		return img
	}

	if len(raw) < 12 || string(raw[:4]) != "RIFF" || string(raw[8:12]) != "WEBP" {
		return nil
	}

	img, err := dwebp(ctx, raw)
	if err != nil {
		return nil
	}

	return img
}

// dwebp converts WebP to PNG with the libwebp CLI: the standard library has no WebP decoder and
// golang.org/x/image would be a new dependency (SDD 14 §11 item 12).
func dwebp(ctx context.Context, raw []byte) (image.Image, error) {
	bin, err := exec.LookPath("dwebp")
	if err != nil {
		return nil, err
	}

	dir, err := os.MkdirTemp("", "og-card-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	in, out := filepath.Join(dir, "in.webp"), filepath.Join(dir, "out.png")
	if err := os.WriteFile(in, raw, 0o600); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// #nosec G204 -- fixed binary from PATH, arguments are paths in our own temp dir.
	if err := exec.CommandContext(ctx, bin, "-quiet", in, "-o", out).Run(); err != nil {
		return nil, fmt.Errorf("dwebp: %w", err)
	}

	f, err := os.Open(out)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return png.Decode(f)
}

// Querier is the read side of *sql.DB the cover lookup needs.
type Querier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// notBanned matches users.banned_at the way the profile and lists do.
const notBanned = `NOT (u.banned_at IS NOT NULL AND (u.ban_expires_at IS NULL OR u.ban_expires_at > now()))`

// CardCover returns the cover key of the share card kind/id and whether an anonymous visitor may see
// the target: items (public), lists (public, not hidden, owner public and not banned; first item's
// cover) and profiles (id is the nickname; public and not banned; latest rating's cover).
func CardCover(ctx context.Context, db Querier, kind, id string) (string, bool, error) {
	var q string

	switch kind {
	case "items", "lists":
		if _, err := uuid.Parse(id); err != nil {
			return "", false, nil //nolint:nilerr // a malformed id is simply not found
		}
		q = `SELECT cover_key FROM items WHERE id = $1 AND visibility = 'public'`
		if kind == "lists" {
			q = `SELECT (SELECT i.cover_key FROM list_items li JOIN items i ON i.id = li.item_id AND i.visibility = 'public'
  WHERE li.list_id = l.id ORDER BY li.position LIMIT 1)
FROM lists l JOIN users u ON u.id = l.user_id
WHERE l.id = $1 AND l.is_public AND NOT l.hidden AND NOT u.profile_private AND u.anonymized_at IS NULL AND ` + notBanned
		}
	case "profiles":
		q = `SELECT (SELECT i.cover_key FROM ratings r JOIN items i ON i.id = r.item_id AND i.visibility = 'public'
  WHERE r.user_id = u.id ORDER BY r.updated_at DESC LIMIT 1)
FROM users u WHERE u.username = lower($1) AND NOT u.profile_private AND u.anonymized_at IS NULL AND ` + notBanned
	default:
		return "", false, nil
	}

	var key sql.NullString
	err := db.QueryRowContext(ctx, q, id).Scan(&key)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("og card %s %s: %w", kind, id, err)
	}

	return key.String, true, nil
}
