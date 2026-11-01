//go:build integration

package ingest_test

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/wikipedia"
)

func TestWikipediaTargetsComeFromWikidataSitelinks(t *testing.T) {
	db := newDB(t)
	ctx := t.Context()

	seed := func(id, links string, ageHours int) {
		_, err := db.ExecContext(ctx, `INSERT INTO source_records (source, external_id, payload, hash, fetched_at)
VALUES ('wikidata', $1, $2::jsonb, repeat('a', 64), now() - make_interval(hours => $3))`,
			id, fmt.Sprintf(`{"sitelinks":%s}`, links), ageHours)
		if err != nil {
			t.Fatal(err)
		}
	}

	seed("Q1", `{"enwiki":"One","ptwiki":"Um"}`, 1)
	seed("Q2", `{"enwiki":"Two"}`, 100)
	seed("Q3", `{"ptwiki":"Tres"}`, 1)
	seed("Q4", `{}`, 1)
	seed("Q5", `{"enwiki":"Five"}`, 1)

	if _, err := db.ExecContext(ctx, `INSERT INTO source_records (source, external_id, payload, hash)
VALUES ('wikipedia:en', 'Q5', '{}'::jsonb, repeat('b', 64))`); err != nil {
		t.Fatal(err)
	}

	w, _ := wikipedia.New("en")

	collect := func(p ingest.RunParams, after string) []string {
		var ids []string

		for tg, err := range w.Targets(ctx, db, p, after) {
			if err != nil {
				t.Fatal(err)
			}

			ids = append(ids, tg.ID+"="+tg.Type)
		}

		return ids
	}

	if got := collect(
		ingest.RunParams{Mode: run.ModeFull},
		"",
	); !slices.Equal(
		got,
		[]string{"Q1=One", "Q2=Two", "Q5=Five"},
	) {
		t.Errorf("full = %v", got)
	}

	if got := collect(ingest.RunParams{Mode: run.ModeFull}, "Q1"); !slices.Equal(got, []string{"Q2=Two", "Q5=Five"}) {
		t.Errorf("after Q1 = %v", got)
	}

	since := time.Now().Add(-24 * time.Hour)
	if got := collect(
		ingest.RunParams{Mode: run.ModeIncremental, Since: since},
		"",
	); !slices.Equal(
		got,
		[]string{"Q1=One", "Q5=Five"},
	) {
		t.Errorf("incremental = %v", got)
	}

	if got := collect(ingest.RunParams{Mode: run.ModeBootstrap}, ""); !slices.Equal(got, []string{"Q1=One", "Q2=Two"}) {
		t.Errorf("bootstrap must skip stored extracts, got %v", got)
	}

	if got := collect(ingest.RunParams{Mode: run.ModeFull, Limit: 1}, ""); len(got) != 1 {
		t.Errorf("limit = %v", got)
	}
}
