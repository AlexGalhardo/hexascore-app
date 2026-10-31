//go:build integration

package tests

import (
	"testing"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

// TestMergeSumsActivity: merging an item folds its daily activity into the target, adding counts per day
// (SDD 08 section 8), and leaves no source rows behind.
func TestMergeSumsActivity(t *testing.T) {
	r := newListRig(t)
	db, err := r.app.DB()
	if err != nil {
		t.Fatal(err)
	}

	q, err := r.app.Queue()
	if err != nil {
		t.Fatal(err)
	}

	from, into := r.item("game", "public"), r.item("game", "public")

	for _, row := range []struct {
		item, day string
		views, w  int
	}{
		{from, "2026-10-30", 5, 2},
		{from, "2026-10-31", 7, 1},
		{into, "2026-10-30", 3, 4},
	} {
		if _, err := db.ExecContext(t.Context(), `INSERT INTO item_activity_daily
(item_id, day, item_type, views, search_clicks, ratings_new_w) VALUES ($1, $2, 'game', $3, $3, $4)`,
			row.item, row.day, row.views, float32(row.w)); err != nil {
			t.Fatal(err)
		}
	}

	if err := catalog.Merge(t.Context(), db, q, from, into, 0); err != nil {
		t.Fatal(err)
	}

	got := map[string][3]float64{}

	rows, err := db.QueryContext(t.Context(),
		`SELECT day::text, views, search_clicks, ratings_new_w FROM item_activity_daily WHERE item_id = $1`, into)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var day string

		var v [3]float64
		if err := rows.Scan(&day, &v[0], &v[1], &v[2]); err != nil {
			t.Fatal(err)
		}

		got[day] = v
	}

	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	if want := (map[string][3]float64{"2026-10-30": {8, 8, 6}, "2026-10-31": {7, 7, 1}}); len(got) != 2 ||
		got["2026-10-30"] != want["2026-10-30"] || got["2026-10-31"] != want["2026-10-31"] {
		t.Fatalf("activity of target = %v, want %v", got, want)
	}

	var left int
	if err := db.QueryRowContext(t.Context(),
		`SELECT count(*) FROM item_activity_daily WHERE item_id = $1`, from).Scan(&left); err != nil || left != 0 {
		t.Fatalf("source rows left = %d, err %v", left, err)
	}
}
