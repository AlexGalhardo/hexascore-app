//go:build integration

package tests

import (
	"net/url"
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

const (
	itemA = "0197f3a0-0000-7000-8000-0000000000a1"
	itemB = "0197f3a0-0000-7000-8000-0000000000b2"
)

func catalogAdminApp(t *testing.T) *vukemtest.App {
	t.Helper()

	app := vukemtest.New(t, bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}})
	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	for _, q := range []string{
		`INSERT INTO items (id, type, slug, wikidata_qid) VALUES ('` + itemA + `', 'game', 'alpha', 'Q1')`,
		`INSERT INTO items (id, type, slug, wikidata_qid) VALUES ('` + itemB + `', 'game', 'beta', 'Q2')`,
	} {
		if _, err := db.SQL().ExecContext(t.Context(), q); err != nil {
			t.Fatal(err)
		}
	}

	return app
}

func TestSecurity_A01_AdminPagesNotFoundForMembers(t *testing.T) {
	t.Parallel()

	app := catalogAdminApp(t)
	plain := browserFor(t, app, makeUser(t, app, "plain@example.com", "user", true))
	admin := browserFor(t, app, makeUser(t, app, adminEmail, "admin", true))

	for _, p := range []string{"/admin/items", "/admin/items/" + itemA, "/admin/items/" + itemA + "/manage", "/admin/ingestion"} {
		plain.Get(p).AssertNotFound()
		app.Client(t).Get(p).AssertStatus(302)
		admin.Get(p).AssertOK()
	}

	plain.PostForm("/admin/items/"+itemA+"/merge", url.Values{"into": {itemB}, "confirm": {"merge"}}).AssertNotFound()
	plain.PostForm("/admin/items/"+itemA+"/affiliate-links", url.Values{"store": {"gog"}, "url": {"https://x.example/a"}}).
		AssertNotFound()
}

func TestAffiliateLinks(t *testing.T) {
	t.Parallel()

	app := catalogAdminApp(t)
	admin := browserFor(t, app, makeUser(t, app, adminEmail, "admin", true))
	db, _ := app.DB()
	base := "/admin/items/" + itemA

	admin.PostForm(base+"/affiliate-links", url.Values{"store": {"gog"}, "url": {"http://insecure.example"}}).
		AssertStatus(303)
	admin.PostForm(base+"/affiliate-links", url.Values{"store": {"ebay"}, "url": {"https://x.example/a"}}).
		AssertStatus(303)
	admin.PostForm(base+"/affiliate-links", url.Values{"store": {"gog"}, "url": {"https://gog.example/a"}, "active": {"true"}}).
		AssertRedirect(base + "/manage")
	admin.Get(base + "/manage").AssertOK().AssertSee("https://gog.example/a")
	admin.PostForm(base+"/affiliate-links/gog/delete", url.Values{}).AssertRedirect(base + "/manage")

	var n int
	if err := db.SQL().QueryRowContext(t.Context(),
		`SELECT count(*) FROM admin_audit_logs WHERE resource = 'affiliate-links' AND record_id = $1`, itemA+"/gog").
		Scan(&n); err != nil || n != 2 {
		t.Fatalf("audit rows = %d, %v; want create + delete", n, err)
	}
}

func TestAdminMerge(t *testing.T) {
	t.Parallel()

	app := catalogAdminApp(t)
	admin := browserFor(t, app, makeUser(t, app, adminEmail, "admin", true))
	db, _ := app.DB()
	base := "/admin/items/" + itemA

	admin.PostForm(base+"/merge", url.Values{"into": {itemB}}).AssertStatus(303)
	admin.PostForm(base+"/merge", url.Values{"into": {itemA}, "confirm": {"merge"}}).AssertStatus(422)
	admin.PostForm(base+"/merge", url.Values{"into": {itemB}, "confirm": {"merge"}}).
		AssertRedirect("/admin/items/" + itemB)

	var n int
	if err := db.SQL().
		QueryRowContext(t.Context(), `SELECT count(*) FROM items WHERE id = $1`, itemA).
		Scan(&n); err != nil ||
		n != 0 {
		t.Fatalf("source item still present: %d, %v", n, err)
	}

	admin.Get("/admin/items/" + itemA + "/manage").AssertNotFound()
}
