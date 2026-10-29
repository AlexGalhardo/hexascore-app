//go:build integration

package tests

import (
	"net/url"
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func TestAdminScoringResourcesReadOnly(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}})
	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	admin := makeUser(t, app, adminEmail, "admin", true)
	plain := makeUser(t, app, "plain@example.com", "user", true)

	const item = "0197f3a0-0000-7000-8000-000000000001"

	for _, q := range []string{
		`INSERT INTO items (id, type, slug, wikidata_qid) VALUES ('` + item + `', 'game', 'burst-game', 'Q42')`,
		`INSERT INTO score_bursts (item_id, started_at, limited_until, trigger, stats)
			VALUES ('` + item + `', now(), now() + interval '1 day', 'volume', '{"inserts": 40}')`,
		`INSERT INTO rating_flags (user_id, item_id, shadow_weight, reason, burst_id)
			VALUES (` + itoa(plain.ID) + `, '` + item + `', 0.25, 'burst_hold', 1)`,
		`INSERT INTO score_actions (kind, item_id, note, before, after)
			VALUES ('burst_approved', '` + item + `', 'looked organic', '{"w": 0.25}', '{"w": 1}')`,
	} {
		if _, err := db.SQL().ExecContext(t.Context(), q); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}

	adminC := browserFor(t, app, admin)
	plainC := browserFor(t, app, plain)

	pages := map[string]string{
		"/admin/score-bursts":           "volume",
		"/admin/rating-flags":           "burst_hold",
		"/admin/score-actions":          "burst_approved",
		"/admin/score-formula-versions": "1",
	}
	for path, see := range pages {
		adminC.Get(path).AssertOK().AssertSee(see)
		plainC.Get(path).AssertNotFound()
	}

	adminC.Get("/admin/score-bursts/1").AssertOK()
	adminC.Get("/admin/score-actions/1").AssertOK()
	adminC.Get("/admin/score-formula-versions/1").AssertOK()

	// The composite key cannot be addressed by one id: no detail page.
	adminC.Get("/admin/rating-flags/" + itoa(plain.ID)).AssertForbidden()

	for path := range pages {
		id := "/1"
		if path == "/admin/rating-flags" {
			id = "/" + itoa(plain.ID)
		}

		adminC.PostForm(path, url.Values{"x": {"1"}}).AssertForbidden()
		adminC.PostForm(path+id, url.Values{"x": {"1"}}).AssertForbidden()
		adminC.PostForm(path+id+"/delete", url.Values{}).AssertForbidden()
	}
}
