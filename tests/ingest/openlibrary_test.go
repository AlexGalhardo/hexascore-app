//go:build integration

package ingest_test

import (
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/normalize"
)

var openLibraryHits atomic.Int32

// openlibrary serves the edition fixture for any /isbn/ or /books/ path and the work fixture for /works/.
func (f *fakeSources) openlibrary(w http.ResponseWriter, r *http.Request) {
	openLibraryHits.Add(1)

	name := "openlibrary_isbn.json"
	if strings.Contains(r.URL.Path, "/works/") {
		name = "openlibrary_work.json"
	}

	raw, err := os.ReadFile("../../internal/ingest/testdata/" + name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	_, _ = w.Write(raw)
}

func TestOpenLibraryOffWithoutLegalFile(t *testing.T) {
	f := loadFake(t)
	app := newIngestApp(t, f, "ENABLED", "WIKIDATA")

	runCommand(t, app, "ingest:bootstrap", "--limit", "20")
	runCommand(t, app, "queue:work", "-queue", "ingest,default", "-stop-when-empty")

	openLibraryHits.Store(0)
	startChain(t, app, ingest.OpenLibraryWeekly)
	runCommand(t, app, "queue:work", "-queue", "ingest,default", "-stop-when-empty")

	got := jobStatus(t, app)
	if got["fetch.openlibrary"] != "skipped" {
		t.Errorf("fetch.openlibrary = %q, want skipped (ingest_openlibrary defaults off)", got["fetch.openlibrary"])
	}

	if n := openLibraryHits.Load(); n != 0 {
		t.Errorf("open library hit %d times with its kill switch off", n)
	}

	if n := appCount(t, app, `SELECT count(*) FROM source_records WHERE source = 'openlibrary'`); n != 0 {
		t.Errorf("openlibrary records = %d, want 0", n)
	}
}

func TestOpenLibraryOnFetchesBooks(t *testing.T) {
	f := loadFake(t)
	app := newIngestApp(t, f, "ENABLED", "WIKIDATA", "OPENLIBRARY")

	runCommand(t, app, "ingest:bootstrap", "--limit", "20")
	runCommand(t, app, "queue:work", "-queue", "ingest,default", "-stop-when-empty")

	startChain(t, app, ingest.OpenLibraryWeekly)
	runCommand(t, app, "queue:work", "-queue", "ingest,default", "-stop-when-empty")

	if got := jobStatus(t, app); got["fetch.openlibrary"] != "succeeded" {
		t.Fatalf("fetch.openlibrary = %q (all: %v)", got["fetch.openlibrary"], got)
	}

	if n := appCount(t, app, `SELECT count(*) FROM source_records WHERE source = 'openlibrary'
AND external_id = 'Q208460' AND payload ? 'overview'`); n != 1 {
		t.Errorf("openlibrary records for the book = %d, want 1", n)
	}
}

func TestSeedDataSourcesReplacesLicencePlaceholders(t *testing.T) {
	db := newDB(t)

	if _, err := db.ExecContext(t.Context(), `INSERT INTO data_sources (code, name, licence, url) VALUES
('openlibrary', 'Open Library', 'OPEN-LIBRARY-PENDING-A05', 'https://openlibrary.org'),
('steam', 'Steam', 'owner-reviewed', 'https://store.steampowered.com')`); err != nil {
		t.Fatal(err)
	}

	if err := normalize.SeedDataSources(t.Context(), db); err != nil {
		t.Fatal(err)
	}

	var ol, steam string
	if err := db.QueryRowContext(t.Context(), `SELECT
(SELECT licence FROM data_sources WHERE code = 'openlibrary'),
(SELECT licence FROM data_sources WHERE code = 'steam')`).Scan(&ol, &steam); err != nil {
		t.Fatal(err)
	}

	if ol != normalize.Licence(normalize.SourceOpenLibrary) || ol == "OPEN-LIBRARY-PENDING-A05" {
		t.Errorf("openlibrary licence = %q, the placeholder must be replaced", ol)
	}

	if steam != "owner-reviewed" {
		t.Errorf("steam licence = %q, a reviewed value must be kept", steam)
	}
}
