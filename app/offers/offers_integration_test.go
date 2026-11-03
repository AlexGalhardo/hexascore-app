//go:build integration

package offers_test

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/offers"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

type rig struct {
	t   testing.TB
	db  *database.DB
	ctx context.Context
	n   int
}

func newRig(tb testing.TB) *rig {
	tb.Helper()

	app := vukemtest.New(
		tb,
		bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(tb, "")}},
	)

	db, err := app.DB()
	if err != nil {
		tb.Fatal(err)
	}

	return &rig{t: tb, db: db, ctx: context.Background()}
}

func (r *rig) exec(q string, args ...any) {
	r.t.Helper()

	if _, err := r.db.ExecContext(r.ctx, q, args...); err != nil {
		r.t.Fatalf("%v\n%s", err, q)
	}
}

func (r *rig) item(typ string) string {
	r.t.Helper()

	r.n++
	id := uuid.NewV7().String()
	r.exec(
		`INSERT INTO items (id, type, slug, wikidata_qid) VALUES ($1, $2, $3, $4)`,
		id,
		typ,
		fmt.Sprintf("i-%d", r.n),
		fmt.Sprintf("Q%d", r.n),
	)

	return id
}

func (r *rig) offer(item, country, store, kind string, extra string) string {
	r.t.Helper()

	id := uuid.NewV7().String()
	r.exec(`INSERT INTO offers (id, item_id, country, store, kind, url, source, licence)
VALUES ($1, $2, $3, $4, $5, $6, 'admin', 'x')`, id, item, country, store, kind, "https://example.test/"+id)

	if extra != "" {
		r.exec(`UPDATE offers SET `+extra+` WHERE id = $1`, id)
	}

	return id
}

func names(vs []offers.View) []string {
	out := make([]string, len(vs))
	for i, v := range vs {
		out[i] = v.Kind + ":" + v.Name
	}

	return out
}

var now = time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)

func TestLoadSection(t *testing.T) {
	r := newRig(t)
	film := r.item("movie")

	r.exec(`INSERT INTO stores (code, name, kind, attribution) VALUES ('tmdb_8', 'Netflix', 'streaming', 'JustWatch'),
('tmdb_2', 'Apple TV', 'streaming', 'JustWatch')`)

	r.offer(film, "BR", "tmdb_8", "stream", "last_verified_at = '2026-10-30'")
	r.offer(film, "BR", "tmdb_2", "buy", "")
	r.offer(film, "US", "tmdb_8", "stream", "")
	r.offer(film, "ZZ", "steam", "buy_digital", "")                                    // valid everywhere
	r.offer(film, "BR", "tmdb_2", "rent", "expires_at = '2026-11-01'")                 // expired
	r.offer(film, "PT", "tmdb_8", "stream", "active = false")                          // deactivated
	r.offer(film, "ES", "tmdb_8", "stream", "verify_status = 'gone', broken_runs = 2") // gone

	got, err := offers.Load(r.ctx, r.db, film, now, offers.EnvTagger(func(string) string { return "" }), "pt-BR")
	if err != nil {
		t.Fatal(err)
	}

	if got.Default != "BR" || !slices.Equal(got.Countries, offers.Countries) {
		t.Fatalf("default=%q countries=%v", got.Default, got.Countries)
	}

	// kinds in display order, the everywhere offer merged into every country
	if want := []string{
		"stream:Netflix",
		"buy:Apple TV",
		"buy_digital:Steam",
	}; !slices.Equal(
		names(got.ByCountry["BR"]),
		want,
	) {
		t.Fatalf("BR = %v, want %v", names(got.ByCountry["BR"]), want)
	}

	if want := []string{"stream:Netflix", "buy_digital:Steam"}; !slices.Equal(names(got.ByCountry["US"]), want) {
		t.Fatalf("US = %v, want %v", names(got.ByCountry["US"]), want)
	}

	for _, c := range []string{"PT", "ES", "MX", "AR"} { // only the everywhere offer is left
		if want := []string{"buy_digital:Steam"}; !slices.Equal(names(got.ByCountry[c]), want) {
			t.Fatalf("%s = %v, want %v", c, names(got.ByCountry[c]), want)
		}
	}

	if got.ByCountry["BR"][0].VerifiedAt != "2026-10-30" || got.ByCountry["BR"][1].VerifiedAt != "" {
		t.Fatalf("verified dates: %+v", got.ByCountry["BR"])
	}

	if !slices.Equal(got.Attribution, []string{"JustWatch"}) {
		t.Fatalf("attribution = %v", got.Attribution)
	}

	if got.Commission || got.AmazonAssociate {
		t.Fatal("no affiliate store is shown: no commission line")
	}
}

