package llm

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"
)

// ensureSource makes sure the data source the model's values point at exists (field_reviews and provenance reference
// data_sources).
func ensureSource(ctx context.Context, tx *database.Tx) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO data_sources (code, name, licence, url)
VALUES ('llm', 'Claude (language model)', 'generated', 'https://www.anthropic.com') ON CONFLICT (code) DO NOTHING`)
	if err != nil {
		return fmt.Errorf("llm: data source: %w", err)
	}

	return nil
}

var months = []string{
	"january", "february", "march", "april", "may", "june",
	"july", "august", "september", "october", "november", "december",
}

func hasNumber(text string, n int) bool {
	for _, form := range []string{strconv.Itoa(n), groupThousands(n)} {
		if regexp.MustCompile(`(^|[^0-9])` + regexp.QuoteMeta(form) + `([^0-9]|$)`).MatchString(text) {
			return true
		}
	}

	return false
}

func groupThousands(n int) string {
	s := strconv.Itoa(n)
	if len(s) <= 3 {
		return s
	}

	var b strings.Builder

	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}

		b.WriteRune(r)
	}

	return b.String()
}

// Grounded reports whether a value the model gave is written in the text it read: the number itself, or the year,
// the English month name and the day of a date. Only a grounded value is written to the item directly; the rest goes
// to the review queue (owner 2026-11-04: "direct only with high confidence").
func Grounded(text, field string, value any) bool {
	lower := strings.ToLower(text)

	switch field {
	case "runtime_minutes", "page_count":
		n, ok := value.(int)

		return ok && hasNumber(lower, n)
	case "release_date":
		s, ok := value.(string)
		if !ok {
			return false
		}

		d, err := time.Parse(time.DateOnly, s)
		if err != nil {
			return false
		}

		month := months[d.Month()-1]

		return hasNumber(lower, d.Year()) && hasNumber(lower, d.Day()) &&
			(strings.Contains(lower, month) || strings.Contains(lower, month[:3]+" ") || strings.Contains(lower, s))
	}

	return false
}

// applyDirect writes a grounded value to the item when the item still lacks it, and names the model in the
// provenance. It reports whether a row changed.
func applyDirect(ctx context.Context, tx *database.Tx, itemID, field string, value any) (bool, error) {
	var q string

	switch field {
	case "release_date":
		q = `UPDATE items SET release_date = $2::date, release_precision = 'day',
  provenance = provenance || jsonb_build_object('release_date', 'llm'), updated_at = now()
WHERE id = $1 AND release_date IS NULL`
	case "runtime_minutes":
		q = `UPDATE items SET runtime_minutes = $2::int,
  provenance = provenance || jsonb_build_object('runtime_minutes', 'llm'), updated_at = now()
WHERE id = $1 AND runtime_minutes IS NULL`
	case "page_count":
		q = `UPDATE items SET page_count = $2::int,
  provenance = provenance || jsonb_build_object('page_count', 'llm'), updated_at = now()
WHERE id = $1 AND page_count IS NULL`
	default:
		return false, fmt.Errorf("llm: no direct write for %s", field)
	}

	res, err := tx.ExecContext(ctx, q, itemID, value)
	if err != nil {
		return false, fmt.Errorf("llm: write %s: %w", field, err)
	}

	n, _ := res.RowsAffected()

	return n > 0, nil
}
