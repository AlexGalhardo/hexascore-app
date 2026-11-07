package trending

import (
	"cmp"
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/AlexGalhardo/vukem/database"
)

// Types are the catalogue types that each get a top list.
var Types = []string{"game", "movie", "series", "book"}

// WindowDays is how many UTC days of activity feed T_i ("Top 10 trending (30 days)").
const WindowDays = 30

// ListSize is the number of ranks per type.
const ListSize = 10

// Candidate is one public, released item with its activity already scored.
type Candidate struct {
	ID               string
	Trend, Actors    float64
	Prior            float64
	DaysSinceRelease float64
}

// Entry is one snapshot row.
type Entry struct {
	Rank  int
	ID    string
	Score float64
	Basis string
}

// Pick takes the eligible candidates by (T desc, popularity_prior desc, id) and fills the free ranks with the prior
// fill F_i in the same tie-break order (SDD 08 §7.3). Ranks are gapless. The result depends on the set of
// candidates, not on their order, so reruns are identical.
func Pick(cands []Candidate) (entries []Entry, eligible int) {
	var trend, rest []Candidate

	for _, c := range cands {
		if Eligible(c.Trend, c.Actors, true, true, false) {
			trend = append(trend, c)
		} else {
			rest = append(rest, c)
		}
	}

	eligible = len(trend)

	slices.SortFunc(trend, func(a, b Candidate) int {
		return cmp.Or(cmp.Compare(b.Trend, a.Trend), cmp.Compare(b.Prior, a.Prior), cmp.Compare(a.ID, b.ID))
	})

	for _, c := range trend[:min(ListSize, len(trend))] {
		entries = append(entries, Entry{len(entries) + 1, c.ID, c.Trend, "trend"})
	}

	if len(entries) == ListSize {
		return entries, eligible
	}

	fill := func(c Candidate) float64 { return PriorFill(c.Prior, c.DaysSinceRelease) }

	slices.SortFunc(rest, func(a, b Candidate) int {
		return cmp.Or(cmp.Compare(fill(b), fill(a)), cmp.Compare(b.Prior, a.Prior), cmp.Compare(a.ID, b.ID))
	})

	for _, c := range rest[:min(ListSize-len(entries), len(rest))] {
		entries = append(entries, Entry{len(entries) + 1, c.ID, fill(c), "prior"})
	}

	return entries, eligible
}

// candidatesSQL scores the type's public, released items. The pool is every item with activity in the window, every
// item released within 180 days (the only ones whose F_i has a freshness term) and the 10 most popular ones.
const candidatesSQL = `
SELECT i.id::text, i.popularity_prior, coalesce($2::date - i.release_date, 100000) AS since,
  coalesce(a.day, $2::date) AS d,
  coalesce(a.views, 0), coalesce(a.search_clicks, 0),
  coalesce(a.ratings_new_w, 0), coalesce(a.ratings_updated_w, 0), coalesce(a.want_w, 0), coalesce(a.in_progress_w, 0),
  coalesce(a.completed_w, 0), coalesce(a.dropped_w, 0), coalesce(a.list_adds_w, 0), coalesce(a.suggestions_w, 0),
  a.item_id IS NOT NULL
FROM items i
LEFT JOIN item_activity_daily a ON a.item_id = i.id AND a.day > $2::date - $3::int AND a.day <= $2::date
WHERE i.type = $1 AND i.visibility = 'public' AND i.release_status = 'released'
  AND (i.release_date IS NULL OR i.release_date <= $2::date)
  AND NOT EXISTS (SELECT 1 FROM trending_overrides o WHERE o.item_id = i.id AND o.kind = 'banned'
    AND (o.until IS NULL OR o.until > $5))
  AND (a.item_id IS NOT NULL OR i.release_date > $2::date - 180
    OR i.id IN (SELECT id FROM items WHERE type = $1 AND visibility = 'public' AND release_status = 'released'
      AND (release_date IS NULL OR release_date <= $2::date) ORDER BY popularity_prior DESC, id LIMIT $4))
ORDER BY i.id, d`

func candidates(ctx context.Context, q Querier, typ string, today, at time.Time, window int) ([]Candidate, error) {
	rows, err := q.QueryContext(ctx, candidatesSQL, typ, today, window, ListSize, at)
	if err != nil {
		return nil, fmt.Errorf("trending candidates %s: %w", typ, err)
	}

	defer func() { _ = rows.Close() }()

	var (
		out  []Candidate
		days []Day
		cur  Candidate
	)

	flush := func() {
		if cur.ID != "" {
			cur.Trend, cur.Actors = Score(days)
			out = append(out, cur)
		}

		days = days[:0]
	}

	for rows.Next() {
		var (
			c      Candidate
			d      Day
			day    time.Time
			active bool
		)

		if err := rows.Scan(&c.ID, &c.Prior, &c.DaysSinceRelease, &day, &d.Views, &d.SearchClicks, &d.RatingsNewW,
			&d.RatingsUpdatedW, &d.WantW, &d.InProgressW, &d.CompletedW, &d.DroppedW, &d.ListAddsW, &d.SuggestionsW,
			&active); err != nil {
			return nil, err
		}

		if c.ID != cur.ID {
			flush()
			cur = c
		}

		if active {
			d.Age = int(today.Sub(day.UTC().Truncate(24*time.Hour)) / (24 * time.Hour))
			days = append(days, d)
		}
	}

	flush()

	return out, rows.Err()
}

