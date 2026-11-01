//go:build integration

package tests

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/recommend/eval"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// TestCFTriggerCheck: below the threshold the monthly check records "not triggered" and nothing else; at the
// threshold it says so. Two users hold 10 ratings, one holds 9.
func TestCFTriggerCheck(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(
		t,
		bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}},
	)
	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	exec := func(q string, args ...any) {
		t.Helper()

		if _, err := acc.DB.ExecContext(t.Context(), q, args...); err != nil {
			t.Fatalf("%v\n%s", err, q)
		}
	}

	exec(
		`INSERT INTO data_sources (code, name, licence, url) VALUES ('wikidata', 'W', 'CC0-1.0', 'https://www.wikidata.org')`,
	)

	for i := range 10 {
		exec(`INSERT INTO items (id, type, slug, wikidata_qid) VALUES (gen_random_uuid(), 'game', $1, $2)`,
			fmt.Sprintf("cf-%d", i), fmt.Sprintf("Q%d", 7000+i))
	}

	for i, n := range []int{10, 10, 9} {
		u := makeUser(t, app, fmt.Sprintf("cf%d@example.com", i), "user", true)
		exec(`INSERT INTO ratings (user_id, item_id, c1, c2, c3, c4, c5, c6)
SELECT $1, id, 80, 80, 80, 80, 80, 80 FROM items ORDER BY slug LIMIT $2`, u.ID, n)
	}

	now := time.Date(2026, 11, 1, 5, 30, 0, 0, time.UTC)

	var buf bytes.Buffer

	tr, err := eval.RunMonthly(t.Context(), acc.DB, now, eval.Thresholds{}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	if tr.Triggered || tr.Users != 2 {
		t.Fatalf("trigger = %+v, want not triggered with 2 users", tr)
	}

	if out := buf.String(); !strings.Contains(out, "triggered,false") || strings.Contains(out, "next_step") {
		t.Fatalf("report = %s", out)
	}

	buf.Reset()

	tr, err = eval.RunMonthly(t.Context(), acc.DB, now, eval.Thresholds{MinUsers: 2}, &buf)
	if err != nil || !tr.Triggered || !strings.Contains(buf.String(), "triggered,true") {
		t.Fatalf("at threshold: %+v %v\n%s", tr, err, buf.String())
	}
}
