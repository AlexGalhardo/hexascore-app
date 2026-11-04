//go:build integration

package normalize_test

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/normalize"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/store"
)

// A per-category lane normalises only the QIDs whose Wikidata record is of its type; without types every pending QID
// is taken, as before.
func TestPendingFiltersByCategory(t *testing.T) {
	ctx := context.Background()
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{
		Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")},
	})

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(ctx, `INSERT INTO data_sources (code, name, licence, url)
VALUES ('wikidata', 'Wikidata', 'CC0-1.0', 'https://www.wikidata.org')`); err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(ctx, `INSERT INTO job_runs (id, job, mode, status, triggered_by) OVERRIDING SYSTEM VALUE
VALUES (1, 'fetch.wikidata', 'full', 'running', 't')`); err != nil {
		t.Fatal(err)
	}

	for qid, typ := range map[string]string{"Q1": "movie", "Q2": "series", "Q3": "movie", "Q4": "book"} {
		if err := db.Tx(ctx, func(tx *database.Tx) error {
			_, err := store.Upsert(ctx, tx, 1, store.Record{
				Source: "wikidata", ExternalID: qid, Payload: json.RawMessage(`{"type":"` + typ + `"}`),
			})

			return err
		}); err != nil {
			t.Fatal(err)
		}
	}

	read := func(types []string) []string {
		var out []string

		if err := db.Tx(ctx, func(tx *database.Tx) (err error) {
			out, err = normalize.Pending(ctx, tx, 100, types)

			return err
		}); err != nil {
			t.Fatal(err)
		}

		return out
	}

	if got := read([]string{"movie"}); !slices.Equal(got, []string{"Q1", "Q3"}) {
		t.Errorf("movie lane = %v, want [Q1 Q3]", got)
	}

	if got := read(nil); len(got) != 4 {
		t.Errorf("no lane = %v, want all four", got)
	}
}
