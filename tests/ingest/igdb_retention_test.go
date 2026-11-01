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

var igdbHits atomic.Int32

func (f *fakeSources) twitchToken(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte(`{"access_token": "abc", "expires_in": 5000, "token_type": "bearer"}`))
}

func (f *fakeSources) igdb(w http.ResponseWriter, _ *http.Request) {
	igdbHits.Add(1)

	raw, err := os.ReadFile("../../internal/ingest/testdata/igdb_games.json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	_, _ = w.Write(raw)
}

func TestIgdbRetention(t *testing.T) {
	f := loadFake(t)
	app := newIngestApp(t, f, "ENABLED", "WIKIDATA", "IGDB")
	ctx := context.Background()

	runCommand(t, app, "ingest:bootstrap", "--limit", "20")
	runCommand(t, app, "queue:work", "-queue", "ingest,default", "-stop-when-empty")

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	seed := func(qid, slug string, age time.Duration) {
		_, err := db.ExecContext(ctx, `INSERT INTO source_records (source, external_id, payload, hash, checked_at)
VALUES ('igdb', $1, jsonb_build_object('qid', $1::text, 'slug', $2::text), encode(sha256(convert_to($1, 'UTF8')), 'hex'),
now() - $3 * interval '1 second')`, qid, slug, int64(age.Seconds()))
		if err != nil {
			t.Fatal(err)
		}
	}

	// 2 days: past the 24 h refresh. 8 days: past the 7 day limit, deleted with its IGDB cover.
	seed("Q4267401", "the-witcher-3-wild-hunt", 48*time.Hour)
	seed("Q83495", "some-gone-game", 8*24*time.Hour)

	if _, err := db.ExecContext(ctx, `UPDATE items SET cover_key = 'covers/x.webp', cover_thumb_key = 'covers/x-t.webp',
cover_source = 'igdb' WHERE wikidata_qid = 'Q83495'`); err != nil {
		t.Fatal(err)
	}

	igdbHits.Store(0)
	startChain(t, app, ingest.IGDBRefresh)
	runCommand(t, app, "queue:work", "-queue", "ingest,default", "-stop-when-empty")

	if got := jobStatus(t, app)["igdb.refresh"]; got != "succeeded" {
		t.Fatalf("igdb.refresh = %q (all: %v)", got, jobStatus(t, app))
	}

	if igdbHits.Load() == 0 {
		t.Error("the 2 day record was not refreshed through IGDB")
	}

	if n := appCount(t, app, `SELECT count(*) FROM source_records WHERE source = 'igdb' AND external_id = 'Q4267401'
AND checked_at > now() - interval '1 hour'`); n != 1 {
		t.Errorf("2 day record not refreshed: %d", n)
	}

	if n := appCount(t, app, `SELECT count(*) FROM source_records WHERE source = 'igdb'
AND checked_at < now() - interval '7 days'`); n != 0 {
		t.Errorf("IGDB records older than 7 days: %d", n)
	}

	if n := appCount(
		t,
		app,
		`SELECT count(*) FROM items WHERE cover_source = 'igdb' AND wikidata_qid = 'Q83495'`,
	); n != 0 {
		t.Errorf("IGDB cover of the expired record kept: %d", n)
	}
}
