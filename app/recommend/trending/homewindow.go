package trending

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Windows of the snapshot table (trending_snapshots.window_days).
const (
	// APIWindow is the 30-day list of the paid API and the for-you fallback (score v1, 10 ranks).
	APIWindow = 30
	// HomeWindow is the 7-day carousel of the landing page (score v2, 12 ranks).
	HomeWindow = 7
)

// minQuality is the provisional quality floor of RFC-002 section 4 until RFC-001 stores a quality_score: cover 25, an
// English synopsis 25, a pt-BR title 15, an es title 15 and a release date 20 must reach 60. An item without a cover
// or without a synopsis therefore never reaches the carousel, but missing translations alone do not block it.
//
// The 60 in homeCandidatesSQL is that floor.
const (
	qualitySQL = `(25 * (i.cover_key IS NOT NULL)::int
  + 25 * EXISTS (SELECT 1 FROM item_translations t WHERE t.item_id = i.id AND t.locale = 'en' AND t.synopsis IS NOT NULL)::int
  + 15 * EXISTS (SELECT 1 FROM item_translations t WHERE t.item_id = i.id AND t.locale = 'pt-BR')::int
  + 15 * EXISTS (SELECT 1 FROM item_translations t WHERE t.item_id = i.id AND t.locale = 'es')::int
  + 20 * (i.release_date IS NOT NULL)::int)`
	activeOverrideSQL = `(o.until IS NULL OR o.until > $3)`
)

// homeCandidatesSQL lists the public, released (or soon announced) items of a type that clear the quality floor and are
// not banned, with the activity of the last 7 days. $1 type, $2 UTC day, $3 the run's hour, $4 how many of the most
// popular items to keep as prior fill. The pool is every item with activity, every release within 180 days, the most
// popular ones and the pinned ones (a pin skips the quality floor, not visibility or a ban).
const homeCandidatesSQL = `
WITH base AS (
  SELECT i.id, i.popularity_prior, i.release_date
  FROM items i
  WHERE i.type = $1 AND i.visibility = 'public'
    AND ((i.release_status = 'released' AND (i.release_date IS NULL OR i.release_date <= $2::date))
      OR (i.release_status = 'announced' AND i.release_date IS NOT NULL AND i.release_date <= $2::date + 90))
    AND NOT EXISTS (SELECT 1 FROM trending_overrides o WHERE o.item_id = i.id AND o.kind = 'banned' AND ` + activeOverrideSQL + `)
    AND (` + qualitySQL + ` >= 60
      OR EXISTS (SELECT 1 FROM trending_overrides o WHERE o.item_id = i.id AND o.kind = 'pinned' AND ` + activeOverrideSQL + `))
), pop AS (
  SELECT id FROM base ORDER BY popularity_prior DESC, id LIMIT $4
)
SELECT b.id::text, b.popularity_prior, coalesce($2::date - b.release_date, 100000) AS since,
  coalesce(a.day, $2::date) AS d,
  coalesce(a.views, 0), coalesce(a.search_clicks, 0),
  coalesce(a.ratings_new_w, 0), coalesce(a.ratings_updated_w, 0), coalesce(a.want_w, 0), coalesce(a.in_progress_w, 0),
  coalesce(a.completed_w, 0), coalesce(a.dropped_w, 0), coalesce(a.list_adds_w, 0),
  a.item_id IS NOT NULL
FROM base b
LEFT JOIN item_activity_daily a ON a.item_id = b.id AND a.day > $2::date - 7 AND a.day <= $2::date
WHERE a.item_id IS NOT NULL OR b.release_date > $2::date - 180 OR b.id IN (SELECT id FROM pop)
  OR EXISTS (SELECT 1 FROM trending_overrides o WHERE o.item_id = b.id AND o.kind = 'pinned' AND ` + activeOverrideSQL + `)
ORDER BY b.id, d`

type rawCandidate struct {
	HomeCandidate
	days []Day
}

