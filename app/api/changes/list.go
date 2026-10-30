package changes

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// Window is the oldest since listChanges accepts. Exactly Window ago is allowed.
const Window = 14 * 24 * time.Hour

// Retain is how long an api_changes row is kept. Prune deletes rows older than this.
const Retain = 30 * 24 * time.Hour

// DefaultLimit and MaxLimit bound listChanges. The cap is the feed's own 1000, not the plan page size.
const (
	DefaultLimit = 100
	MaxLimit     = 1000
)

// Querier is the read side of the database.
type Querier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// After is the last row of the previous page, compared as (changed_at, id).
type After struct {
	At time.Time
	ID int64
}

// Filter is one listChanges page. ItemType is the singular catalogue type, or empty for every type.
type Filter struct {
	Since    time.Time
	ItemType string
	After    *After
	Limit    int
}

// Row is one stored change.
type Row struct {
	ID         int64
	ItemID     string
	ItemType   string
	Action     string
	Fields     []string
	MergedInto string
	ChangedAt  time.Time
}

// List returns at most f.Limit rows at or after Since, ordered by (changed_at, id).
// more is true when another row follows the page. The keyset is signed by the caller:
// vukem's cursor pages by primary key only.
func List(ctx context.Context, q Querier, f Filter) ([]Row, bool, error) {
	if f.Limit < 1 {
		return nil, false, fmt.Errorf("api changes: limit %d", f.Limit)
	}

	var afterAt any
	var afterID any
	if f.After != nil {
		afterAt = f.After.At.UTC()
		afterID = f.After.ID
	}

	rows, err := q.QueryContext(ctx, `
SELECT id, item_id::text, item_type, action,
       coalesce(array_to_string(fields, ','), ''),
       coalesce(merged_into::text, ''),
       changed_at
FROM api_changes
WHERE changed_at >= $1
  AND ($2::text = '' OR item_type = $2)
  AND ($3::timestamptz IS NULL OR (changed_at, id) > ($3::timestamptz, $4::bigint))
ORDER BY changed_at, id
LIMIT $5`, f.Since.UTC(), f.ItemType, afterAt, afterID, f.Limit+1)
	if err != nil {
		return nil, false, fmt.Errorf("api changes: list: %w", err)
	}
	defer rows.Close()

	out := make([]Row, 0, f.Limit)
	for rows.Next() {
		var row Row
		var fields string
		if err := rows.Scan(&row.ID, &row.ItemID, &row.ItemType, &row.Action, &fields, &row.MergedInto, &row.ChangedAt); err != nil {
			return nil, false, fmt.Errorf("api changes: scan: %w", err)
		}
		row.Fields = splitFields(fields)
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("api changes: list: %w", err)
	}

	if len(out) > f.Limit {
		return out[:f.Limit], true, nil
	}

	return out, false, nil
}

func splitFields(raw string) []string {
	if raw == "" {
		return []string{}
	}

	return strings.Split(raw, ",")
}
