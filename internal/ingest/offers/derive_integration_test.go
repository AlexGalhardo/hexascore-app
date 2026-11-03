//go:build integration

package offers_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/offers"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

type rig struct {
	t   testing.TB
	db  *database.DB
	ctx context.Context
	n   int
}

func newRig(tb testing.TB) *rig {
	tb.Helper()

	url := pgtest.Database(tb, "")
	app := vukemtest.New(tb, bootstrap.New, vukemtest.Options{Env: map[string]string{"DATABASE_URL": url}})

	db, err := app.DB()
	if err != nil {
		tb.Fatal(err)
	}

	r := &rig{t: tb, db: db, ctx: context.Background()}
	r.exec(
		`INSERT INTO data_sources (code, name, licence, url) VALUES ('wikidata', 'Wikidata', 'CC0-1.0', 'https://www.wikidata.org')`,
	)

	return r
}

func (r *rig) exec(q string, args ...any) {
	r.t.Helper()

	if _, err := r.db.ExecContext(r.ctx, q, args...); err != nil {
		r.t.Fatalf("%v\n%s", err, q)
	}
}

func (r *rig) item(typ, visibility, externalIDs string) string {
	r.t.Helper()

	r.n++
	id := uuid.NewV7().String()
	r.exec(
		`INSERT INTO items (id, type, slug, wikidata_qid, visibility, external_ids) VALUES ($1, $2, $3, $4, $5, $6::jsonb)`,
		id,
		typ,
		fmt.Sprintf("item-%d", r.n),
		fmt.Sprintf("Q%d", r.n),
		visibility,
		externalIDs,
	)

	return id
}

func (r *rig) edition(work, isbn string) string {
	r.t.Helper()

	r.n++
	id := uuid.NewV7().String()
	r.exec(
		`INSERT INTO editions (id, work_id, isbn13, title, language, source) VALUES ($1, $2, $3, 'Edition', 'en', 'wikidata')`,
		id,
		work,
		isbn,
	)

	return id
}

// derive runs the step to completion in pages of batch items and returns how many rows it changed.
func (r *rig) derive(batch int) int {
	r.t.Helper()

	var (
		cur     []byte
		changed int
	)

	for range 100 {
		var res run.Result

		err := r.db.Tx(r.ctx, func(tx *database.Tx) (err error) {
			res, err = offers.Derive{Batch: batch}.Step(r.ctx, tx, run.Run{Cursor: cur})

			return err
		})
		if err != nil {
			r.t.Fatal(err)
		}

		changed += res.Changed
		cur = res.Cursor

		if res.Done {
			return changed
		}
	}

	r.t.Fatal("derive did not finish")

	return 0
}

// all lists "store kind country url" per offer, sorted, for one item.
func (r *rig) all(item string) []string {
	r.t.Helper()

	rows, err := r.db.QueryContext(r.ctx, `SELECT store, kind, country, url, source FROM offers WHERE item_id = $1
ORDER BY store, kind, country, url`, item)
	if err != nil {
		r.t.Fatal(err)
	}

	defer func() { _ = rows.Close() }()

	var out []string

	for rows.Next() {
		var store, kind, country, url, source string
		if err := rows.Scan(&store, &kind, &country, &url, &source); err != nil {
			r.t.Fatal(err)
		}

		out = append(out, strings.Join([]string{store, kind, country, url, source}, " "))
	}

	if err := rows.Err(); err != nil {
		r.t.Fatal(err)
	}

	return out
}

