//go:build integration

package tests

import (
	"testing"
	"time"
)

// Audit T-05: the purge erases time_to_beat and search_events rows, anonymises score_actions and score_bursts, and
// keeps rating_flags with the kept ratings.
func TestSecurity_A03_PurgeRemovesScoringRows(t *testing.T) {
	app, acc, db := deletionApp(t)
	ctx := t.Context()

	ids := map[string]int64{"gone@example.com": 0, "other@example.com": 0}
	for e := range ids {
		register(t, app, app.Client(t), e)

		var id int64
		if err := db.QueryRowContext(ctx, `SELECT id FROM users WHERE email = $1`, e).Scan(&id); err != nil {
			t.Fatal(err)
		}

		ids[e] = id
	}

	gone, other := ids["gone@example.com"], ids["other@example.com"]
	exec := func(q string, args ...any) {
		t.Helper()

		if _, err := db.ExecContext(ctx, q, args...); err != nil {
			t.Fatal(err)
		}
	}

	exec(`INSERT INTO data_sources (code, name, licence, url) VALUES ('wikidata','Wikidata','CC0-1.0','https://wikidata.org')`)
	exec(`INSERT INTO items (id, type, slug, wikidata_qid) VALUES ('00000000-0000-7000-8000-0000000000a1', 'game', 'g', 'Q1')`)
	const item = "00000000-0000-7000-8000-0000000000a1"
	exec(`INSERT INTO score_bursts (id, item_id, started_at, limited_until, trigger, stats, resolution, resolved_at, resolved_by)
OVERRIDING SYSTEM VALUE VALUES (1, $1, now(), now(), 'volume', '{}', 'approved', now(), $2)`, item, gone)
	for _, u := range []int64{gone, other} {
		exec(`INSERT INTO time_to_beat (user_id, item_id, main_minutes) VALUES ($1, $2, 120)`, u, item)
		exec(`INSERT INTO rating_flags (user_id, item_id, shadow_weight, reason) VALUES ($1, $2, 0.5, 'admin_downweight')`, u, item)
		exec(`INSERT INTO score_actions (kind, item_id, user_id, actor) VALUES ('status_changed', $2, $1, $1)`, u, item)
		exec(`INSERT INTO search_events (search_id, surface, query_norm, user_id) VALUES (gen_random_uuid(), 'web', 'zelda', $1)`, u)
	}
	exec(`UPDATE users SET deletion_scheduled_at = $1 WHERE id = $2`, time.Now().UTC().AddDate(0, 0, -31), gone)

	if done, err := acc.PurgeDue(ctx); err != nil || len(done) != 1 {
		t.Fatalf("purged %v, %v", done, err)
	}

	for q, want := range map[string]int{
		`SELECT count(*) FROM time_to_beat WHERE user_id = $1`:                                          0,
		`SELECT count(*) FROM search_events WHERE user_id = $1`:                                         0,
		`SELECT count(*) FROM score_actions WHERE user_id = $1 OR actor = $1`:                           0,
		`SELECT count(*) FROM score_bursts WHERE resolved_by = $1`:                                      0,
		`SELECT count(*) FROM score_actions WHERE user_id IS NULL AND actor IS NULL AND $1::bigint > 0`: 1,
		`SELECT count(*) FROM rating_flags WHERE user_id = $1`:                                          1,
	} {
		if n := count(t, db, q, gone); n != want {
			t.Errorf("%s = %d, want %d", q, n, want)
		}
	}

	for _, q := range []string{
		`SELECT count(*) FROM time_to_beat WHERE user_id = $1`,
		`SELECT count(*) FROM search_events WHERE user_id = $1`,
		`SELECT count(*) FROM score_actions WHERE user_id = $1 AND actor = $1`,
	} {
		if n := count(t, db, q, other); n != 1 {
			t.Errorf("other user's row touched: %s = %d", q, n)
		}
	}
}
