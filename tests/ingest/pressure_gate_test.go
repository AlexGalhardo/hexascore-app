//go:build integration

package ingest_test

import (
	"context"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

type countingGate struct{ waits int }

func (g *countingGate) Wait(context.Context) error { g.waits++; return nil }

// The runner asks the gate before every chunk, so a busy site slows a run without stopping it.
func TestRunnerWaitsOnGateBeforeEachChunk(t *testing.T) {
	db := newDB(t)
	ctx := t.Context()
	rec := &recorder{}

	r, err := run.Create(ctx, db, rec, run.Spec{Job: "fetch.wikidata", Source: "wikidata", Mode: run.ModeFull})
	if err != nil {
		t.Fatal(err)
	}

	gate := &countingGate{}
	runner := newRunner(db, rec, newFlags(t, allOn), run.Limits{Targets: 1000}, map[string]run.Step{
		"fetch.wikidata": upsertStep(3, 1, nil),
	})
	runner.Gate = gate

	if _, err := runner.Execute(ctx, r.ID); err != nil {
		t.Fatal(err)
	}

	if gate.waits != 3 {
		t.Errorf("gate waits = %d, want one per chunk (3)", gate.waits)
	}
}
