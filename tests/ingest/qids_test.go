//go:build integration

package ingest_test

import (
	"context"
	"encoding/json"
	"iter"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/fetch"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/wikidata"
)

// listConnector yields Q1..Q5 and records what Fetch received.
type listConnector struct{ fetched *[]string }

func (listConnector) Source() string { return "wikidata" }

func (listConnector) Targets(
	_ context.Context, _ database.Executor, _ ingest.RunParams, after string,
) iter.Seq2[ingest.Target, error] {
	return func(yield func(ingest.Target, error) bool) {
		for _, id := range []string{"Q1", "Q2", "Q3", "Q4", "Q5"} {
			if id > after && !yield(ingest.Target{ID: id, Type: "movie"}, nil) {
				return
			}
		}
	}
}

func (c listConnector) Fetch(
	_ context.Context, _ *ingest.Client, batch []ingest.Target,
) ([]ingest.Record, []ingest.TargetError) {
	var recs []ingest.Record

	for _, t := range batch {
		*c.fetched = append(*c.fetched, t.ID)
		recs = append(recs, ingest.Record{Source: "wikidata", ExternalID: t.ID, Payload: json.RawMessage(`{"n":1}`)})
	}

	return recs, nil
}

// The qids param of a run limits it to those QIDs, whatever the connector would have yielded, and a limit does not
// cut the list short.
func TestRunQIDsParamRestrictsTargets(t *testing.T) {
	db := newDB(t)
	ctx := t.Context()

	r, err := run.Create(ctx, db, &recorder{}, run.Spec{
		Job: "fetch.wikidata", Source: "wikidata", Mode: run.ModeManual,
		Params: json.RawMessage(`{"qids":["Q2","Q4","Q9"],"limit":1}`),
	})
	if err != nil {
		t.Fatal(err)
	}

	var fetched []string

	step := fetch.Step(listConnector{&fetched}, nil, time.Now)

	var res run.Result

	err = db.Tx(ctx, func(tx *database.Tx) error {
		var serr error

		res, serr = step(ctx, tx, r)

		return serr
	})
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(fetched, []string{"Q2", "Q4"}) {
		t.Errorf("fetched = %v, want [Q2 Q4]", fetched)
	}

	if !res.Done {
		t.Error("run with qids must be done after one pass")
	}
}

// Without qids nothing changes.
func TestRunWithoutQIDsFetchesEverything(t *testing.T) {
	db := newDB(t)
	ctx := t.Context()

	r, err := run.Create(ctx, db, &recorder{}, run.Spec{Job: "fetch.wikidata", Source: "wikidata", Mode: run.ModeFull})
	if err != nil {
		t.Fatal(err)
	}

	var fetched []string

	step := fetch.Step(listConnector{&fetched}, nil, time.Now)

	err = db.Tx(ctx, func(tx *database.Tx) error {
		_, serr := step(ctx, tx, r)

		return serr
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(fetched) != 5 {
		t.Errorf("fetched = %v, want 5 targets", fetched)
	}
}

// The Wikidata connector does not run its SPARQL selection for a qids run: it takes the item type from the stored
// record, or from the single requested type, and refuses a QID whose type it cannot know.
func TestWikidataQIDsSkipSelection(t *testing.T) {
	db := newDB(t)
	ctx := t.Context()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("the SPARQL endpoint must not be called for a qids run")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	if _, err := db.ExecContext(ctx, `INSERT INTO source_records (source, external_id, payload, hash)
VALUES ('wikidata', 'Q7', '{"type":"book"}'::jsonb, repeat('a', 64))`); err != nil {
		t.Fatal(err)
	}

	w := wikidata.New(nil)
	w.SparqlURL = srv.URL

	collect := func(p ingest.RunParams, after string) ([]string, error) {
		var out []string

		for tg, err := range w.Targets(ctx, db, p, after) {
			if err != nil {
				return out, err
			}

			out = append(out, tg.ID+"="+tg.Type)
		}

		return out, nil
	}

	got, err := collect(ingest.RunParams{Mode: run.ModeManual, QIDs: []string{"Q8", "Q7"}, Types: []string{"game"}}, "")
	if err != nil || !slices.Equal(got, []string{"Q7=book", "Q8=game"}) {
		t.Errorf("targets = %v, %v; want [Q7=book Q8=game]", got, err)
	}

	got, err = collect(
		ingest.RunParams{Mode: run.ModeManual, QIDs: []string{"Q8", "Q7"}, Types: []string{"game"}},
		"Q7",
	)
	if err != nil || !slices.Equal(got, []string{"Q8=game"}) {
		t.Errorf("after Q7 = %v, %v; want [Q8=game]", got, err)
	}

	if _, err := collect(ingest.RunParams{Mode: run.ModeManual, QIDs: []string{"Q8"}}, ""); err == nil {
		t.Error("a QID without a known type must be refused")
	}
}
