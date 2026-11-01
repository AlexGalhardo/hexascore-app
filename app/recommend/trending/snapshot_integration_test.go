//go:build integration

package trending_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/AlexGalhardo/hexascore-app/app/recommend/trending"
)

func (r *rig) activity(item, day string, views int, ratingsW float64) {
	r.exec(`INSERT INTO item_activity_daily (item_id, day, item_type, views, ratings_new_w)
VALUES ($1, $2, 'game', $3, $4)`, item, day, views, ratingsW)
}

func (r *rig) snapshot(typ string) []string {
	r.t.Helper()

	rows, err := r.db.QueryContext(r.ctx, `SELECT rank, item_id::text, trend_score, basis FROM trending_snapshots
WHERE item_type = $1 AND computed_at = $2 ORDER BY rank`, typ, now.Truncate(time.Hour))
	if err != nil {
		r.t.Fatal(err)
	}

	defer func() { _ = rows.Close() }()

	var out []string

	for rows.Next() {
		var (
			rank       int
			id, basis  string
			trendScore float32
		)
		if err := rows.Scan(&rank, &id, &trendScore, &basis); err != nil {
			r.t.Fatal(err)
		}

		out = append(out, fmt.Sprint(rank, id, trendScore, basis))
	}

	return out
}

func TestBurstDoesNotEnterTop10(t *testing.T) {
	r := newRig(t)

	steady := make([]string, 10)
	for i := range steady {
		steady[i] = r.item()
		r.activity(steady[i], today, 100, 6)
	}

	burst := r.item()
	r.activity(burst, today, 300, 2.625) // T 72.8 but A 2.625 < 5

	if err := trending.Snapshot(r.ctx, r.db, now); err != nil {
		t.Fatal(err)
	}

	got := r.snapshot("game")
	if len(got) != 10 {
		t.Fatalf("got %d rows, want 10: %v", len(got), got)
	}

	for _, row := range got {
		if want := burst; contains(row, want) {
			t.Fatalf("burst item is in the top 10: %v", got)
		}
	}

	before := fmt.Sprint(got)
	if err := trending.Snapshot(r.ctx, r.db, now.Add(20*time.Minute)); err != nil {
		t.Fatal(err)
	}

	if after := fmt.Sprint(r.snapshot("game")); after != before {
		t.Fatalf("rerun within the hour differs:\n%s\n%s", before, after)
	}

	if rows := r.snapshot("movie"); len(rows) != 0 {
		t.Fatalf("type without items has rows: %v", rows)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}

	return false
}

func TestSnapshotPriorFillFromDB(t *testing.T) {
	r := newRig(t)
	a, b := r.item(), r.item()
	r.exec(`UPDATE items SET popularity_prior = 50 WHERE id = $1`, a)
	r.exec(`UPDATE items SET popularity_prior = 5 WHERE id = $1`, b)
	r.exec(`UPDATE items SET visibility = 'hidden' WHERE id = $1`, r.item())

	if err := trending.Snapshot(r.ctx, r.db, now); err != nil {
		t.Fatal(err)
	}

	got := r.snapshot("game")
	if len(got) != 2 || !contains(got[0], a) || !contains(got[0], "prior") {
		t.Fatalf("got %v", got)
	}
}

// BenchmarkSnapshot20k seeds 20k active items with 30 days of activity; a snapshot must take at most 30 s.
func BenchmarkSnapshot20k(b *testing.B) {
	r := newRig(b)
	r.exec(`INSERT INTO items (id, type, slug, wikidata_qid, visibility, popularity_prior)
SELECT gen_random_uuid(), 'game', 'b-' || g, 'Q' || (1000000 + g), 'public', g % 90 FROM generate_series(1, 20000) g`)
	r.exec(`INSERT INTO item_activity_daily (item_id, day, item_type, views, ratings_new_w)
SELECT i.id, ($1::date - d), 'game', 50 + d, 1 FROM items i, generate_series(0, 29) d`, today)

	for b.Loop() {
		start := time.Now()
		if err := trending.Snapshot(r.ctx, r.db, now); err != nil {
			b.Fatal(err)
		}

		if took := time.Since(start); took > 30*time.Second {
			b.Fatalf("snapshot took %s, budget 30s", took)
		}
	}
}

func TestRebuildOnlyTouchesItsItems(t *testing.T) {
	r := newRig(t)
	a, b := r.item(), r.item()
	u := r.user(400, false)
	r.event("rating_created", u, a, now.AddDate(0, 0, -5), nil)
	r.event("rating_created", u, b, now.AddDate(0, 0, -5), nil)

	job := trending.Rebuild{ItemIDs: []string{a}, FromDay: "2026-10-20"}
	if err := job.Run(r.ctx, r.db, now); err != nil {
		t.Fatal(err)
	}

	r.want(a, "2026-10-26", "ratings_new", 1, 1)
	r.want(b, "2026-10-26", "ratings_new", 0, 0)
}

func TestWindowSevenDays(t *testing.T) {
	r := newRig(t)

	recent := make([]string, 4)
	for i := range recent {
		recent[i] = r.item()
		r.activity(recent[i], today, 100, float64(6+i))
	}

	// Strong activity 20 days ago: inside the 30-day snapshot window, outside the 7-day one.
	old := r.item()
	r.activity(old, now.AddDate(0, 0, -20).Format(time.DateOnly), 5000, 40)

	got, err := trending.Window(r.ctx, r.db, "game", now, 7, 3)
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 3 {
		t.Fatalf("got %d entries, want 3: %+v", len(got), got)
	}

	for i, e := range got {
		if e.ID == old || e.Basis != "trend" || e.Rank != i+1 {
			t.Fatalf("entry %d = %+v (old item %s must not rank, ranks are gapless)", i, e, old)
		}
	}

	// The strongest recent item leads.
	if got[0].ID != recent[3] {
		t.Fatalf("first = %s, want %s", got[0].ID, recent[3])
	}

	// The 30-day window still sees the old activity.
	wide, err := trending.Window(r.ctx, r.db, "game", now, 30, 10)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, e := range wide {
		found = found || (e.ID == old && e.Basis == "trend")
	}

	if !found {
		t.Fatalf("30-day window lost the old item: %+v", wide)
	}
}