func TestLoadBooksShowOneAmazonPerMarketplace(t *testing.T) {
	r := newRig(t)
	book := r.item("book")

	r.exec(
		`INSERT INTO data_sources (code, name, licence, url) VALUES ('wikidata', 'Wikidata', 'CC0-1.0', 'https://www.wikidata.org')`,
	)

	var primary string

	for i, isbn := range []string{"9780261103573", "9780261103580", "9780261103597"} {
		id := uuid.NewV7().String()
		r.exec(
			`INSERT INTO editions (id, work_id, isbn13, title, language, source) VALUES ($1, $2, $3, 'E', 'en', 'wikidata')`,
			id,
			book,
			isbn,
		)

		if i == 1 {
			primary = id
		}

		r.exec(`INSERT INTO offers (item_id, edition_id, country, store, kind, url, source, licence)
VALUES ($1, $2, 'BR', 'amazon_br', 'buy', 'https://www.amazon.com.br/s?k='||$3, 'editions', 'derived')`, book, id, isbn)
	}

	r.exec(`UPDATE items SET primary_edition_id = $2 WHERE id = $1`, book, primary)

	tag := offers.EnvTagger(func(k string) string {
		if k == "AFFILIATE_AMAZON_BR_TAG" {
			return "hexascore-20"
		}

		return ""
	})

	got, err := offers.Load(r.ctx, r.db, book, now, tag, "pt-BR")
	if err != nil {
		t.Fatal(err)
	}

	br := got.ByCountry["BR"]
	if len(br) != 1 || br[0].Name != "Amazon Brasil" || !br[0].Sponsored {
		t.Fatalf("BR = %+v, want a single sponsored Amazon entry", br)
	}

	if !got.Commission || !got.AmazonAssociate {
		t.Fatalf("a tagged Amazon link needs the disclosure: %+v", got)
	}

	// the entry stands for the primary edition
	link, err := offers.Resolve(r.ctx, r.db, br[0].ID, now)
	if err != nil || !strings.HasSuffix(link.URL, "9780261103580") {
		t.Fatalf("link = %+v, %v; want the primary edition's ISBN", link, err)
	}
}

func TestResolveOnlyKnownActiveOffers(t *testing.T) {
	r := newRig(t)
	film := r.item("movie")
	ok := r.offer(film, "ZZ", "steam", "buy_digital", "")
	off := r.offer(film, "ZZ", "gog", "buy_digital", "active = false")
	old := r.offer(film, "ZZ", "epic", "buy_digital", "expires_at = '2026-11-01'")

	if l, err := offers.Resolve(r.ctx, r.db, ok, now); err != nil || l.Store != "steam" || l.Country != "ZZ" {
		t.Fatalf("active offer: %+v, %v", l, err)
	}

	for name, id := range map[string]string{
		"deactivated": off, "expired": old, "unknown": uuid.NewV7().String(), "not a uuid": "../../etc/passwd",
		"empty": "",
	} {
		if _, err := offers.Resolve(r.ctx, r.db, id, now); err == nil {
			t.Errorf("%s offer resolved", name)
		}
	}
}

func TestClickCounterIncrementsPerDayOfferAndCountry(t *testing.T) {
	r := newRig(t)
	film := r.item("movie")
	id := r.offer(film, "BR", "steam", "buy_digital", "")

	for range 3 {
		if err := offers.Click(r.ctx, r.db, id, "BR", now); err != nil {
			t.Fatal(err)
		}
	}

	if err := offers.Click(r.ctx, r.db, id, "BR", now.AddDate(0, 0, 1)); err != nil {
		t.Fatal(err)
	}

	var n, rows int
	if err := r.db.QueryRowContext(r.ctx, `SELECT coalesce(max(clicks) FILTER (WHERE day = '2026-11-03'), 0), count(*) FROM offer_clicks
WHERE offer_id = $1`, id).
		Scan(&n, &rows); err != nil ||
		n != 3 ||
		rows != 2 {
		t.Fatalf("clicks today=%d rows=%d, %v; want 3 and 2 (one row per day)", n, rows, err)
	}
}
