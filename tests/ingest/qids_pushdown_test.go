//go:build integration

package ingest_test

import (
	"context"
	"database/sql"
	"slices"
	"testing"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/igdb"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/openlibrary"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/tmdb"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/trailers"
)

// spy records the arguments of every query a connector runs.
type spy struct {
	database.Executor
	args [][]any
}

func (s *spy) QueryContext(ctx context.Context, q string, args ...any) (*sql.Rows, error) {
	s.args = append(s.args, args)

	return s.Executor.QueryContext(ctx, q, args...)
}

func (s *spy) sawQIDs(want []string) bool {
	for _, a := range s.args {
		for _, v := range a {
			if got, ok := v.([]string); ok && slices.Equal(got, want) {
				return true
			}
		}
	}

	return false
}

// The qids filter reaches each connector's target SQL, and a run with qids yields exactly the requested targets
// of the unrestricted list, in the same order.
func TestConnectorTargetsPushQIDsIntoSQL(t *testing.T) {
	db := newDB(t)
	ctx := t.Context()

	exec(t, db, `INSERT INTO items (id, type, slug, wikidata_qid, external_ids) VALUES
(gen_random_uuid(), 'movie', 'm1', 'Q1', '{"tmdb":"1"}'),
(gen_random_uuid(), 'series', 's2', 'Q2', '{"tmdb":"2"}'),
(gen_random_uuid(), 'game', 'g3', 'Q3', '{"igdb":"g3"}'),
(gen_random_uuid(), 'game', 'g4', 'Q4', '{"igdb":"g4"}'),
(gen_random_uuid(), 'book', 'b5', 'Q5', '{"isbn13":"9780000000001"}'),
(gen_random_uuid(), 'book', 'b6', 'Q6', '{"isbn13":"9780000000002"}')`)
	exec(t, db, `INSERT INTO source_records (source, external_id, payload, hash) VALUES
('wikidata', 'Q1', '{"claims":{"P1651":[{"value":"a","rank":"normal"}]}}', repeat('a', 64)),
('wikidata', 'Q2', '{"claims":{"P1651":[{"value":"b","rank":"normal"}]}}', repeat('b', 64)),
('wikidata', 'Q3', '{"claims":{"P1651":[{"value":"c","rank":"normal"}]}}', repeat('c', 64))`)

	connectors := map[string]struct {
		c    ingest.Connector
		qids []string
		all  []string
	}{
		"tmdb":        {tmdb.New("x"), []string{"Q2", "Q9"}, []string{"Q1", "Q2"}},
		"igdb":        {igdb.New("a", "b"), []string{"Q4"}, []string{"Q3", "Q4"}},
		"trailers":    {&trailers.Connector{}, []string{"Q1", "Q3"}, []string{"Q1", "Q2", "Q3"}},
		"openlibrary": {&openlibrary.Connector{}, []string{"Q6"}, []string{"Q5", "Q6"}},
	}

	ids := func(c ingest.Connector, ex database.Executor, p ingest.RunParams, after string) []string {
		var out []string

		for tg, err := range c.Targets(ctx, ex, p, after) {
			if err != nil {
				t.Fatal(err)
			}

			out = append(out, tg.ID)
		}

		return out
	}

	for name, tc := range connectors {
		t.Run(name, func(t *testing.T) {
			base := ingest.RunParams{Mode: run.ModeFull}

			if got := ids(tc.c, db, base, ""); !slices.Equal(got, tc.all) {
				t.Fatalf("without qids = %v, want %v", got, tc.all)
			}

			p := base
			p.QIDs = tc.qids
			s := &spy{Executor: db}
			got := ids(tc.c, s, p, "")

			var want []string

			for _, id := range tc.all {
				if slices.Contains(tc.qids, id) {
					want = append(want, id)
				}
			}

			if !slices.Equal(got, want) {
				t.Errorf("with qids %v = %v, want %v", tc.qids, got, want)
			}

			if !s.sawQIDs(tc.qids) {
				t.Errorf("qids %v never reached the SQL; args = %v", tc.qids, s.args)
			}

			if got := ids(tc.c, db, p, want[0]); len(want) > 1 && !slices.Equal(got, want[1:]) {
				t.Errorf("after %s = %v, want %v", want[0], got, want[1:])
			}
		})
	}
}
