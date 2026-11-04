//go:build integration

package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// `alerts:test --digest` posts exactly one silent message to the ops chat of a fake Bot API (no live call), even
// though nothing was ingested.
func TestIngestDigest_AlertsTestCommandSendsOne(t *testing.T) {
	var (
		mu   sync.Mutex
		sent []map[string]any
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var m map[string]any
		_ = json.NewDecoder(r.Body).Decode(&m)
		mu.Lock()
		sent = append(sent, m)
		mu.Unlock()
		_, _ = w.Write([]byte(`{"ok":true,"result":{}}`))
	}))
	t.Cleanup(srv.Close)

	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{
		"DATABASE_URL":       pgtest.Database(t, ""),
		"TELEGRAM_BOT_TOKEN": "123:test-token", "TELEGRAM_CHAT_ID": "page-chat",
		"TELEGRAM_ERROR_CHAT_ID": "error-chat", "TELEGRAM_OPS_CHAT_ID": "ops-chat", "TELEGRAM_API_URL": srv.URL,
	}})

	for _, cmd := range app.Commands() {
		if cmd.Name() == "alerts:test" {
			if err := cmd.Run(t.Context(), app.App, []string{"--digest"}); err != nil {
				t.Fatal(err)
			}
		}
	}

	mu.Lock()
	defer mu.Unlock()

	if len(sent) != 1 {
		t.Fatalf("Bot API calls = %d, want 1: %v", len(sent), sent)
	}

	if sent[0]["chat_id"] != "ops-chat" || sent[0]["disable_notification"] != true ||
		!strings.Contains(sent[0]["text"].(string), "<b>Ingest</b>") {
		t.Errorf("message = %v", sent[0])
	}
}
