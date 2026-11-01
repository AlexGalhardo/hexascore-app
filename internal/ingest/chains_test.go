package ingest

import (
	"encoding/json"
	"testing"
	"time"
)

func sinceOf(t *testing.T, raw json.RawMessage) (time.Time, bool) {
	t.Helper()

	var p map[string]json.RawMessage
	if err := json.Unmarshal(raw, &p); err != nil {
		t.Fatal(err)
	}

	v, ok := p["since"]
	if !ok {
		return time.Time{}, false
	}

	var ts time.Time
	if err := json.Unmarshal(v, &ts); err != nil {
		t.Fatal(err)
	}

	return ts, true
}

func TestIncrementalStagePersistsSince(t *testing.T) {
	before := time.Now()

	spec, err := WikidataIncremental.spec(0, nil, "test")
	if err != nil {
		t.Fatal(err)
	}

	after := time.Now()

	got, ok := sinceOf(t, spec.Params)
	if !ok {
		t.Fatal("incremental stage params carry no since")
	}

	if got.Before(before.Add(-IncrementalWindow).Add(-time.Second)) || got.After(after.Add(-IncrementalWindow)) {
		t.Fatalf("since = %s, want about now-%s", got, IncrementalWindow)
	}
}

func TestExplicitSinceIsKept(t *testing.T) {
	want := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	raw, _ := json.Marshal(want)

	spec, err := WikidataIncremental.spec(0, map[string]json.RawMessage{"since": raw}, "test")
	if err != nil {
		t.Fatal(err)
	}

	if got, _ := sinceOf(t, spec.Params); !got.Equal(want) {
		t.Fatalf("since = %s, want %s", got, want)
	}
}

func TestSuccessorStageReusesSince(t *testing.T) {
	first, err := WikidataIncremental.spec(0, nil, "test")
	if err != nil {
		t.Fatal(err)
	}

	var params map[string]json.RawMessage
	if err := json.Unmarshal(first.Params, &params); err != nil {
		t.Fatal(err)
	}

	next, err := WikidataIncremental.spec(1, params, "chain:1")
	if err != nil {
		t.Fatal(err)
	}

	a, _ := sinceOf(t, first.Params)
	if b, _ := sinceOf(t, next.Params); !a.Equal(b) {
		t.Fatalf("successor since %s differs from first %s", b, a)
	}
}

func TestNonIncrementalStageHasNoSince(t *testing.T) {
	spec, err := WikidataWeekly.spec(0, nil, "test")
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := sinceOf(t, spec.Params); ok {
		t.Fatal("full-mode stage got a since")
	}
}