func TestDeriveStoreOffers(t *testing.T) {
	r := newRig(t)
	game := r.item("game", "public", `{"steam":"292030","gog":"game/the_witcher_3_wild_hunt","epic":"fortnite",
"google_play":"com.supercell.clashofclans","app_store":"529479190"}`)

	r.derive(500)

	want := []string{
		"app_store mobile ZZ https://apps.apple.com/app/id529479190 wikidata",
		"epic buy_digital ZZ https://store.epicgames.com/en-US/p/fortnite wikidata",
		"gog buy_digital ZZ https://www.gog.com/en/game/the_witcher_3_wild_hunt wikidata",
		"google_play mobile ZZ https://play.google.com/store/apps/details?id=com.supercell.clashofclans wikidata",
		"steam buy_digital ZZ https://store.steampowered.com/app/292030 wikidata",
	}

	got := r.all(game)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("offers =\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

// An id from an external source never becomes part of a URL unless it matches the shape of that store's ids.
func TestDeriveRefusesHostileIDs(t *testing.T) {
	r := newRig(t)
	game := r.item("game", "public", `{"steam":"292030/../../evil","gog":"game/../../evil","epic":"a b",
"google_play":"com.x?y=1#z","app_store":"12 34"}`)

	r.derive(500)

	if got := r.all(game); len(got) != 0 {
		t.Fatalf("hostile ids produced offers: %v", got)
	}
}

func TestDeriveSkipsHiddenItemsAndKeepsAdminOffers(t *testing.T) {
	r := newRig(t)
	hidden := r.item("game", "hidden", `{"steam":"1"}`)
	game := r.item("game", "public", `{"steam":"292030"}`)

	r.exec(`INSERT INTO offers (item_id, country, store, kind, url, source, licence)
VALUES ($1, 'ZZ', 'steam', 'buy_digital', 'https://store.steampowered.com/app/292030?curator=1', 'admin', 'proprietary')`, game)

	r.derive(500)

	if got := r.all(hidden); len(got) != 0 {
		t.Fatalf("hidden item got offers: %v", got)
	}

	got := r.all(game)
	if len(got) != 1 || !strings.Contains(got[0], "curator=1") || !strings.HasSuffix(got[0], " admin") {
		t.Fatalf("the admin offer must stay as entered: %v", got)
	}
}

func TestDeriveUpdatesItsOwnOffersAndLeavesDeactivatedOnes(t *testing.T) {
	r := newRig(t)
	game := r.item("game", "public", `{"steam":"111"}`)

	r.derive(500)

	r.exec(`UPDATE offers SET active = false WHERE item_id = $1`, game)
	r.exec(`UPDATE items SET external_ids = '{"steam":"222"}' WHERE id = $1`, game)

	if changed := r.derive(500); changed != 1 {
		t.Fatalf("changed = %d, want the one offer whose id moved", changed)
	}

	var url string

	var active bool
	if err := r.db.QueryRowContext(r.ctx, `SELECT url, active FROM offers WHERE item_id = $1`, game).
		Scan(&url, &active); err != nil {
		t.Fatal(err)
	}

	if url != "https://store.steampowered.com/app/222" || active {
		t.Fatalf("url=%s active=%v: the id update must apply and an admin's deactivation must stay", url, active)
	}
}

func TestDeriveIsIdempotentAndResumable(t *testing.T) {
	r := newRig(t)

	for i := range 5 {
		r.item("game", "public", fmt.Sprintf(`{"steam":"%d"}`, 1000+i))
	}

	if changed := r.derive(2); changed != 5 { // pages of two items
		t.Fatalf("first pass changed %d, want 5", changed)
	}

	if changed := r.derive(2); changed != 0 {
		t.Fatalf("second pass changed %d, want 0", changed)
	}
}

func TestDeriveBookOffersPerEdition(t *testing.T) {
	r := newRig(t)
	book := r.item("book", "public", `{}`)
	e1 := r.edition(book, "9780261103573")
	r.edition(book, "9788533613379")
	r.exec(`UPDATE editions SET isbn13 = NULL, wikidata_qid = 'Q999' WHERE id = $1`, e1) // no isbn: no offer

	r.edition(book, "9780261103580")

	hiddenBook := r.item("book", "hidden", `{}`)
	r.edition(hiddenBook, "9780261103597")

	r.derive(500)

	got := r.all(book)
	want := []string{
		"amazon_br buy BR https://www.amazon.com.br/s?k=9780261103580 editions",
		"amazon_br buy BR https://www.amazon.com.br/s?k=9788533613379 editions",
		"amazon_us buy US https://www.amazon.com/s?k=9780261103580 editions",
		"amazon_us buy US https://www.amazon.com/s?k=9788533613379 editions",
	}

	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("offers =\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}

	if hidden := r.all(hiddenBook); len(hidden) != 0 {
		t.Fatalf("hidden book got offers: %v", hidden)
	}
}
