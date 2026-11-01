package community

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/uuid"
)

// RecentPageSize is the Recent changes page size (FR-04-13).
const RecentPageSize = 20

// ErrTargetMissing is a Recent changes page for a record that does not exist or is not public.
var ErrTargetMissing = errors.New("community: recent changes target not found")

// RecentTarget is the record a Recent changes page belongs to.
type RecentTarget struct {
	Type  string `json:"type"`
	ID    string `json:"id"`
	Title string `json:"title"`
}

// RecentEntry is one public audit row. Actor is "user" (with Nickname), "team" (admins: "HexaScore team"),
// "deleted" (actor id nulled by account deletion, NFR-04-5) or "system"; staff never get a nickname.
type RecentEntry struct {
	ID         int64     `json:"id"`
	OccurredAt time.Time `json:"occurredAt"`
	Source     string    `json:"source"`
	Field      string    `json:"field"`
	Before     string    `json:"before"`
	After      string    `json:"after"`
	Actor      string    `json:"actor"`
	Nickname   string    `json:"nickname,omitempty"`
}

// ResolveItem finds a public item of typ by slug.
func ResolveItem(ctx context.Context, db *database.DB, typ, slug string) (RecentTarget, error) {
	return resolve(ctx, db, "item", `SELECT id::text, original_title FROM items
WHERE type = $1 AND slug = $2 AND visibility = 'public'`, typ, slug)
}

// ResolveEntity finds an entity of kind by slug (or by code for countries and languages, byCode).
func ResolveEntity(ctx context.Context, db *database.DB, kind, key string, byCode bool) (RecentTarget, error) {
	match := "slug = $2"
	if byCode {
		match = "lower(code) = lower($2)"
	}

	return resolve(ctx, db, "entity", `SELECT id::text, default_name FROM entities WHERE kind = $1 AND `+match,
		kind, key)
}

// ResolveEdition finds an edition id of the public book slug.
func ResolveEdition(ctx context.Context, db *database.DB, slug, id string) (RecentTarget, error) {
	if _, err := uuid.Parse(id); err != nil {
		return RecentTarget{}, ErrTargetMissing
	}

	return resolve(ctx, db, "edition", `SELECT e.id::text, COALESCE(e.title, i.original_title) FROM editions e
JOIN items i ON i.id = e.work_id WHERE i.type = 'book' AND i.slug = $1 AND i.visibility = 'public' AND e.id = $2`,
		slug, id)
}

func resolve(ctx context.Context, db *database.DB, typ, q string, args ...any) (RecentTarget, error) {
	t := RecentTarget{Type: typ}

	var title sql.NullString

	err := db.QueryRowContext(ctx, q, args...).Scan(&t.ID, &title)
	if errors.Is(err, sql.ErrNoRows) {
		return RecentTarget{}, ErrTargetMissing
	}

	if err != nil {
		return RecentTarget{}, fmt.Errorf("community: resolve %s: %w", typ, err)
	}

	t.Title = title.String

	return t, nil
}

// RecentChanges returns page (1-based) of the target's audit rows, newest first, and whether a next page exists.
// Merge rows keep their field but not their snapshots, which hold whole records.
func RecentChanges(ctx context.Context, db *database.DB, t RecentTarget, page int) ([]RecentEntry, bool, error) {
	page = max(page, 1)

	var raw []byte

	err := db.QueryRowContext(ctx, `SELECT COALESCE(json_agg(row_to_json(r) ORDER BY r.id DESC), '[]') FROM (
SELECT a.id, a.occurred_at AS "occurredAt", a.source::text AS source, a.field,
  CASE WHEN a.source = 'merge' THEN '' ELSE COALESCE(a.before::text, 'null') END AS before,
  CASE WHEN a.source = 'merge' THEN '' ELSE COALESCE(a.after::text, 'null') END AS after,
  CASE WHEN a.actor_role IN ('admin','moderator') THEN 'team'
       WHEN a.actor_role = 'system' THEN 'system'
       WHEN u.id IS NULL OR u.anonymized_at IS NOT NULL THEN 'deleted'
       ELSE 'user' END AS actor,
  CASE WHEN a.actor_role = 'author' AND u.anonymized_at IS NULL THEN COALESCE(u.display_username, u.username) END
    AS nickname
FROM catalog_audit a LEFT JOIN users u ON u.id = a.actor_id
WHERE a.target_type = $1 AND a.target_id = $2::uuid
ORDER BY a.id DESC LIMIT `+strconv.Itoa(RecentPageSize+1)+` OFFSET $3) r`,
		t.Type, t.ID, (page-1)*RecentPageSize).Scan(&raw)
	if err != nil {
		return nil, false, fmt.Errorf("community: recent changes: %w", err)
	}

	var entries []RecentEntry
	if err := json.Unmarshal(raw, &entries); err != nil {
		return nil, false, fmt.Errorf("community: decode recent changes: %w", err)
	}

	more := len(entries) > RecentPageSize
	if more {
		entries = entries[:RecentPageSize]
	}

	return entries, more, nil
}
