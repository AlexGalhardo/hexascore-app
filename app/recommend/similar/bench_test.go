//go:build integration

package similar_test

import (
	"os"
	"strconv"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/app/recommend/similar"
)

// BenchmarkComputeSimilar seeds SIMILAR_BENCH_ITEMS games (default 5000) with a franchise, developer, genres and
// cast drawn from skewed pools, then times one full run. The SDD 08 issue 11 figure needs SIMILAR_BENCH_ITEMS=500000.
func BenchmarkComputeSimilar(b *testing.B) {
	n := 5000
	if s := os.Getenv("SIMILAR_BENCH_ITEMS"); s != "" {
		n, _ = strconv.Atoi(s)
	}

	r := newRig(b)
	r.exec(`INSERT INTO entities (id, kind, slug, default_name)
SELECT gen_random_uuid(), 'person', 'p-' || g, 'p' FROM generate_series(1, $1::int / 5) g`, n)
	r.exec(`CREATE TEMP TABLE pool AS SELECT id, row_number() OVER () AS k FROM entities`)
	r.exec(`ANALYZE entities`)
	r.exec(`INSERT INTO items (id, type, slug, wikidata_qid)
SELECT gen_random_uuid(), 'game', 'b-' || g, 'Q' || (1000000 + g) FROM generate_series(1, $1::int) g`, n)
	r.exec(`ANALYZE items`)
	r.exec(`INSERT INTO item_entities (item_id, entity_id, role, position, source)
SELECT i.id, p.id, CASE WHEN c = 0 THEN 'developer' ELSE 'cast' END, c, 'hexascore'
FROM (SELECT id, row_number() OVER () AS k FROM items WHERE slug LIKE 'b-%') i
CROSS JOIN generate_series(0, 3) c
JOIN pool p ON p.k = 1 + ((i.k * (c + 7)) % (SELECT count(*) FROM pool))
ON CONFLICT DO NOTHING`)
	// Autovacuum would have analysed these tables long before the nightly run; without it the first plans are quadratic.
	r.exec(`ANALYZE`)

	for b.Loop() {
		r.exec(`DELETE FROM item_similar`)

		if _, err := similar.Compute(r.ctx, r.db, similar.Options{Pause: -1, Now: now}); err != nil {
			b.Fatal(err)
		}
	}
}
