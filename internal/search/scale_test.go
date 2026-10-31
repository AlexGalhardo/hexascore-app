//go:build integration

package search_test

import (
	"context"
	"errors"
	"os"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/AlexGalhardo/hexascore-app/internal/search"
)

// TestSearchScaleP95 reproduces the production failure (every /search answered 503 "search timeout" on a catalogue
// of about 160k items) and keeps p95 <= 500 ms on a deliberately skewed word distribution (NFR-07-1 asks 300 ms on real data; the timeout is 800 ms). It seeds SEARCH_SCALE_ITEMS items (about 1.8
// translations and 0.33 alternative titles each, hash-made words so trigram sets look like real titles) and is
// skipped without the variable: SEARCH_SCALE_ITEMS=160000 go test -tags integration ./internal/search -run Scale.
func TestSearchScaleP95(t *testing.T) {
	n, _ := strconv.Atoi(os.Getenv("SEARCH_SCALE_ITEMS"))
	if n == 0 {
		t.Skip("set SEARCH_SCALE_ITEMS (160000 = production size)")
	}

	db := migratedDB(t)
	exec(
		t,
		db,
		`INSERT INTO data_sources (code, name, licence, url) VALUES ('hexascore', 'x', 'proprietary', 'https://x')`,
	)
	exec(t, db, `INSERT INTO items (id, type, slug, wikidata_qid, popularity_prior)
		SELECT gen_random_uuid(), (ARRAY['game','movie','series','book'])[1 + g % 4], 'item-' || g, 'Q' || (100000 + g),
		  (random() ^ 4) * 200 FROM generate_series(1, $1) g`, n)
	// The search_dirty trigger would queue every row for the suggest sync, which this test does not exercise.
	exec(t, db, `ALTER TABLE item_translations DISABLE TRIGGER item_translations_search_dirty_trg`)
	exec(t, db, `ALTER TABLE item_alt_titles DISABLE TRIGGER item_alt_titles_search_dirty_trg`)
	exec(t, db, `CREATE TEMP TABLE vocab AS
		SELECT g AS id, substr(translate(md5('w' || g), '0123456789', 'ghijklmnop'), 1, 4 + g % 6) AS w
		FROM generate_series(1, 60000) g`)
	exec(t, db, `CREATE TEMP TABLE pick AS
		SELECT i.id AS item_id, l.loc, k, 1 + floor(power(random(), 2) * 60000)::int AS vid
		FROM (SELECT id, row_number() OVER () rn FROM items) i
		JOIN (VALUES ('en', 0), ('pt-BR', 1), ('es', 2)) l(loc, n)
		  ON (l.n = 0 OR (l.n = 1 AND i.rn % 2 = 0) OR (l.n = 2 AND i.rn % 10 < 3))
		JOIN LATERAL generate_series(1, 2 + (i.rn + l.n) % 3) k ON true`)
	exec(t, db, `INSERT INTO item_translations (item_id, locale, title, title_source)
		SELECT p.item_id, p.loc, string_agg(initcap(v.w), ' ' ORDER BY p.k), 'hexascore'
		FROM pick p JOIN vocab v ON v.id = p.vid GROUP BY p.item_id, p.loc`)
	exec(t, db, `INSERT INTO item_alt_titles (item_id, title, kind, source)
		SELECT i.id, 'Alt ' || initcap(v.w) || ' ' || i.rn, 'alias', 'hexascore'
		FROM (SELECT id, row_number() OVER () rn FROM items) i JOIN vocab v ON v.id = 1 + i.rn % 60000
		WHERE i.rn % 3 = 0`)
	exec(t, db, `ANALYZE`)

	var words []string

	rows, err := db.QueryContext(context.Background(),
		`SELECT lower(split_part(title, ' ', 1)) FROM item_translations WHERE locale = 'en' ORDER BY item_id LIMIT 40`)
	if err != nil {
		t.Fatal(err)
	}

	for rows.Next() {
		var w string
		if err := rows.Scan(&w); err != nil {
			t.Fatal(err)
		}

		words = append(words, w)
	}

	_ = rows.Close()

	s := searcher(db)
	// The first call also computes the popularity p99, which is cached for 24 h.
	_, _ = s.Search(context.Background(), search.Query{Text: words[0], Audience: search.AudienceSite})
	took := make([]time.Duration, 0, len(words))

	for _, w := range words {
		start := time.Now()
		_, err := s.Search(context.Background(), search.Query{Text: w, Audience: search.AudienceSite})

		if errors.Is(err, search.ErrTimeout) {
			t.Fatalf("search %q timed out at %d items", w, n)
		}

		if err != nil {
			t.Fatal(err)
		}

		took = append(took, time.Since(start))
	}

	slices.Sort(took)

	p95 := took[len(took)*95/100]
	t.Logf("%d items: p50 %v p95 %v max %v", n, took[len(took)/2], p95, took[len(took)-1])

	if p95 > 500*time.Millisecond {
		t.Fatalf("p95 %v is too close to the 800 ms statement timeout", p95)
	}
}
