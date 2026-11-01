//go:build integration

package legacyimport_test

import (
	"context"
	"io"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/database/legacyimport"
)

type statementCounter struct{ n atomic.Int64 }

func (c *statementCounter) ObserveQuery(context.Context, database.Query) { c.n.Add(1) }

// TestEntitiesThroughput compares the per-item reference with the set-based stage on 3,000 items
// (HEXASCORE_BENCH=1; it only reports, the equivalence tests assert). Statements are counted per batch
// of 1,000 items; latency adds one network round trip per statement on a remote link.
func TestEntitiesThroughput(t *testing.T) {
	if os.Getenv("HEXASCORE_BENCH") == "" {
		t.Skip("set HEXASCORE_BENCH=1 to run the benchmark")
	}

	const items, batch = 3000, 1000

	for _, c := range []struct {
		name      string
		reference bool
		workers   int
	}{{"per-item reference", true, 1}, {"set-based, 1 worker", false, 1}, {"set-based, 4 workers", false, 4}} {
		_, db := fixtureDB(t)
		seedSynthetic(t, db, items)

		catalogOnly, err := legacyimport.Select(legacyimport.Stages(), []string{"catalog"})
		if err != nil {
			t.Fatal(err)
		}

		if err := (&legacyimport.Runner{DB: db, Batch: batch, Out: io.Discard, Now: time.Now}).Run(
			t.Context(),
			catalogOnly,
		); err != nil {
			t.Fatal(err)
		}

		var counter statementCounter

		db.SetQueryObserver(&counter)

		stages, _ := legacyimport.Select(legacyimport.Stages(), []string{"entities"})
		if c.reference {
			stages[0].Batch = legacyimport.ReferenceEntitiesBatch()
		}

		start := time.Now()

		r := &legacyimport.Runner{DB: db, Batch: batch, Workers: c.workers, Out: io.Discard, Now: time.Now}
		if err := r.Run(t.Context(), stages); err != nil {
			t.Fatal(err)
		}

		took := time.Since(start)
		stmts := counter.n.Load()
		t.Logf(
			"%-20s %5d items in %7s (%6.0f items/s locally); %5d statements = %.1f per 1,000 items; at 150 ms per statement: %s",
			c.name,
			items,
			took.Round(time.Millisecond),
			items/took.Seconds(),
			stmts,
			float64(stmts)*1000/items,
			(time.Duration(stmts) * 150 * time.Millisecond).Round(time.Second),
		)
	}
}
