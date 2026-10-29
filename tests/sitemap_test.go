//go:build integration

package tests

import (
	"encoding/xml"
	"fmt"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func TestSitemaps(t *testing.T) {
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{
		"DATABASE_URL": pgtest.Database(t, ""),
		"APP_URL":      "https://example.test/",
	}})

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	exec := func(q string, args ...any) {
		t.Helper()

		if _, err := db.ExecContext(t.Context(), q, args...); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}

	exec(
		`INSERT INTO data_sources (code, name, licence, url) VALUES ('wikidata', 'Wikidata', 'CC0-1.0', 'https://www.wikidata.org')`,
	)

	studio3, studio2 := uuid.NewV7().String(), uuid.NewV7().String()
	exec(
		`INSERT INTO entities (id, kind, slug, default_name) VALUES ($1,'organization','big-studio','Big'), ($2,'organization','small-studio','Small')`,
		studio3,
		studio2,
	)

	link := func(item, studio string) {
		exec(
			`INSERT INTO item_entities (item_id, entity_id, role, source) VALUES ($1,$2,'developer','wikidata')`,
			item,
			studio,
		)
	}

	for i, vis := range []string{"public", "public", "public", "hidden", "takedown"} {
		id := uuid.NewV7().String()
		exec(`INSERT INTO items (id, type, slug, wikidata_qid, visibility) VALUES ($1,'game',$2,$3,$4)`,
			id, fmt.Sprintf("game-%d", i), fmt.Sprintf("Q%d", i+1), vis)
		// Hidden and takedown links must not count towards the threshold.
		link(id, studio3)

		if i < 2 || i == 3 {
			link(id, studio2)
		}
	}

	index := app.Client(t).Get("/sitemap.xml").AssertOK().AssertHeaderContains("Content-Type", "application/xml").
		AssertHeaderContains("Cache-Control", "max-age=3600").AssertSee("https://example.test/sitemaps/static.xml").
		AssertSee("https://example.test/sitemaps/items-game-1.xml").AssertSee("entities-organization-1.xml").
		AssertDontSee("items-movie").AssertDontSee("entities-person")
	index.AssertDontSee("example.test//")

	items := app.Client(t).Get("/sitemaps/items-game-1.xml").AssertOK()
	if n := strings.Count(items.Body(), "<url>"); n != 3 {
		t.Fatalf("game urls = %d, want only the 3 public items", n)
	}

	items.AssertSee(`<loc>https://example.test/games/game-0</loc>`).
		AssertSee(`hreflang="pt-br" href="https://example.test/pt-br/games/game-0"`).
		AssertSee(`hreflang="es" href="https://example.test/es/games/game-0"`).
		AssertSee(`hreflang="x-default" href="https://example.test/games/game-0"`).
		AssertDontSee("game-3").AssertDontSee("game-4")

	// Entities need 3 public linked items: hidden and takedown links do not count.
	ent := app.Client(t).Get("/sitemaps/entities-organization-1.xml").AssertOK()
	ent.AssertSee("/companies/big-studio")
	ent.AssertDontSee("small-studio") // 3 links, but one is hidden
}

func TestSitemapsWellFormedAndMissing(t *testing.T) {
	app := vukemtest.New(
		t,
		bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}},
	)

	static := app.Client(t).Get("/sitemaps/static.xml").AssertOK().Body()
	if err := xml.Unmarshal([]byte(static), new(struct{ XMLName xml.Name })); err != nil {
		t.Fatalf("static sitemap is not well-formed: %v", err)
	}

	for _, p := range []string{"/sitemaps/items-game-1.xml", "/sitemaps/items-game-2.xml", "/sitemaps/items-bogus-1.xml", "/sitemaps/x.xml", "/sitemaps/items-game-0.xml"} {
		app.Client(t).Get(p).AssertStatus(404)
	}
}
