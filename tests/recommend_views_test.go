//go:build integration

package tests

import (
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/recommend/primitives"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
	"github.com/AlexGalhardo/hexascore-app/internal/search"
)

type viewsRig struct {
	app  *vukemtest.App
	game string
}

func newViewsRig(t *testing.T) *viewsRig {
	t.Helper()

	app := vukemtest.New(
		t,
		bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}},
	)

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	game := uuid.NewV7().String()
	for _, q := range []string{
		`INSERT INTO data_sources (code, name, licence, url) VALUES ('hexascore', 'x', 'proprietary', 'https://x')`,
	} {
		if _, err := db.ExecContext(t.Context(), q); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := db.ExecContext(t.Context(),
		`INSERT INTO items (id, type, slug, wikidata_qid) VALUES ($1, 'game', 'zelda', 'Q9001')`, game); err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(t.Context(),
		`INSERT INTO item_translations (item_id, locale, title, title_source) VALUES ($1, 'en', 'Zelda', 'hexascore')`,
		game); err != nil {
		t.Fatal(err)
	}

	return &viewsRig{app: app, game: game}
}

// counts flushes the buffer and returns views and search_clicks summed over every day of the item.
func (r *viewsRig) counts(t *testing.T) (views, clicks int) {
	t.Helper()

	if err := primitives.For(r.app.App).Flush(t.Context()); err != nil {
		t.Fatal(err)
	}

	db, err := r.app.DB()
	if err != nil {
		t.Fatal(err)
	}

	if err := db.QueryRowContext(t.Context(),
		`SELECT coalesce(sum(views), 0), coalesce(sum(search_clicks), 0) FROM item_activity_daily WHERE item_id = $1`,
		r.game).Scan(&views, &clicks); err != nil {
		t.Fatal(err)
	}

	return views, clicks
}

func TestViewsDedupPerVisitorPerDay(t *testing.T) {
	r := newViewsRig(t)
	client := r.app.Client(t)

	for range 5 {
		client.Inertia("").Get("/games/zelda").AssertOK()
	}

	// Same visitor through the locale prefix and a missing page: no extra views.
	client.Inertia("").Get("/pt-br/games/zelda").AssertOK()
	client.Inertia("").Get("/games/missing").AssertStatus(301)

	if v, _ := r.counts(t); v != 1 {
		t.Fatalf("views = %d after 5 GETs by one visitor, want 1", v)
	}

	r.app.Client(t).WithCookie("vukem_session", "another-visitor").Inertia("").Get("/games/zelda").AssertOK()

	if v, _ := r.counts(t); v != 2 {
		t.Fatalf("views = %d after a second visitor, want 2", v)
	}
}

func TestViewsExclusions(t *testing.T) {
	r := newViewsRig(t)

	for name, hdr := range map[string][2]string{
		"prefetch": {"Sec-Purpose", "prefetch"},
		"partial":  {"X-Inertia-Partial-Data", "item"},
		"bot":      {"User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1)"},
	} {
		r.app.Client(t).WithHeader(hdr[0], hdr[1]).Inertia("").Get("/games/zelda").AssertOK()

		if v, _ := r.counts(t); v != 0 {
			t.Fatalf("%s request counted %d views", name, v)
		}
	}
}

func TestAppReadsDedupPerToken(t *testing.T) {
	r := newAPICatalog(t)
	db, err := r.app.DB()
	if err != nil {
		t.Fatal(err)
	}

	first := appCatalogClient(t, r, "catalog:read")
	for range 5 {
		first.GetJSON("/app/v1/games/" + r.game).AssertOK()
	}

	rec := primitives.For(r.app.App)
	if err := rec.Flush(t.Context()); err != nil {
		t.Fatal(err)
	}

	var views int
	if err := db.QueryRowContext(t.Context(),
		`SELECT coalesce(sum(views), 0) FROM item_activity_daily WHERE item_id = $1`, r.game).Scan(&views); err != nil {
		t.Fatal(err)
	}

	if views != 1 {
		t.Fatalf("views = %d after 5 reads with one token, want 1", views)
	}
}

func TestSearchClickCounted(t *testing.T) {
	r := newViewsRig(t)

	ev := search.ResultClicked{ItemID: r.game, Position: 1, At: time.Now()}
	if err := r.app.Events.Dispatch(t.Context(), ev); err != nil {
		t.Fatal(err)
	}

	if _, c := r.counts(t); c != 1 {
		t.Fatalf("search_clicks = %d, want 1", c)
	}
}

func TestNoRawIPStored(t *testing.T) {
	r := newViewsRig(t)
	r.app.Client(t).Inertia("").Get("/games/zelda").AssertOK()
	r.counts(t)

	db, err := r.app.DB()
	if err != nil {
		t.Fatal(err)
	}

	var hits int
	if err := db.QueryRowContext(
		t.Context(),
		`SELECT count(*) FROM item_activity_daily a WHERE row_to_json(a)::text LIKE '%127.0.0.1%'`,
	).Scan(&hits); err != nil {
		t.Fatal(err)
	}

	if hits != 0 {
		t.Fatal("an IP reached item_activity_daily")
	}
}

// TestFlushIdempotentOnFailure: a failed flush keeps the counts for the next one, which writes them once.
func TestFlushIdempotentOnFailure(t *testing.T) {
	r := newViewsRig(t)
	r.app.Client(t).Inertia("").Get("/games/zelda").AssertOK()

	db, err := r.app.DB()
	if err != nil {
		t.Fatal(err)
	}

	rec := primitives.For(r.app.App)

	exec := func(q string) {
		t.Helper()

		if _, err := db.ExecContext(t.Context(), q); err != nil {
			t.Fatal(err)
		}
	}

	exec(`ALTER TABLE item_activity_daily RENAME TO item_activity_daily_off`)

	if err := rec.Flush(t.Context()); err == nil {
		t.Fatal("flush into a missing table succeeded")
	}

	exec(`ALTER TABLE item_activity_daily_off RENAME TO item_activity_daily`)

	if v, _ := r.counts(t); v != 1 {
		t.Fatalf("views = %d after a failed then a good flush, want 1", v)
	}

	if v, _ := r.counts(t); v != 1 {
		t.Fatalf("views = %d after an empty flush, want still 1", v)
	}
}