// homeCandidates scores the type's pool with score v2 and the z-score against the items that had activity.
func homeCandidates(ctx context.Context, q Querier, typ string, today, at time.Time) ([]HomeCandidate, error) {
	rows, err := q.QueryContext(ctx, homeCandidatesSQL, typ, today, at, HomeListSize)
	if err != nil {
		return nil, fmt.Errorf("trending home candidates %s: %w", typ, err)
	}

	defer func() { _ = rows.Close() }()

	var (
		pool []rawCandidate
		cur  rawCandidate
	)

	flush := func() {
		if cur.ID != "" {
			pool = append(pool, cur)
		}

		cur = rawCandidate{}
	}

	for rows.Next() {
		var (
			id     string
			prior  float64
			since  float64
			day    time.Time
			d      Day
			active bool
		)

		if err := rows.Scan(&id, &prior, &since, &day, &d.Views, &d.SearchClicks, &d.RatingsNewW, &d.RatingsUpdatedW,
			&d.WantW, &d.InProgressW, &d.CompletedW, &d.DroppedW, &d.ListAddsW, &active); err != nil {
			return nil, err
		}

		if id != cur.ID {
			flush()
			cur = rawCandidate{ID: id, Prior: prior, DaysSinceRelease: since}
		}

		if active {
			d.Age = int(today.Sub(day.UTC().Truncate(24*time.Hour)) / (24 * time.Hour))
			cur.days = append(cur.days, d)
		}
	}

	flush()

	if err := rows.Err(); err != nil {
		return nil, err
	}

	ids := make([]string, len(pool))
	for i, p := range pool {
		ids[i] = p.ID
	}

	ext, err := externalDeltas(ctx, q, ids, today)
	if err != nil {
		return nil, err
	}

	var counted []float64

	for i := range pool {
		p := &pool[i]
		p.Trend, p.Actors = ScoreV2(
			Signals{Days: p.days, ExternalDelta: ext[p.ID], DaysSinceRelease: p.DaysSinceRelease},
		)

		if p.Actors >= 1 {
			counted = append(counted, p.Trend)
		}
	}

	basis := NewZBasis(counted)

	structure, err := structureOf(ctx, q, ids)
	if err != nil {
		return nil, err
	}

	pins, err := pinsOf(ctx, q, ids, at)
	if err != nil {
		return nil, err
	}

	prev, err := previousHome(ctx, q, typ, at)
	if err != nil {
		return nil, err
	}

	out := make([]HomeCandidate, len(pool))

	for i, p := range pool {
		c := p.HomeCandidate
		c.Z = basis.Z(c.Trend)
		c.DaysSinceRelease = max(0, c.DaysSinceRelease)
		c.Franchise, c.Genre = structure[c.ID].franchise, structure[c.ID].genre
		c.PinOrder, c.Pinned = pins[c.ID], pins[c.ID] > 0
		c.Prev = prev[c.ID]
		out[i] = c
	}

	return out, nil
}

// uuidArray renders ids as a Postgres array literal for `$n::uuid[]`.
func uuidArray(ids []string) string { return "{" + strings.Join(ids, ",") + "}" }

// externalDeltas is, per item, the mean 7-day relative change of its external popularity values (RFC-002 section 2).
// A source with no value 7 days earlier, or none in the last two days, contributes nothing.
func externalDeltas(ctx context.Context, q Querier, ids []string, today time.Time) (map[string]float64, error) {
	out := map[string]float64{}
	if len(ids) == 0 {
		return out, nil
	}

	rows, err := q.QueryContext(ctx, `SELECT t.item_id::text, t.value, p.value
FROM item_popularity_external t
JOIN item_popularity_external p ON p.item_id = t.item_id AND p.source = t.source AND p.day = t.day - 7
WHERE t.item_id = ANY($1::uuid[]) AND t.day <= $2::date AND t.day >= $2::date - 1
  AND t.day = (SELECT max(x.day) FROM item_popularity_external x
    WHERE x.item_id = t.item_id AND x.source = t.source AND x.day <= $2::date)`, uuidArray(ids), today)
	if err != nil {
		return nil, fmt.Errorf("trending external deltas: %w", err)
	}

	defer func() { _ = rows.Close() }()

	sum, n := map[string]float64{}, map[string]int{}

	for rows.Next() {
		var (
			id            string
			now, previous float64
		)

		if err := rows.Scan(&id, &now, &previous); err != nil {
			return nil, err
		}

		sum[id] += ExternalDelta(now, previous)
		n[id]++
	}

	for id, s := range sum {
		out[id] = s / float64(n[id])
	}

	return out, rows.Err()
}

