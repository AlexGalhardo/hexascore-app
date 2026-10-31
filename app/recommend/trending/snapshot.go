package trending

import (
	"cmp"
	"context"
	"database/sql"
	"fmt"
	"slices"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
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
  AND (a.item_id IS NOT NULL OR i.release_date > $2::date - 180
    OR i.id IN (SELECT id FROM items WHERE type = $1 AND visibility = 'public' AND release_status = 'released'
      AND (release_date IS NULL OR release_date <= $2::date) ORDER BY popularity_prior DESC, id LIMIT $4))
ORDER BY i.id, d`

func candidates(ctx context.Context, q Querier, typ string, today time.Time, window int) ([]Candidate, error) {
	rows, err := q.QueryContext(ctx, candidatesSQL, typ, today, window, ListSize)
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
	cands, err := candidates(ctx, q, typ, now.UTC().Truncate(24*time.Hour), windowDays)
	if err != nil {
		return nil, err
	}

	entries, _ := Pick(cands)

	return entries[:min(n, len(entries))], nil
}

// Snapshot writes the top lists for the hour of now, one transaction for the four types, so readers see either the
// previous snapshot or the whole new one (SDD 08 §7.3). Running it again within the hour yields identical rows.
func Snapshot(ctx context.Context, db *database.DB, now time.Time) error {
	started := time.Now()
	at := now.UTC().Truncate(time.Hour)
	today := at.Truncate(24 * time.Hour)
	meter := otel.Meter("hexascore")
	eligibleG, _ := meter.Int64Gauge("trending.eligible_items")
	priorG, _ := meter.Int64Gauge("trending.prior_slots")
	churnG, _ := meter.Float64Gauge("trending.churn")
	duration, _ := meter.Float64Histogram("trending.compute.duration", metric.WithUnit("s"))

	err := db.Tx(ctx, func(tx *database.Tx) error {
		for _, typ := range Types {
			cands, err := candidates(ctx, tx, typ, today, WindowDays)
			if err != nil {
				return err
			}

			entries, eligible := Pick(cands)

			fresh, err := write(ctx, tx, typ, at, entries)
			if err != nil {
				return err
			}

			prior := 0

			for _, e := range entries {
				if e.Basis == "prior" {
					prior++
				}
			}

			attr := metric.WithAttributes(attribute.String("type", typ))
			eligibleG.Record(ctx, int64(eligible), attr)
			priorG.Record(ctx, int64(prior), attr)
			churnG.Record(ctx, float64(fresh)/ListSize, attr)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("trending snapshot: %w", err)
	}

	duration.Record(ctx, time.Since(started).Seconds())

	return nil
}

// write replaces the rows of (at, typ) and returns how many entries were not in the previous snapshot.
func write(ctx context.Context, tx *database.Tx, typ string, at time.Time, entries []Entry) (int, error) {
	if _, err := tx.ExecContext(ctx, `DELETE FROM trending_snapshots WHERE computed_at = $1 AND item_type = $2`,
		at, typ); err != nil {
		return 0, err
	}

	for _, e := range entries {
		if _, err := tx.ExecContext(ctx, `INSERT INTO trending_snapshots (computed_at, item_type, rank, item_id,
trend_score, basis) VALUES ($1, $2, $3, $4, $5, $6)`, at, typ, e.Rank, e.ID, float32(e.Score), e.Basis); err != nil {
			return 0, err
		}
	}

	var fresh int

	err := tx.QueryRowContext(ctx, `SELECT count(*) FROM trending_snapshots s
WHERE s.computed_at = $1 AND s.item_type = $2
  AND NOT EXISTS (SELECT 1 FROM trending_snapshots p WHERE p.item_type = s.item_type AND p.item_id = s.item_id
    AND p.computed_at = (SELECT max(computed_at) FROM trending_snapshots WHERE item_type = $2 AND computed_at < $1))`,
		at, typ).Scan(&fresh)

	return fresh, err
}

// AgeSeconds is the age of the newest snapshot, for the trending.snapshot_age_seconds gauge and alert A14.
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

// CacheKey is the cache entry of the home trending payload of one type and locale; the hourly job forgets it.
func CacheKey(typ, locale string) string { return "trending:v1:" + typ + ":" + locale }
