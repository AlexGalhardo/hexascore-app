//go:build integration

package tests

import (
	"net/http"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

type offerFixture struct {
	app                                *vukemtest.App
	steam, amazon, off, expired, other string
}

func offersApp(t *testing.T, env map[string]string) offerFixture {
	t.Helper()

	env["DATABASE_URL"] = pgtest.Database(t, "")
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: env})

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	exec := func(q string, args ...any) {
		t.Helper()

		if _, err := db.ExecContext(t.Context(), q, args...); err != nil {
			t.Fatalf("%v\n%s", err, q)
		}
	}

	item := uuid.NewV7().String()
	exec(`INSERT INTO items (id, type, slug, wikidata_qid) VALUES ($1, 'book', 'a-book', 'Q1')`, item)

	add := func(country, store, kind, url, extra string) string {
		id := uuid.NewV7().String()
		exec(
			`INSERT INTO offers (id, item_id, country, store, kind, url, source, licence) VALUES ($1, $2, $3, $4, $5, $6, 'admin', 'x')`,
			id,
			item,
			country,
			store,
			kind,
			url,
		)

		if extra != "" {
			exec(`UPDATE offers SET ` + extra + ` WHERE id = '` + id + `'`)
		}

		return id
	}

	return offerFixture{
		app:    app,
		steam:  add("ZZ", "steam", "buy_digital", "https://store.steampowered.com/app/1", ""),
		amazon: add("BR", "amazon_br", "buy", "https://www.amazon.com.br/s?k=9780261103573", ""),
		off:    add("ZZ", "gog", "buy_digital", "https://www.gog.com/en/game/x", "active = false"),
		expired: add(
			"ZZ",
			"epic",
			"buy_digital",
			"https://store.epicgames.com/en-US/p/x",
			"expires_at = now() - interval '1 day'",
		),
		other: add("US", "amazon_us", "buy", "https://www.amazon.com/s?k=9780261103573", ""),
	}
}

func clicks(t *testing.T, f offerFixture, id string) int {
	t.Helper()

	db, _ := f.app.DB()

	var n int
	if err := db.QueryRowContext(t.Context(), `SELECT coalesce(sum(clicks), 0) FROM offer_clicks WHERE offer_id = $1`, id).
		Scan(&n); err != nil {
		t.Fatal(err)
	}

	return n
}

// TestOfferGoRedirectsAndCountsOnce: the stored URL, a 302, one counter step per request, nothing cached or referred.
func TestOfferGoRedirectsAndCountsOnce(t *testing.T) {
	f := offersApp(t, map[string]string{})

	resp := f.app.Client(t).Get("/go/" + f.steam)
	resp.AssertStatus(http.StatusFound).AssertRedirect("https://store.steampowered.com/app/1")

	if got := resp.Header("Cache-Control"); !strings.Contains(got, "no-store") {
		t.Fatalf("Cache-Control = %q", got)
	}

	if got := resp.Header("Referrer-Policy"); got != "no-referrer" {
		t.Fatalf("Referrer-Policy = %q", got)
	}

	if got := resp.Header("X-Robots-Tag"); !strings.Contains(got, "noindex") {
		t.Fatalf("X-Robots-Tag = %q", got)
	}

	if n := clicks(t, f, f.steam); n != 1 {
		t.Fatalf("clicks = %d, want 1", n)
	}

	f.app.Client(t).Get("/go/" + f.steam).AssertStatus(http.StatusFound)

	if n := clicks(t, f, f.steam); n != 2 {
		t.Fatalf("clicks after a second request = %d, want 2", n)
	}
}

// TestSecurity_A10_OfferRedirectOnlyKnownURLs: /go only ever sends the visitor to a URL the catalogue stores for an
// active offer; there is no parameter that names a target.
func TestSecurity_A10_OfferRedirectOnlyKnownURLs(t *testing.T) {
	f := offersApp(t, map[string]string{})
	c := f.app.Client(t)

	for name, path := range map[string]string{
		"deactivated":     "/go/" + f.off,
		"expired":         "/go/" + f.expired,
		"unknown id":      "/go/" + uuid.NewV7().String(),
		"not a uuid":      "/go/not-a-uuid",
		"url as id":       "/go/https:%2F%2Fevil.example",
		"target in query": "/go/" + f.off + "?url=https://evil.example",
	} {
		resp := c.Get(path)
		resp.AssertNotFound()

		if loc := resp.Header("Location"); loc != "" {
			t.Errorf("%s: redirected to %q", name, loc)
		}
	}

	// a query parameter never changes where an active offer goes
	c.Get("/go/" + f.steam + "?url=https://evil.example&next=https://evil.example").
		AssertRedirect("https://store.steampowered.com/app/1")

	for _, id := range []string{f.off, f.expired} {
		if n := clicks(t, f, id); n != 0 {
			t.Errorf("a refused request counted a click on %s", id)
		}
	}
}

// TestOfferGoAddsTheAffiliateTagOnlyWhenConfigured: the stored URL stays clean, the tag is added at redirect time.
func TestOfferGoAddsTheAffiliateTagOnlyWhenConfigured(t *testing.T) {
	plain := offersApp(t, map[string]string{})
	plain.app.Client(t).Get("/go/" + plain.amazon).AssertRedirect("https://www.amazon.com.br/s?k=9780261103573")

	tagged := offersApp(t, map[string]string{"AFFILIATE_AMAZON_BR_TAG": "hexascore-20"})
	tagged.app.Client(t).Get("/go/" + tagged.amazon).
		AssertRedirect("https://www.amazon.com.br/s?k=9780261103573&tag=hexascore-20")
	// another marketplace has no tag of its own: it stays plain
	tagged.app.Client(t).Get("/go/" + tagged.other).AssertRedirect("https://www.amazon.com/s?k=9780261103573")

	unsafe := offersApp(t, map[string]string{"AFFILIATE_AMAZON_BR_TAG": "x&url=https://evil.example"})
	unsafe.app.Client(t).Get("/go/" + unsafe.amazon).AssertRedirect("https://www.amazon.com.br/s?k=9780261103573")
}
