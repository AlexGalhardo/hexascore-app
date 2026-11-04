//go:build integration

package trending_test

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/uuid"

	"github.com/AlexGalhardo/hexascore-app/app/recommend/trending"
)

// rich is an item that clears the quality floor: cover, English synopsis and a release date (25+25+20 = 70).
func (r *rig) rich() string {
	r.t.Helper()

	id := r.item()
	r.exec(
		`INSERT INTO data_sources (code, name, licence, url) VALUES ('wikidata', 'Wikidata', 'CC0-1.0', 'https://www.wikidata.org')
ON CONFLICT DO NOTHING`,
	)
	r.exec(`UPDATE items SET cover_key = 'covers/x.webp', release_date = '2020-01-01' WHERE id = $1`, id)
	r.exec(
		`INSERT INTO item_translations (item_id, locale, title, title_source, synopsis, synopsis_source, synopsis_licence)
VALUES ($1, 'en', 'Title', 'wikidata', 'A synopsis.', 'wikidata', 'CC0-1.0')`,
		id,
	)

	return id
}

// hot adds today's engaged activity of weight w (A = 2w, T grows with w).
func (r *rig) hot(item string, w float64) {
	r.t.Helper()
	r.exec(`INSERT INTO item_activity_daily (item_id, day, item_type, views, ratings_new_w, completed_w)
VALUES ($1, $2, 'game', 50, $3, $3) ON CONFLICT (item_id, day) DO UPDATE SET ratings_new_w = $3, completed_w = $3`,
		item, today, w)
}

// crowd makes n rich items with growing activity so the z-scores have a population; it returns their ids.
func (r *rig) crowd(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = r.rich()
		r.hot(out[i], float64(3+i))
	}

	return out
}

// home returns the 7-day list of the type at the run of now as "rank:id:basis".
func (r *rig) home(typ string, at time.Time) []string {
	r.t.Helper()

	rows, err := r.db.QueryContext(r.ctx, `SELECT rank, item_id::text, basis FROM trending_snapshots
WHERE item_type = $1 AND window_days = 7 AND computed_at = $2 ORDER BY rank`, typ, at.Truncate(time.Hour))
	if err != nil {
		r.t.Fatal(err)
	}

	defer func() { _ = rows.Close() }()

	var out []string

	for rows.Next() {
		var (
			rank      int
			id, basis string
		)
		if err := rows.Scan(&rank, &id, &basis); err != nil {
			r.t.Fatal(err)
		}

		out = append(out, fmt.Sprintf("%d:%s:%s", rank, id, basis))
	}

	return out
}

func has(list []string, id string) bool {
	return slices.ContainsFunc(list, func(s string) bool { return contains(s, id) })
}

func (r *rig) snap(at time.Time) {
	r.t.Helper()

	if err := trending.Snapshot(r.ctx, r.db, at); err != nil {
		r.t.Fatal(err)
	}
}

func TestHomeListQualityFloor(t *testing.T) {
	r := newRig(t)
	good := r.rich()
	r.hot(good, 6)
	r.crowd(5)

	bare := r.item() // no cover, no synopsis: quality 20 of 100
	r.hot(bare, 20)

	noCover := r.rich()
	r.exec(`UPDATE items SET cover_key = NULL WHERE id = $1`, noCover) // 50: below the floor
	r.hot(noCover, 20)

	// a synopsis, both translated titles and a date reach 75 without a cover: the cover is still required
	richNoCover := r.rich()
	r.exec(`UPDATE items SET cover_key = NULL WHERE id = $1`, richNoCover)
	r.exec(
		`INSERT INTO item_translations (item_id, locale, title, title_source) VALUES ($1, 'pt-BR', 'Título', 'wikidata'),
($1, 'es', 'Título', 'wikidata')`,
		richNoCover,
	)
	r.hot(richNoCover, 30)

	r.snap(now)

	list := r.home("game", now)
	if !has(list, good) {
		t.Fatalf("item over the floor is missing: %v", list)
	}

	for _, id := range []string{bare, noCover, richNoCover} {
		if has(list, id) {
			t.Fatalf("item under the quality floor is on the home list: %v", list)
		}
	}
}

func TestHomePinsAndBans(t *testing.T) {
	r := newRig(t)
	crowd := r.crowd(6)

	// no activity and no synopsis: a pin skips the quality floor and the entry rule, but not the cover (the landing
	// never shows an item without one)
	pinned := r.item()
	r.exec(`UPDATE items SET cover_key = 'covers/pin.webp' WHERE id = $1`, pinned)
	r.exec(`INSERT INTO trending_overrides (item_id, kind, reason) VALUES ($1, 'pinned', 'editor pick')`, pinned)

	coverless := r.item()
	r.exec(`INSERT INTO trending_overrides (item_id, kind, reason) VALUES ($1, 'pinned', 'no cover')`, coverless)

	banned := r.rich()
	r.hot(banned, 30) // the strongest item of the type
	r.exec(`INSERT INTO trending_overrides (item_id, kind, reason) VALUES ($1, 'banned', 'spam')`, banned)

	lapsed := r.rich()
	r.hot(lapsed, 25)
	r.exec(
		`INSERT INTO trending_overrides (item_id, kind, until) VALUES ($1, 'banned', $2)`,
		lapsed,
		now.Add(-time.Hour),
	)

	r.snap(now)

	list := r.home("game", now)
	if has(list, coverless) {
		t.Fatalf("a pinned item without a cover must stay off the list: %v", list)
	}
	if len(list) == 0 || list[0] != fmt.Sprintf("1:%s:pinned", pinned) {
		t.Fatalf("pin must take rank 1: %v", list)
	}

	if has(list, banned) {
		t.Fatalf("banned item on the home list: %v", list)
	}

	if !has(list, lapsed) || !has(list, crowd[0]) {
		t.Fatalf("an expired ban must not hide an item: %v", list)
	}

	// the 30-day API list drops the ban as well
	rows, err := r.db.QueryContext(r.ctx, `SELECT item_id::text FROM trending_snapshots WHERE window_days = 30`)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			t.Fatal(err)
		}

		if id == banned {
			t.Fatal("banned item on the 30-day list")
		}
	}
}

