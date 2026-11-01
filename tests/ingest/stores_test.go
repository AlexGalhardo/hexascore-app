//go:build integration

package ingest_test

import (
	"net/http"
	"os"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
)

func serveFixture(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		raw, err := os.ReadFile("../../internal/ingest/testdata/" + name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		_, _ = w.Write(raw)
	}
}

func TestSteamAndGogChainsFetchGames(t *testing.T) {
	f := loadFake(t)
	app := newIngestApp(t, f, "ENABLED", "WIKIDATA", "STEAM", "GOG")

	runCommand(t, app, "ingest:bootstrap", "--limit", "20")
	runCommand(t, app, "queue:work", "-queue", "ingest,default", "-stop-when-empty")

	startChain(t, app, ingest.SteamWeekly)
	startChain(t, app, ingest.GOGWeekly)
	runCommand(t, app, "queue:work", "-queue", "ingest,default", "-stop-when-empty")

	got := jobStatus(t, app)
	for _, job := range []string{"fetch.steam", "fetch.gog"} {
		if got[job] != "succeeded" {
			t.Errorf("%s = %q (all: %v)", job, got[job], got)
		}
	}

	for _, source := range []string{"steam", "gog"} {
		if n := appCount(t, app, `SELECT count(*) FROM source_records WHERE source = $1 AND external_id = 'Q4267401'`,
			source); n != 1 {
			t.Errorf("%s records for the game = %d, want 1", source, n)
		}
	}
}