// Window ranks the type over the last windowDays UTC days ending at the day of now and returns the first n entries
// (trend ranks first, then the prior fill). Nothing is stored; the home page uses it with 7 days and 3 entries
// (SDD 08 section 7.2 "Home window").
func Window(ctx context.Context, q Querier, typ string, now time.Time, windowDays, n int) ([]Entry, error) {
	cands, err := candidates(ctx, q, typ, now.UTC().Truncate(24*time.Hour), now.UTC(), windowDays)
	if err != nil {
		return nil, err
	}

	entries, _ := Pick(cands)

	return entries[:min(n, len(entries))], nil
}

// Snapshot writes the top lists for the hour of now, one transaction for the four types and both windows, so readers
// see either the previous snapshot or the whole new one (SDD 08 §7.3). The 30-day list (score v1) feeds the API and
// the for-you fallback; the 7-day list (score v2, RFC-002) feeds the landing page. Running it again within the hour
// yields identical rows.
func Snapshot(ctx context.Context, db *database.DB, now time.Time) error {
	at := now.UTC().Truncate(time.Hour)
	today := at.Truncate(24 * time.Hour)

	err := db.Tx(ctx, func(tx *database.Tx) error {
		for _, typ := range Types {
			cands, err := candidates(ctx, tx, typ, today, at, WindowDays)
			if err != nil {
				return err
			}

			entries, eligible := Pick(cands)

			rows := make([]HomeEntry, len(entries))
			for i, e := range entries {
				rows[i] = HomeEntry{Entry: e}
			}

			fresh, err := write(ctx, tx, typ, APIWindow, at, rows)
			if err != nil {
				return err
			}

			slog.InfoContext(ctx, "trending snapshot written", "type", typ, "window", APIWindow, "eligible", eligible,
				"prior", count(entries, "prior"), "churn", float64(fresh)/ListSize)

			home, err := homeCandidates(ctx, tx, typ, today, at)
			if err != nil {
				return err
			}

			list := PickHome(home, at)

			fresh, err = write(ctx, tx, typ, HomeWindow, at, list)
			if err != nil {
				return err
			}

			slog.InfoContext(
				ctx,
				"trending snapshot written",
				"type",
				typ,
				"window",
				HomeWindow,
				"candidates",
				len(home),
				"trend",
				countHome(list, "trend"),
				"pinned",
				countHome(list, "pinned"),
				"prior",
				countHome(list, "prior"),
				"churn",
				float64(fresh)/HomeListSize,
			)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("trending snapshot: %w", err)
	}

	return nil
}

func count(entries []Entry, basis string) int {
	n := 0

	for _, e := range entries {
		if e.Basis == basis {
			n++
		}
	}

	return n
}

func countHome(entries []HomeEntry, basis string) int {
	n := 0

	for _, e := range entries {
		if e.Basis == basis {
			n++
		}
	}

	return n
}

// write replaces the rows of (at, typ, window) and returns how many entries were not in the previous snapshot of the
// same window.
func write(
	ctx context.Context,
	tx *database.Tx,
	typ string,
	window int,
	at time.Time,
	entries []HomeEntry,
) (int, error) {
	if _, err := tx.ExecContext(ctx, `DELETE FROM trending_snapshots WHERE computed_at = $1 AND item_type = $2
AND window_days = $3`, at, typ, window); err != nil {
		return 0, err
	}

	for _, e := range entries {
		var z, entered, met any

		if window == HomeWindow {
			z, entered, met = float32(e.Z), nullTime(e.EnteredAt), nullTime(e.MetAt)
		}

		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO trending_snapshots (computed_at, item_type, window_days, rank, item_id,
trend_score, basis, z, entered_at, met_at, weak_runs) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
			at,
			typ,
			window,
			e.Rank,
			e.ID,
			float32(e.Score),
			e.Basis,
			z,
			entered,
			met,
			e.WeakRuns,
		); err != nil {
			return 0, err
		}
	}

	var fresh int

	err := tx.QueryRowContext(ctx, `SELECT count(*) FROM trending_snapshots s
WHERE s.computed_at = $1 AND s.item_type = $2 AND s.window_days = $3
  AND NOT EXISTS (SELECT 1 FROM trending_snapshots p WHERE p.item_type = s.item_type AND p.window_days = s.window_days
    AND p.item_id = s.item_id AND p.computed_at =
      (SELECT max(computed_at) FROM trending_snapshots WHERE item_type = $2 AND window_days = $3 AND computed_at < $1))`,
		at, typ, window).Scan(&fresh)

	return fresh, err
}

func nullTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}

	return t
}

// AgeSeconds is the age of the newest snapshot, read by the A14 in-app check (SDD 12 section 7.5).
// It reports sql.ErrNoRows while no snapshot exists.
func AgeSeconds(ctx context.Context, db *database.DB, now time.Time) (float64, error) {
	var at sql.NullTime
	if err := db.QueryRowContext(ctx, `SELECT max(computed_at) FROM trending_snapshots`).Scan(&at); err != nil {
		return 0, err
	}

	if !at.Valid {
		return 0, sql.ErrNoRows
	}

	return now.Sub(at.Time).Seconds(), nil
}

// Locales are the UI locales whose home payload is cached per type.
var Locales = []string{"en", "pt-BR", "es"}

// HighlightsCacheKey is the cache entry of the landing page's highlights and upcoming blocks for one locale (cover
// showcase variant apart); forgotten with HomeCacheKey when a snapshot lands.
func HighlightsCacheKey(locale string, cover bool) string {
	if cover {
		return "home:highlights:v1:" + locale + ":cover"
	}

	return "home:highlights:v1:" + locale
}

// HomeCacheKey is the cache entry of the landing page's trending payload (all four types) for one locale; the hourly
// job forgets it when a new snapshot lands, so the home never serves a list older than the last run.
func HomeCacheKey(locale string) string { return "home:trending:v2:" + locale }
