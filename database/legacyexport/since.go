// Package legacyexport prints the rollback upserts for hexascore:export-since.
package legacyexport

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Queryer is the read side of a database handle.
type Queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// legacyUser is the Better Auth id: the imported legacy_id, or hs:<numeric id>
// for an account that never existed in legacy. The password hash stays behind;
// argon2id is unreadable by Better Auth (SDD 11 §7.6 R3).
const legacyUser = `coalesce(u.legacy_id, 'hs:' || u.id::text)`

// newAccount is an app user with no legacy id that was created, or wrote, after the cutover.
const newAccount = `u.legacy_id IS NULL AND (
  u.created_at >= $1
  OR EXISTS (SELECT 1 FROM ratings r WHERE r.user_id = u.id AND (r.updated_at >= $1 OR r.created_at >= $1))
  OR EXISTS (SELECT 1 FROM item_status s WHERE s.user_id = u.id AND s.updated_at >= $1)
  OR EXISTS (SELECT 1 FROM time_to_beat t WHERE t.user_id = u.id AND (t.updated_at >= $1 OR t.created_at >= $1))
  OR EXISTS (SELECT 1 FROM lists l WHERE l.user_id = u.id AND (l.updated_at >= $1 OR l.created_at >= $1))
  OR EXISTS (SELECT 1 FROM api_keys k WHERE k.user_id = u.id AND k.created_at >= $1)
)`

