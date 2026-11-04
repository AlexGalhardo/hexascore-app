//go:build integration

package tests

import (
	"net/url"
	"strconv"
	"strings"
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
	if got := inertiaProp(t, admin.Get(base+"/manage").AssertOK().Body(), "Admin/ItemManage", "affiliates"); !strings.Contains(got, "https://gog.example/a") {
		t.Fatalf("affiliate link missing from the manage page: %s", got)
	}
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
	admin.PostForm(base+"/merge", url.Values{"into": {itemA}, "confirm": {"merge"}}).AssertStatus(409)
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

// ADR 0026: the source records screen is a React page; its data travels as Inertia props.
func TestAdminIngestionSourcesPage(t *testing.T) {
	t.Parallel()

	app := catalogAdminApp(t)
	admin := browserFor(t, app, makeUser(t, app, adminEmail, "admin", true))

	got, props := inertiaPage(t, admin.Get("/admin/ingestion").AssertOK().Body())
	if got != "Admin/IngestSources" {
		t.Fatalf("component = %q, want Admin/IngestSources", got)
	}

	if _, ok := props["ready"].(bool); !ok {
		t.Fatalf("ready prop missing: %v", props["ready"])
	}
}

// TestTrendingOverrides: an admin pins or bans an item from the trending lists (RFC-002 section 4); the change is
// audited and one active override per kind and item is kept.
func TestTrendingOverrides(t *testing.T) {
	t.Parallel()

	app := catalogAdminApp(t)
	plain := browserFor(t, app, makeUser(t, app, "plain@example.com", "user", true))
	admin := browserFor(t, app, makeUser(t, app, adminEmail, "admin", true))
	db, _ := app.DB()
	base := "/admin/items/" + itemA

	count := func(q string, args ...any) int {
		t.Helper()

		var n int
		if err := db.SQL().QueryRowContext(t.Context(), q, args...).Scan(&n); err != nil {
			t.Fatal(err)
		}

		return n
	}

	plain.PostForm(base+"/trending", url.Values{"kind": {"pinned"}}).AssertNotFound()
	admin.PostForm(base+"/trending", url.Values{"kind": {"featured"}}).AssertStatus(303)
	admin.PostForm(base+"/trending", url.Values{"kind": {"banned"}, "until": {"2020-01-01"}}).AssertStatus(303)
	admin.PostForm(base+"/trending", url.Values{"kind": {"banned"}, "until": {"not a date"}}).AssertStatus(303)

	if n := count(`SELECT count(*) FROM trending_overrides`); n != 0 {
		t.Fatalf("invalid requests wrote %d rows", n)
	}

	admin.PostForm(base+"/trending", url.Values{"kind": {"pinned"}, "reason": {"editor pick"}}).
		AssertRedirect(base + "/manage")
	admin.PostForm(base+"/trending", url.Values{"kind": {"pinned"}, "reason": {"still our pick"}, "until": {"2099-01-01"}}).
		AssertRedirect(base + "/manage")

	if n := count(`SELECT count(*) FROM trending_overrides WHERE item_id = $1 AND kind = 'pinned'`, itemA); n != 1 {
		t.Fatalf("pinned rows = %d, want the second request to replace the first", n)
	}

	admin.Get(base + "/manage").AssertOK().AssertSee("still our pick")

	var id int64
	if err := db.SQL().QueryRowContext(t.Context(), `SELECT id FROM trending_overrides WHERE item_id = $1`, itemA).
		Scan(&id); err != nil {
		t.Fatal(err)
	}

	other := "/admin/items/" + itemB + "/trending/" + strconv.FormatInt(id, 10) + "/delete"
	admin.PostForm(other, url.Values{}).AssertNotFound() // the override belongs to another item

	plain.PostForm(base+"/trending/"+strconv.FormatInt(id, 10)+"/delete", url.Values{}).AssertNotFound()
	admin.PostForm(base+"/trending/"+strconv.FormatInt(id, 10)+"/delete", url.Values{}).AssertRedirect(base + "/manage")

	if n := count(`SELECT count(*) FROM trending_overrides`); n != 0 {
		t.Fatalf("override still stored after delete: %d", n)
	}

	if n := count(`SELECT count(*) FROM admin_audit_logs WHERE resource = 'trending-overrides'`); n != 3 {
		t.Fatalf("audit rows = %d, want create, update and delete", n)
	}
}

// TestAdminOffers: an admin lists the offers of an item, deactivates or restores one and adds one by hand (RFC-003
// OF-07). Every change is audited; invalid input writes nothing; another item's offer is a 404.
func TestAdminOffers(t *testing.T) {
	t.Parallel()

	app := catalogAdminApp(t)
	plain := browserFor(t, app, makeUser(t, app, "plain@example.com", "user", true))
	admin := browserFor(t, app, makeUser(t, app, adminEmail, "admin", true))
	db, _ := app.DB()
	base := "/admin/items/" + itemA

	count := func(q string, args ...any) int {
		t.Helper()

		var n int
		if err := db.SQL().QueryRowContext(t.Context(), q, args...).Scan(&n); err != nil {
			t.Fatal(err)
		}

		return n
	}

	good := url.Values{
		"store":   {"steam"},
		"country": {"BR"},
		"kind":    {"buy_digital"},
		"url":     {"https://store.steampowered.com/app/1"},
	}
	plain.PostForm(base+"/offers", good).AssertNotFound()

	for name, bad := range map[string]url.Values{
		"unknown store":   {"store": {"nope"}, "country": {"BR"}, "kind": {"buy"}, "url": {"https://a.example/x"}},
		"unknown country": {"store": {"steam"}, "country": {"BRA"}, "kind": {"buy"}, "url": {"https://a.example/x"}},
		"unknown kind":    {"store": {"steam"}, "country": {"BR"}, "kind": {"gift"}, "url": {"https://a.example/x"}},
		"http url":        {"store": {"steam"}, "country": {"BR"}, "kind": {"buy"}, "url": {"http://a.example/x"}},
		"no host":         {"store": {"steam"}, "country": {"BR"}, "kind": {"buy"}, "url": {"https:///x"}},
		"credentials":     {"store": {"steam"}, "country": {"BR"}, "kind": {"buy"}, "url": {"https://user:pw@a.example/x"}},
		"javascript url":  {"store": {"steam"}, "country": {"BR"}, "kind": {"buy"}, "url": {"javascript:alert(1)"}},
	} {
		admin.PostForm(base+"/offers", bad).AssertStatus(303)

		if n := count(`SELECT count(*) FROM offers`); n != 0 {
			t.Fatalf("%s wrote %d rows", name, n)
		}
	}

	admin.PostForm(base+"/offers", good).AssertRedirect(base + "/manage")

	var id, source string
	if err := db.SQL().
		QueryRowContext(t.Context(), `SELECT id::text, source FROM offers WHERE item_id = $1`, itemA).
		Scan(&id, &source); err != nil ||
		source != "admin" {
		t.Fatalf("offer = %q source = %q, %v", id, source, err)
	}

	if got := inertiaProp(t, admin.Get(base+"/manage").AssertOK().Body(), "Admin/ItemManage", "offers"); !strings.Contains(got, "https://store.steampowered.com/app/1") {
		t.Fatalf("offers missing from the manage page: %s", got)
	}

	// the same key again replaces the url and keeps one row
	good.Set("url", "https://store.steampowered.com/app/2")
	admin.PostForm(base+"/offers", good).AssertRedirect(base + "/manage")

	if n := count(
		`SELECT count(*) FROM offers WHERE item_id = $1 AND url = 'https://store.steampowered.com/app/2'`,
		itemA,
	); n != 1 {
		t.Fatalf("rows with the new url = %d", n)
	}

	admin.PostForm("/admin/items/"+itemB+"/offers/"+id+"/toggle", url.Values{}).
		AssertNotFound()
		// not that item's offer
	plain.PostForm(base+"/offers/"+id+"/toggle", url.Values{}).AssertNotFound()
	admin.PostForm(base+"/offers/not-a-uuid/toggle", url.Values{}).AssertNotFound()

	admin.PostForm(base+"/offers/"+id+"/toggle", url.Values{}).AssertRedirect(base + "/manage")

	if n := count(`SELECT count(*) FROM offers WHERE id = $1 AND NOT active`, id); n != 1 {
		t.Fatal("toggle did not deactivate the offer")
	}

	admin.PostForm(base+"/offers/"+id+"/toggle", url.Values{}).AssertRedirect(base + "/manage")

	if n := count(`SELECT count(*) FROM offers WHERE id = $1 AND active`, id); n != 1 {
		t.Fatal("a second toggle did not restore the offer")
	}

	if n := count(`SELECT count(*) FROM admin_audit_logs WHERE resource = 'offers'`); n != 4 {
		t.Fatalf("audit rows = %d, want create, update and two toggles", n)
	}
}
