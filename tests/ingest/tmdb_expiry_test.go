//go:build integration

package ingest_test

import (
	"context"
	"net/http"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
)

// tmdbHits counts the requests the fake TMDB answered.
var tmdbHits atomic.Int64

func (f *fakeSources) tmdb(w http.ResponseWriter, _ *http.Request) {
	tmdbHits.Add(1)

	raw, err := os.ReadFile("../../internal/ingest/testdata/tmdb_movie.json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	_, _ = w.Write(raw)
}

func TestTmdbExpiry(t *testing.T) {
	f := loadFake(t)
	app := newIngestApp(t, f, "ENABLED", "WIKIDATA", "TMDB")
	ctx := context.Background()

	runCommand(t, app, "ingest:bootstrap", "--limit", "20")
	runCommand(t, app, "queue:work", "-queue", "ingest,default", "-stop-when-empty")

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	// 151 days: refreshed. 181 days: deleted, its cover dropped, the item normalised again.
	seed := func(qid string, id int, age time.Duration) {
		_, err := db.ExecContext(ctx, `INSERT INTO source_records (source, external_id, payload, hash, checked_at)
VALUES ('tmdb', $1, jsonb_build_object('qid', $1::text, 'id', $2::int), encode(sha256(convert_to($1, 'UTF8')), 'hex'), now() - $3 * interval '1 second')`,
			qid, id, int64(age.Seconds()))
		if err != nil {
			t.Fatal(err)
		}
	}

	seed("Q83495", 603, 151*24*time.Hour)
	seed("Q1079", 1396, 181*24*time.Hour)

	if _, err := db.ExecContext(ctx, `UPDATE items SET cover_key = 'covers/x.webp', cover_thumb_key = 'covers/x-t.webp',
cover_source = 'tmdb' WHERE wikidata_qid = 'Q1079'`); err != nil {
		t.Fatal(err)
	}

	var before time.Time
	if err := db.QueryRowContext(ctx,
		`SELECT normalized_at FROM source_records WHERE source = 'wikidata' AND external_id = 'Q1079'`).
		Scan(&before); err != nil {
		t.Fatal(err)
	}

	hits := tmdbHits.Load()

	startChain(t, app, ingest.TMDBExpire)
	runCommand(t, app, "queue:work", "-queue", "ingest,default", "-stop-when-empty")

	if got := jobStatus(t, app)["tmdb.expire"]; got != "succeeded" {
		t.Fatalf("tmdb.expire = %q (all: %v)", got, jobStatus(t, app))
	}

	if tmdbHits.Load() == hits {
		t.Error("the 151 day record was not refreshed through TMDB")
	}

	if n := appCount(
		t,
		app,
		`SELECT count(*) FROM source_records WHERE source = 'tmdb' AND external_id = 'Q83495' AND checked_at > now() - interval '1 hour'`,
	); n != 1 {
		t.Errorf("151 day record not refreshed: %d", n)
	}

	if n := appCount(
		t,
		app,
		`SELECT count(*) FROM source_records WHERE source = 'tmdb' AND external_id = 'Q1079'`,
	); n != 0 {
		t.Errorf("181 day record still stored: %d", n)
	}

	if n := appCount(
		t,
		app,
		`SELECT count(*) FROM source_records WHERE source = 'tmdb' AND checked_at < now() - interval '180 days'`,
	); n != 0 {
		t.Errorf("records older than 180 days: %d", n)
	}

	if n := appCount(
		t,
		app,
		`SELECT count(*) FROM items WHERE wikidata_qid = 'Q1079' AND cover_source IS NULL AND cover_key IS NULL`,
	); n != 1 {
		t.Errorf("TMDB cover not dropped: %d", n)
	}

	if n := appCount(
		t,
		app,
		`SELECT count(*) FROM source_records WHERE source = 'wikidata' AND external_id = 'Q1079' AND normalized_at > $1`,
		before,
	); n != 1 {
		t.Errorf("item not normalised again: %d", n)
	}
}
