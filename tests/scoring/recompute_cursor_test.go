//go:build integration

package scoring_test

import (
	"testing"
	"time"

	"github.com/AlexGalhardo/hexascore-app/app/scoring"
)

func (r *rig) seedCountedItems(n int) {
	r.t.Helper()
	r.exec(`INSERT INTO items (id, type, slug, wikidata_qid, release_precision, release_status, rating_count)
SELECT gen_random_uuid(), 'game', 'cur-' || g, 'Q9' || g, 'day', 'released', 1 FROM generate_series(1, $1::int) g`, n)
}

func TestRecomputeAll_ResumesFromCursor(t *testing.T) {
	r := newRig(t)
	r.seedCountedItems(1200)
	db, err := r.app.DB()
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()

	p, err := scoring.RecomputeChunk(t.Context(), db, now, time.Nanosecond)
	if err != nil || p.Done || p.Items != scoring.BatchSize {
		t.Fatalf("first chunk = %+v, %v; want one page and not done", p, err)
	}
	if n := r.countOf(`SELECT count(*) FROM job_cursors WHERE job = 'scoring.recompute_all'`); n != 1 {
		t.Fatalf("cursor rows = %d, want 1", n)
	}

	p, err = scoring.RecomputeChunk(t.Context(), db, now, time.Hour)
	if err != nil || !p.Done || p.Items != 1200 || p.Chunks != 2 {
		t.Fatalf("resumed chunk = %+v, %v; want done with 1200 items in 2 chunks", p, err)
	}
	if n := r.countOf(`SELECT count(*) FROM job_cursors`); n != 0 {
		t.Fatalf("cursor rows after completion = %d, want 0", n)
	}
}

func TestRecomputeAll_ChunksAndFinishes(t *testing.T) {
	r := newRig(t)
	r.seedCountedItems(2000)
	db, err := r.app.DB()
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()

	chunks := 0
	for {
		p, err := scoring.RecomputeChunk(t.Context(), db, now, time.Nanosecond)
		if err != nil {
			t.Fatal(err)
		}
		chunks++
		if p.Done {
			if p.Items != 2000 {
				t.Fatalf("items = %d, want 2000", p.Items)
			}
			break
		}
		if chunks > 10 {
			t.Fatal("pass did not finish")
		}
	}
	if chunks < 4 {
		t.Fatalf("chunks = %d, want at least 4 pages", chunks)
	}
}