func TestHomeCooldownAcrossRuns(t *testing.T) {
	r := newRig(t)
	r.crowd(6)

	fading := r.rich()
	r.hot(fading, 9)

	r.snap(now)

	if !has(r.home("game", now), fading) {
		t.Fatalf("setup: item should enter: %v", r.home("game", now))
	}

	var entered time.Time
	if err := r.db.QueryRowContext(r.ctx, `SELECT entered_at FROM trending_snapshots WHERE item_id = $1 AND window_days = 7`,
		fading).
		Scan(&entered); err != nil ||
		!entered.Equal(now.Truncate(time.Hour)) {
		t.Fatalf("entered_at = %v, %v", entered, err)
	}

	r.exec(`DELETE FROM item_activity_daily WHERE item_id = $1`, fading) // its signal is gone

	next := now.Add(time.Hour)
	r.snap(next)

	if !has(r.home("game", next), fading) { // one weak run is forgiven
		t.Fatalf("cooldown must keep the item for one weak run: %v", r.home("game", next))
	}

	if err := r.db.QueryRowContext(r.ctx, `SELECT entered_at FROM trending_snapshots WHERE item_id = $1 AND window_days = 7
AND computed_at = $2`, fading, next.Truncate(time.Hour)).
		Scan(&entered); err != nil ||
		!entered.Equal(now.Truncate(time.Hour)) {
		t.Fatalf("entered_at must survive the run: %v, %v", entered, err)
	}

	again := now.Add(2 * time.Hour) // second consecutive run below zero
	r.snap(again)

	if list := r.home("game", again); has(list, fading) && contains(fmt.Sprint(list), fading+":trend") {
		t.Fatalf("two weak runs must evict: %v", list)
	}
}

func TestHomeRerunWithinTheHourIsIdentical(t *testing.T) {
	r := newRig(t)
	r.crowd(8)

	r.snap(now)
	first := fmt.Sprint(r.home("game", now))

	r.snap(now.Add(25 * time.Minute))

	if second := fmt.Sprint(r.home("game", now)); first != second {
		t.Fatalf("rerun differs:\n%s\n%s", first, second)
	}
}

func TestHomeExternalDeltaRaisesTheRank(t *testing.T) {
	r := newRig(t)
	r.crowd(6)

	a, b := r.rich(), r.rich()
	r.hot(a, 12) // both twins clear the entry rule; only the external signal separates them
	r.hot(b, 12)

	// the larger id loses the id tie-break, so only the external signal can put it first
	boosted, plain := a, b
	if b > a {
		boosted, plain = b, a
	}

	r.exec(`INSERT INTO item_popularity_external (item_id, day, source, value) VALUES
($1, $2, 'tmdb', 200), ($1, $3, 'tmdb', 100)`, boosted, today, "2026-10-24")

	r.snap(now)

	list := r.home("game", now)
	if !has(list, boosted) || !has(list, plain) {
		t.Fatalf("both items should be listed: %v", list)
	}

	rank := func(id string) int { return slices.IndexFunc(list, func(s string) bool { return contains(s, id) }) }
	if rank(boosted) >= rank(plain) {
		t.Fatalf("an item whose external popularity doubled must outrank its twin: %v", list)
	}
}

func TestHomeFranchiseCap(t *testing.T) {
	r := newRig(t)
	r.crowd(4)
	r.exec(
		`INSERT INTO data_sources (code, name, licence, url) VALUES ('wikidata', 'Wikidata', 'CC0-1.0', 'https://www.wikidata.org')
ON CONFLICT DO NOTHING`,
	)

	franchise := uuid.NewV7().String()
	r.exec(`INSERT INTO entities (id, kind, slug, default_name) VALUES ($1, 'franchise', 'saga', 'Saga')`, franchise)

	var saga []string

	for range 4 {
		id := r.rich()
		r.hot(id, 12) // all four outrank the crowd
		r.exec(
			`INSERT INTO item_entities (item_id, entity_id, role, source) VALUES ($1, $2, 'franchise', 'wikidata')`,
			id,
			franchise,
		)

		saga = append(saga, id)
	}

	r.snap(now)

	n := 0

	for _, id := range saga {
		if has(r.home("game", now), id) {
			n++
		}
	}

	if n != trending.MaxPerFranchise {
		t.Fatalf(
			"%d entries of one franchise on the list, want %d: %v",
			n,
			trending.MaxPerFranchise,
			r.home("game", now),
		)
	}
}

func TestLatestReadsEachWindow(t *testing.T) {
	r := newRig(t)
	r.crowd(6)
	r.snap(now)

	_, home, err := trending.Latest(r.ctx, r.db, "game", trending.HomeWindow)
	if err != nil || len(home) == 0 {
		t.Fatalf("home window: %v, %v", home, err)
	}

	_, api, err := trending.Latest(r.ctx, r.db, "game", trending.APIWindow)
	if err != nil || len(api) == 0 || len(api) > trending.ListSize {
		t.Fatalf("api window: %v, %v", api, err)
	}
}
