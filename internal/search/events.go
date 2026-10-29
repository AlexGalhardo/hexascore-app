package search

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"sync/atomic"
	"time"

	"github.com/AlexGalhardo/vukem/database"
)

// EventRetention is how long search_events rows live (SDD 07 §7.8).
const EventRetention = 30 * 24 * time.Hour

// Redacted replaces e-mail, phone and CPF patterns in stored queries.
const Redacted = "[redacted]"

var (
	emailPattern = regexp.MustCompile(`[^\s@]+@[^\s@]+\.[^\s@]+`)
	cpfPattern   = regexp.MustCompile(`\b\d{3}\.?\d{3}\.?\d{3}-?\d{2}\b`)
	// Seven or more digits, optionally separated by spaces, dots, dashes or parentheses.
	phonePattern = regexp.MustCompile(`\+?\(?\d[\d\s().-]{5,}\d`)
)

// Redact hides personal data a visitor typed into the search box before the
// query is stored; the order matters because a CPF also matches the phone
// pattern.
func Redact(text string) string {
	text = emailPattern.ReplaceAllString(text, Redacted)
	text = cpfPattern.ReplaceAllString(text, Redacted)

	return phonePattern.ReplaceAllString(text, Redacted)
}

// epoch changes whenever the suggest sync applied changes, so cached search
// pages from before it are never served (SDD 07 §7.6).
var epoch atomic.Int64

// Epoch is the current cache epoch.
func Epoch() int64 { return epoch.Load() }

// BumpEpoch invalidates the cached search pages and suggestions.
func BumpEpoch() { epoch.Add(1) }

// Performed is one executed search.
type Performed struct {
	SearchID    string
	Surface     string
	Text        string
	UserID      *int64
	Locale      string
	Types       []string
	ResultCount int
	Latency     time.Duration
	Page        int
}

// RecordPerformed stores a `search.performed` event. The query is redacted
// and normalised by hs_norm in SQL (Go never normalises, SDD 07 §5).
func RecordPerformed(ctx context.Context, db *database.DB, e Performed) error {
	_, err := db.ExecContext(ctx, `INSERT INTO search_events
		(search_id, surface, query_norm, user_id, locale, types, result_count, latency_ms, page)
		VALUES ($1::uuid, $2, hs_norm($3), $4, $5, coalesce($6::text[], '{}'), $7, $8, $9)`,
		e.SearchID, e.Surface, Redact(e.Text), e.UserID, e.Locale, textArray(e.Types), e.ResultCount,
		int(e.Latency.Milliseconds()), e.Page)
	if err != nil {
		return fmt.Errorf("record search event: %w", err)
	}

	return nil
}

// Click is a result click reported by the page.
type Click struct {
	SearchID string
	RefID    string
	Position int
	Group    string
}

// ResultClicked is the in-process event SDD 08 consumes by this name
// (open issue 08-5). It never carries the query text or a user id.
type ResultClicked struct {
	ItemID    string
	QueryHash string
	Position  int
	Locale    string
	At        time.Time
}

// RecordClick stores a `search.clicked` row copied from the search it belongs
// to and returns the event to dispatch; ok is false when the search is
// unknown (expired, forged id), in which case nothing is stored.
func RecordClick(ctx context.Context, db *database.DB, c Click, now time.Time) (ev ResultClicked, ok bool, err error) {
	var queryNorm, locale string

	err = db.QueryRowContext(ctx, `SELECT query_norm, locale FROM search_events
		WHERE search_id = $1::uuid AND clicked_ref IS NULL ORDER BY at LIMIT 1`, c.SearchID).Scan(&queryNorm, &locale)
	if errors.Is(err, sql.ErrNoRows) {
		return ev, false, nil
	}

	if err != nil {
		return ev, false, fmt.Errorf("find search for click: %w", err)
	}

	if _, err = db.ExecContext(ctx, `INSERT INTO search_events
		(at, search_id, surface, query_norm, locale, clicked_ref, click_position, click_group)
		VALUES ($1, $2::uuid, 'web', $3, $4, $5::uuid, $6, $7)`,
		now, c.SearchID, queryNorm, locale, c.RefID, c.Position, c.Group); err != nil {
		return ev, false, fmt.Errorf("record search click: %w", err)
	}

	sum := sha256.Sum256([]byte(queryNorm + "|" + locale))

	return ResultClicked{
		ItemID: c.RefID, QueryHash: hex.EncodeToString(sum[:]), Position: c.Position, Locale: locale, At: now,
	}, true, nil
}

// PurgeEvents deletes events older than EventRetention and returns how many.
func PurgeEvents(ctx context.Context, db *database.DB, now time.Time) (int64, error) {
	res, err := db.ExecContext(ctx, `DELETE FROM search_events WHERE at < $1`, now.Add(-EventRetention))
	if err != nil {
		return 0, fmt.Errorf("purge search events: %w", err)
	}

	return res.RowsAffected()
}
