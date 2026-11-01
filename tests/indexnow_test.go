//go:build integration

package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/config"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/indexnow"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/jobs"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// fakeIndexNow is an https IndexNow endpoint that records every urlList it receives.
func fakeIndexNow(t *testing.T) (*httptest.Server, func() [][]string) {
	t.Helper()

	var (
		mu  sync.Mutex
		got [][]string
	)

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Key     string   `json:"key"`
			URLList []string `json:"urlList"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Key != "indexnow-test-key" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		mu.Lock()
		got = append(got, body.URLList)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	return srv, func() [][]string {
		mu.Lock()
		defer mu.Unlock()

		return slices.Clone(got)
	}
}

// Catalogue writes reach IndexNow from the api_changes feed, batched off the request path; score-only updates do
// not, and a second run sends nothing new.
func TestIndexNowChanges(t *testing.T) {
	now := time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{
		Env:   map[string]string{"DATABASE_URL": pgtest.Database(t, "")},
		Clock: vukemtest.NewFakeClock(now).Now,
	})

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()
	zelda, mario := uuid.NewV7().String(), uuid.NewV7().String()

	for _, q := range []string{
		`INSERT INTO items (id, type, slug, wikidata_qid, popularity_prior, visibility) VALUES
  ($1, 'game', 'zelda', 'Q1', 0.5, 'public'), ($2, 'game', 'mario', 'Q2', 0.5, 'public')`,
		`INSERT INTO api_changes (item_id, item_type, action, fields) VALUES
  ($1, 'game', 'created', '{}'), ($2, 'game', 'updated', '{score,ratingCount}')`,
		// The visibility trigger records the hide as a "deleted" change.
		`UPDATE items SET visibility = 'takedown' WHERE id = $1 AND $2::text <> ''`,
	} {
		if _, err := db.SQL().ExecContext(ctx, q, zelda, mario); err != nil {
			t.Fatal(err)
		}
	}

	srv, got := fakeIndexNow(t)
	n, err := indexnow.New(indexnow.Config{Key: config.NewSecret("indexnow-test-key"), Endpoint: srv.URL},
		"https://hexascore.app", srv.Client())
	if err != nil {
		t.Fatal(err)
	}

	notify := func(ctx context.Context, urls []string) error { return n.Notify(ctx, urls...) }

	if err := jobs.RunIndexNowChanges(ctx, db, "https://hexascore.app/", now, notify); err != nil {
		t.Fatal(err)
	}

	calls := got()
	if len(calls) != 1 {
		t.Fatalf("IndexNow calls = %d, want 1", len(calls))
	}

	for _, want := range []string{
		"https://hexascore.app/games/zelda", "https://hexascore.app/pt-br/games/zelda", "https://hexascore.app/es/games/zelda",
	} {
		if !slices.Contains(calls[0], want) {
			t.Errorf("urlList %v lacks %s", calls[0], want)
		}
	}

	if slices.Contains(calls[0], "https://hexascore.app/games/mario") {
		t.Errorf("score-only update notified: %v", calls[0])
	}

	if err := jobs.RunIndexNowChanges(ctx, db, "https://hexascore.app", now.Add(time.Minute), notify); err != nil {
		t.Fatal(err)
	}

	if len(got()) != 1 {
		t.Fatalf("second run resent the batch: %v", got())
	}
}

// A failed notification keeps the cursor, so the next run sends the same URLs.
func TestIndexNowChanges_RetryKeepsCursor(t *testing.T) {
	now := time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{
		Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")},
	})

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()
	id := uuid.NewV7().String()

	for _, q := range []string{
		`INSERT INTO items (id, type, slug, wikidata_qid, popularity_prior, visibility) VALUES ($1, 'book', 'dune', 'Q3', 0.5, 'public')`,
		`INSERT INTO api_changes (item_id, item_type, action, fields) VALUES ($1, 'book', 'updated', '{title}')`,
	} {
		if _, err := db.SQL().ExecContext(ctx, q, id); err != nil {
			t.Fatal(err)
		}
	}

	fail := func(context.Context, []string) error { return context.DeadlineExceeded }
	if err := jobs.RunIndexNowChanges(ctx, db, "https://hexascore.app", now, fail); err == nil {
		t.Fatal("a failed notify reported success")
	}

	var sent []string
	keep := func(_ context.Context, urls []string) error { sent = urls; return nil }

	if err := jobs.RunIndexNowChanges(ctx, db, "https://hexascore.app", now, keep); err != nil {
		t.Fatal(err)
	}

	if !slices.Contains(sent, "https://hexascore.app/books/dune") {
		t.Fatalf("retry sent %v", sent)
	}
}

// INDEXNOW_KEY publishes the key file at /{key}.txt (served by vukem).
func TestIndexNowKeyFile(t *testing.T) {
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{
		"DATABASE_URL": pgtest.Database(t, ""),
		"APP_URL":      "https://example.test",
		"INDEXNOW_KEY": "indexnow-test-key",
	}})
	app.Client(t).Get("/indexnow-test-key.txt").AssertOK().AssertSee("indexnow-test-key")
}
