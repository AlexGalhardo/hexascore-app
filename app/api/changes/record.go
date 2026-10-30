// Package changes writes the public API change feed inside the cause's transaction.
package changes

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
)

// Change is one api_changes row. Fields are API names; names outside the public
// allowlist are dropped, and an update with nothing left is not written.
type Change struct {
	ItemID     string
	ItemType   string
	Action     string
	Fields     []string
	MergedInto string
}

// Exec is the transaction the cause already holds.
type Exec interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// ScoreSnap is the stored aggregate the API shows as one decimal, a confidence and a count.
type ScoreSnap struct {
	Score       sql.NullFloat64
	Confidence  sql.NullString
	RatingCount int
}

var publicFields = map[string]struct{}{
	"title": {}, "synopsis": {}, "slug": {}, "wikidata": {}, "releaseDate": {},
	"runtimeMinutes": {}, "pageCount": {}, "externalIds": {}, "entities": {},
	"score": {}, "confidence": {}, "ratingCount": {},
}

// Visible keeps the public field names, in the caller's order, without duplicates.
func Visible(fields []string) []string {
	out := make([]string, 0, len(fields))
	seen := map[string]struct{}{}
	for _, f := range fields {
		if _, ok := publicFields[f]; !ok {
			continue
		}
		if _, ok := seen[f]; ok {
			continue
		}
		seen[f] = struct{}{}
		out = append(out, f)
	}
	return out
}

// Record inserts c. An update whose fields are all hidden writes nothing.
func Record(ctx context.Context, ex Exec, c Change) error {
	fields := Visible(c.Fields)
	if c.Action == "updated" && len(fields) == 0 {
		return nil
	}
	if c.Action != "created" && c.Action != "updated" && c.Action != "deleted" {
		return fmt.Errorf("api changes: action %q", c.Action)
	}

	var merged any
	if c.MergedInto != "" {
		merged = c.MergedInto
	}
	_, err := ex.ExecContext(ctx, `INSERT INTO api_changes (item_id, item_type, action, fields, merged_into)
VALUES ($1, $2, $3, $4::text[], $5)`, c.ItemID, c.ItemType, c.Action, pgTextArray(fields), merged)
	if err != nil {
		return fmt.Errorf("api changes: record %s %s: %w", c.Action, c.ItemID, err)
	}
	return nil
}

// ScoreFieldNames returns the public names whose one-decimal score, confidence or count changed.
func ScoreFieldNames(before, after ScoreSnap) []string {
	var fields []string
	bScore, bOK := tenth(before.Score)
	aScore, aOK := tenth(after.Score)
	if bOK != aOK || bScore != aScore {
		fields = append(fields, "score")
	}
	if before.Confidence != after.Confidence {
		fields = append(fields, "confidence")
	}
	if before.RatingCount != after.RatingCount {
		fields = append(fields, "ratingCount")
	}
	return fields
}

// LoadScore reads the stored aggregate of itemID.
func LoadScore(ctx context.Context, ex Exec, itemID string) (ScoreSnap, error) {
	var s ScoreSnap
	err := ex.QueryRowContext(ctx, `SELECT score, score_confidence, rating_count FROM items WHERE id = $1`, itemID).
		Scan(&s.Score, &s.Confidence, &s.RatingCount)
	if err != nil {
		return ScoreSnap{}, fmt.Errorf("api changes: load score %s: %w", itemID, err)
	}
	return s, nil
}

// tenth is the one-decimal star value of a 0–100 stored score. A null score has no tenth.
func tenth(v sql.NullFloat64) (float64, bool) {
	if !v.Valid {
		return 0, false
	}
	return math.Round(v.Float64) / 10, true
}

func pgTextArray(values []string) string {
	if len(values) == 0 {
		return "{}"
	}
	return "{" + strings.Join(values, ",") + "}"
}