// statements are one upsert per row, in replay order. quote_nullable keeps a NULL
// column from erasing the whole line (format() returns NULL if any argument is NULL).
var statements = []string{
	`SELECT concat(
  'INSERT INTO "user" (id, name, email, email_verified, username, created_at, updated_at) VALUES (',
  quote_nullable(` + legacyUser + `), ',', quote_nullable(u.name), ',', quote_nullable(u.email),
  ',false,', quote_nullable(u.username), ',', quote_nullable(u.created_at), ',', quote_nullable(u.updated_at),
  ') ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, username = EXCLUDED.username;'
) FROM users u WHERE ` + newAccount,
	`SELECT concat(
  'INSERT INTO verification (id, identifier, value, expires_at, created_at, updated_at) VALUES (',
  quote_nullable('reset:' || ` + legacyUser + `), ',', quote_nullable('reset-password:' || u.email), ',',
  quote_nullable('required'), ',', quote_nullable(u.created_at + interval '14 days'), ',',
  quote_nullable(u.created_at), ',', quote_nullable(u.updated_at),
  ') ON CONFLICT (id) DO UPDATE SET value = EXCLUDED.value, identifier = EXCLUDED.identifier, expires_at = EXCLUDED.expires_at;'
) FROM users u WHERE ` + newAccount,
	`SELECT concat(
  'INSERT INTO api_keys (id, user_id, name, prefix, hash, created_at) VALUES (',
  quote_nullable(k.id::text), '::uuid,', quote_nullable(` + legacyUser + `), ',', quote_nullable(k.name), ',',
  quote_nullable(k.prefix), ',', quote_nullable(rtrim(k.hash)), ',', quote_nullable(k.created_at),
  ') ON CONFLICT (hash) DO NOTHING;'
) FROM api_keys k JOIN users u ON u.id = k.user_id WHERE k.created_at >= $1`,
	`SELECT concat(
  'INSERT INTO ratings (user_id, item_id, c1, c2, c3, c4, c5, c6, created_at, updated_at) VALUES (',
  quote_nullable(` + legacyUser + `), ',', quote_nullable(r.item_id::text), '::uuid,',
  r.c1::text, ',', r.c2::text, ',', r.c3::text, ',', r.c4::text, ',', r.c5::text, ',', r.c6::text, ',',
  quote_nullable(r.created_at), ',', quote_nullable(r.updated_at),
  ') ON CONFLICT (user_id, item_id) DO UPDATE SET c1 = EXCLUDED.c1, c2 = EXCLUDED.c2, c3 = EXCLUDED.c3, c4 = EXCLUDED.c4, c5 = EXCLUDED.c5, c6 = EXCLUDED.c6, updated_at = EXCLUDED.updated_at;'
) FROM ratings r JOIN users u ON u.id = r.user_id WHERE r.updated_at >= $1 OR r.created_at >= $1`,
	`SELECT concat(
  'INSERT INTO item_status (user_id, item_id, status, updated_at) VALUES (',
  quote_nullable(` + legacyUser + `), ',', quote_nullable(s.item_id::text), '::uuid,',
  quote_nullable(s.status::text), '::item_status_value,', quote_nullable(s.updated_at),
  ') ON CONFLICT (user_id, item_id) DO UPDATE SET status = EXCLUDED.status, updated_at = EXCLUDED.updated_at;'
) FROM item_status s JOIN users u ON u.id = s.user_id WHERE s.updated_at >= $1`,
	`SELECT concat(
  'INSERT INTO time_to_beat (user_id, item_id, main_minutes, extras_minutes, completionist_minutes, created_at, updated_at) VALUES (',
  quote_nullable(` + legacyUser + `), ',', quote_nullable(t.item_id::text), '::uuid,',
  quote_nullable(t.main_minutes), ',', quote_nullable(t.extras_minutes), ',', quote_nullable(t.completionist_minutes), ',',
  quote_nullable(t.created_at), ',', quote_nullable(t.updated_at),
  ') ON CONFLICT (user_id, item_id) DO UPDATE SET main_minutes = EXCLUDED.main_minutes, extras_minutes = EXCLUDED.extras_minutes, completionist_minutes = EXCLUDED.completionist_minutes, updated_at = EXCLUDED.updated_at;'
) FROM time_to_beat t JOIN users u ON u.id = t.user_id WHERE t.updated_at >= $1 OR t.created_at >= $1`,
	`SELECT concat(
  'INSERT INTO lists (id, user_id, type, title, description, is_public, hidden, created_at, updated_at) VALUES (',
  quote_nullable(l.id::text), '::uuid,', quote_nullable(` + legacyUser + `), ',',
  quote_nullable(l.type), '::item_type,', quote_nullable(l.title), ',', quote_nullable(l.description), ',',
  CASE WHEN l.is_public THEN 'true' ELSE 'false' END, ',', CASE WHEN l.hidden THEN 'true' ELSE 'false' END, ',',
  quote_nullable(l.created_at), ',', quote_nullable(l.updated_at),
  ') ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, description = EXCLUDED.description, is_public = EXCLUDED.is_public, hidden = EXCLUDED.hidden, updated_at = EXCLUDED.updated_at;'
) FROM lists l JOIN users u ON u.id = l.user_id
WHERE l.updated_at >= $1 OR l.created_at >= $1
   OR EXISTS (SELECT 1 FROM list_items li WHERE li.list_id = l.id AND li.added_at >= $1)`,
	`SELECT concat(
  'INSERT INTO list_items (list_id, item_id, position, note) VALUES (',
  quote_nullable(li.list_id::text), '::uuid,', quote_nullable(li.item_id::text), '::uuid,',
  li.position::text, ',', quote_nullable(li.note),
  ') ON CONFLICT (list_id, item_id) DO UPDATE SET position = EXCLUDED.position, note = EXCLUDED.note;'
) FROM list_items li JOIN lists l ON l.id = li.list_id
WHERE l.updated_at >= $1 OR l.created_at >= $1 OR li.added_at >= $1`,
}

// Statements returns the legacy upserts for rows written at or after since.
// Ratings are copied as stored (0–100, multiples of 5 stay as they are).
func Statements(ctx context.Context, q Queryer, since time.Time) ([]string, error) {
	if since.IsZero() {
		return nil, errors.New("export-since: cutover time is required")
	}

	var out []string

	for i, query := range statements {
		lines, err := section(ctx, q, query, since.UTC())
		if err != nil {
			return nil, fmt.Errorf("export-since: section %d: %w", i, err)
		}

		out = append(out, lines...)
	}

	return out, nil
}

func section(ctx context.Context, q Queryer, query string, since time.Time) ([]string, error) {
	rows, err := q.QueryContext(ctx, query, since)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var out []string

	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			return nil, err
		}

		if line != "" {
			out = append(out, line)
		}
	}

	return out, rows.Err()
}