type structureInfo struct{ franchise, genre string }

// structureOf reads the franchise (role franchise, else series) and the primary genre (lowest position) of each item.
func structureOf(ctx context.Context, q Querier, ids []string) (map[string]structureInfo, error) {
	out := map[string]structureInfo{}
	if len(ids) == 0 {
		return out, nil
	}

	rows, err := q.QueryContext(ctx, `SELECT ie.item_id::text, ie.role, ie.entity_id::text
FROM item_entities ie
WHERE ie.item_id = ANY($1::uuid[]) AND ie.role IN ('franchise', 'series', 'genre')
ORDER BY ie.item_id, ie.role, ie.position, ie.entity_id`, uuidArray(ids))
	if err != nil {
		return nil, fmt.Errorf("trending structure: %w", err)
	}

	defer func() { _ = rows.Close() }()

	franchise := map[string]string{}
	series := map[string]string{}

	for rows.Next() {
		var id, role, entity string
		if err := rows.Scan(&id, &role, &entity); err != nil {
			return nil, err
		}

		s := out[id]

		switch role {
		case "genre":
			if s.genre == "" {
				s.genre = entity
			}
		case "franchise":
			if _, ok := franchise[id]; !ok {
				franchise[id] = entity
			}
		case "series":
			if _, ok := series[id]; !ok {
				series[id] = entity
			}
		}

		out[id] = s
	}

	for id, s := range out {
		s.franchise = franchise[id]
		if s.franchise == "" {
			s.franchise = series[id]
		}

		out[id] = s
	}

	return out, rows.Err()
}

// pinsOf maps each pinned item to its 1-based pin order (oldest pin first); unpinned items are absent.
func pinsOf(ctx context.Context, q Querier, ids []string, at time.Time) (map[string]int, error) {
	out := map[string]int{}
	if len(ids) == 0 {
		return out, nil
	}

	rows, err := q.QueryContext(ctx, `SELECT item_id::text, row_number() OVER (ORDER BY min(created_at), min(id))
FROM trending_overrides o
WHERE o.kind = 'pinned' AND `+strings.ReplaceAll(activeOverrideSQL, "$3", "$2")+` AND o.item_id = ANY($1::uuid[])
GROUP BY item_id`, uuidArray(ids), at)
	if err != nil {
		return nil, fmt.Errorf("trending pins: %w", err)
	}

	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var (
			id string
			n  int
		)

		if err := rows.Scan(&id, &n); err != nil {
			return nil, err
		}

		out[id] = n
	}

	return out, rows.Err()
}

// previousHome reads the state of the newest 7-day snapshot of the type that is older than this run, so a rerun within
// the hour sees the same past and writes identical rows.
func previousHome(ctx context.Context, q Querier, typ string, at time.Time) (map[string]*Prev, error) {
	rows, err := q.QueryContext(ctx, `SELECT item_id::text, entered_at, met_at, weak_runs FROM trending_snapshots
WHERE item_type = $1 AND window_days = 7 AND computed_at =
  (SELECT max(computed_at) FROM trending_snapshots WHERE item_type = $1 AND window_days = 7 AND computed_at < $2)`,
		typ, at)
	if err != nil {
		return nil, fmt.Errorf("trending previous home: %w", err)
	}

	defer func() { _ = rows.Close() }()

	out := map[string]*Prev{}

	for rows.Next() {
		var (
			id           string
			entered, met *time.Time
			p            Prev
		)

		if err := rows.Scan(&id, &entered, &met, &p.WeakRuns); err != nil {
			return nil, err
		}

		if entered != nil {
			p.EnteredAt = *entered
		}

		if met != nil {
			p.MetAt = *met
		}

		out[id] = &p
	}

	return out, rows.Err()
}
