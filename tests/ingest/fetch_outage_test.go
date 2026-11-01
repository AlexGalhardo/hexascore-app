//go:build integration

package ingest_test

import (
	"context"
	"errors"
	"iter"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/httpclient"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/fetch"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

type outageConnector struct{ err error }

func (outageConnector) Source() string { return "wikidata" }

func (outageConnector) Targets(
	context.Context,
	database.Executor,
	ingest.RunParams,
	string,
) iter.Seq2[ingest.Target, error] {
	return func(yield func(ingest.Target, error) bool) {
		for _, id := range []string{"Q1", "Q2"} {
			if !yield(ingest.Target{ID: id, Type: "movie"}, nil) {
				return
			}
		}
	}
}

func (c outageConnector) Fetch(
	_ context.Context,
	_ *ingest.Client,
	batch []ingest.Target,
) ([]ingest.Record, []ingest.TargetError) {
	errs := make([]ingest.TargetError, 0, len(batch))
	for _, t := range batch {
		errs = append(errs, ingest.TargetError{Target: t, Err: c.err})
	}

	return nil, errs
}

// An open breaker says nothing about the targets: the chunk fails instead of filling the dead-letter table.
func TestFetchOutageIsNotATargetFailure(t *testing.T) {
	db := newDB(t)
	ctx := t.Context()

	r, err := run.Create(
		ctx,
		db,
		&recorder{},
		run.Spec{Job: "fetch.wikidata", Source: "wikidata", Mode: run.ModeBootstrap},
	)
	if err != nil {
		t.Fatal(err)
	}

	step := fetch.Step(outageConnector{err: httpclient.ErrCircuitOpen}, nil, time.Now)

	err = db.Tx(ctx, func(tx *database.Tx) error {
		_, err := step(ctx, tx, r)

		return err
	})
	if !errors.Is(err, httpclient.ErrCircuitOpen) {
		t.Fatalf("step err = %v, want the circuit error", err)
	}

	if n := count(t, db, `SELECT count(*) FROM ingest_failures`); n != 0 {
		t.Fatalf("ingest_failures = %d, want 0", n)
	}
}
