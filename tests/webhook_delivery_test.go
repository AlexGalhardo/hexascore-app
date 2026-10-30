//go:build integration

package tests

import (
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/queue"
	"github.com/AlexGalhardo/vukem/uuid"
	vwebhook "github.com/AlexGalhardo/vukem/webhook"

	"github.com/AlexGalhardo/hexascore-app/app/jobs"
	"github.com/AlexGalhardo/hexascore-app/app/webhooks"
)

func TestWebhookDelivery(t *testing.T) {
	t.Run("signed batch", func(t *testing.T) {
		r := newKeyRig(t)
		const email = "hooks-batch@example.com"
		r.signIn(email)
		r.app.Mails().Reset()

		var hits atomic.Int32
		var body batchPayload
		id, secret := insertHTTPEndpointAt(t, r, email, "http://127.0.0.1/placeholder")
		mw, err := vwebhook.Standard(secret, vwebhook.StandardOptions{})
		if err != nil {
			t.Fatal(err)
		}

		srv := httptest.NewServer(mw(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			hits.Add(1)
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				t.Errorf("decode batch: %v", err)
				http.Error(w, "bad", http.StatusBadRequest)

				return
			}

			w.WriteHeader(http.StatusOK)
		})))
		t.Cleanup(srv.Close)
		pointEndpoint(t, r, id, srv.URL+"/hook")

		svc, q := deliveryQueue(t, r)
		item := uuid.NewV7().String()
		now := time.Now().UTC().Truncate(time.Second)
		insertChange(t, r, item, "game", "updated", "{score}", "", now)
		latest := insertChange(t, r, item, "game", "updated", "{ratingCount}", "", now.Add(time.Second))
		if err := svc.Fanout(t.Context(), q); err != nil {
			t.Fatal(err)
		}

		if hits.Load() != 1 {
			t.Fatalf("deliveries = %d, want 1", hits.Load())
		}

		if body.Type != "item.batch" || len(body.Events) != 1 {
			t.Fatalf("body = %+v", body)
		}

		ev := body.Events[0]
		if ev.ID != "chg_"+strconv.FormatInt(latest, 10) || ev.Type != "item.updated" || ev.ItemType != "games" || ev.ItemID != item {
			t.Fatalf("event = %+v, latest %d", ev, latest)
		}

		if len(ev.Fields) != 2 || ev.Fields[0] != "ratingCount" || ev.Fields[1] != "score" {
			t.Fatalf("fields = %v", ev.Fields)
		}

		if cursor(t, r) != latest {
			t.Fatalf("cursor = %d, want %d", cursor(t, r), latest)
		}

		if err := svc.Fanout(t.Context(), q); err != nil {
			t.Fatal(err)
		}

		if hits.Load() != 1 {
			t.Fatalf("second fan-out delivered again: hits = %d", hits.Load())
		}

		assertDeliveryIndex(t, r)
		assertFanoutSchedule(t, r)
	})

	t.Run("gone", func(t *testing.T) {
		r := newKeyRig(t)
		const email = "hooks-gone@example.com"
		r.signIn(email)
		r.app.Mails().Reset()

		var hits atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			hits.Add(1)
			w.WriteHeader(http.StatusGone)
		}))
		t.Cleanup(srv.Close)

		id, _ := insertHTTPEndpointAt(t, r, email, srv.URL+"/gone")
		svc, q := deliveryQueue(t, r)
		insertChange(t, r, uuid.NewV7().String(), "movie", "updated", "{title}", "", time.Now().UTC())
		if err := svc.Fanout(t.Context(), q); err != nil {
			t.Fatal(err)
		}

		if hits.Load() != 1 {
			t.Fatalf("hits = %d, want 1", hits.Load())
		}

		if !endpointDisabled(t, r, id) {
			t.Fatal("410 left the endpoint enabled")
		}

		if n := len(r.app.Mails().Sent()); n != 0 {
			t.Fatalf("410 sent %d mails", n)
		}
	})

	t.Run("three exhausted", func(t *testing.T) {
		r := newKeyRig(t)
		const email = "hooks-fail@example.com"
		r.signIn(email)
		r.app.Mails().Reset()

		var hits atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			hits.Add(1)
			w.WriteHeader(http.StatusInternalServerError)
		}))
		t.Cleanup(srv.Close)

		id, _ := insertHTTPEndpointAt(t, r, email, srv.URL+"/fail")
		svc, q := deliveryQueue(t, r)
		for n := 1; n <= 4; n++ {
			insertChange(t, r, uuid.NewV7().String(), "book", "updated", "{title}", "", time.Now().UTC())
			if err := svc.Fanout(t.Context(), q); err != nil {
				t.Fatal(err)
			}

			disabled := endpointDisabled(t, r, id)
			mails := r.app.Mails().SentTo(email)
			switch n {
			case 1, 2:
				if disabled || len(mails) != 0 || int(hits.Load()) != n {
					t.Fatalf("after %d: disabled=%v mails=%d hits=%d", n, disabled, len(mails), hits.Load())
				}
			case 3:
				if !disabled || len(mails) != 1 || hits.Load() != 3 {
					t.Fatalf("after 3: disabled=%v mails=%d hits=%d", disabled, len(mails), hits.Load())
				}

				if mails[0].Subject != "Webhook endpoint disabled" {
					t.Fatalf("subject = %q", mails[0].Subject)
				}
			case 4:
				if hits.Load() != 3 || len(r.app.Mails().Sent()) != 1 {
					t.Fatalf("after disable: hits=%d mails=%d", hits.Load(), len(r.app.Mails().Sent()))
				}
			}
		}
	})
}

type batchPayload struct {
	Type   string `json:"type"`
	Events []struct {
		ID       string   `json:"id"`
		Type     string   `json:"type"`
		ItemType string   `json:"itemType"`
		ItemID   string   `json:"itemId"`
		Fields   []string `json:"fields"`
	} `json:"events"`
}

// insertHTTPEndpointAt seals a secret and stores an http endpoint. The https
// check is dropped on this throwaway database so the test can reach httptest
// without a custom TLS client. Production keeps the check.
func insertHTTPEndpointAt(t *testing.T, r *keyRig, email, rawURL string) (string, string) {
	t.Helper()

	dropHTTPSCheck(t, r)

	var userID int64
	if err := r.acc.DB.QueryRowContext(t.Context(), `SELECT id FROM users WHERE email = $1`, email).Scan(&userID); err != nil {
		t.Fatal(err)
	}

	secret, err := vwebhook.NewSecret()
	if err != nil {
		t.Fatal(err)
	}

	id := uuid.NewV7().String()
	_, err = r.acc.DB.ExecContext(t.Context(), `
INSERT INTO webhook_endpoints (id, user_id, url, events, item_types, secret_enc)
VALUES ($1, $2, $3, $4::text[], $5::text[], $6)`,
		id, userID, rawURL,
		"{item.created,item.updated,item.deleted}",
		"{games,movies,series,books}",
		r.app.Crypt.Seal([]byte(secret), webhooks.SecretAD),
	)
	if err != nil {
		t.Fatal(err)
	}

	return id, secret
}

func pointEndpoint(t *testing.T, r *keyRig, id, rawURL string) {
	t.Helper()

	res, err := r.acc.DB.ExecContext(t.Context(), `UPDATE webhook_endpoints SET url = $1 WHERE id = $2`, rawURL, id)
	if err != nil {
		t.Fatal(err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		t.Fatal(err)
	}

	if n != 1 {
		t.Fatalf("pointed %d endpoints, want 1", n)
	}
}

func dropHTTPSCheck(t *testing.T, r *keyRig) {
	t.Helper()

	var name string
	err := r.acc.DB.QueryRowContext(t.Context(), `
SELECT con.conname
FROM pg_constraint con
JOIN pg_class rel ON rel.oid = con.conrelid
WHERE rel.relname = 'webhook_endpoints' AND con.contype = 'c'
  AND pg_get_constraintdef(con.oid) ILIKE '%https://%'`).Scan(&name)
	if err != nil {
		t.Fatal(err)
	}

	quoted := `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
	if _, err := r.acc.DB.ExecContext(t.Context(), `ALTER TABLE webhook_endpoints DROP CONSTRAINT `+quoted); err != nil {
		t.Fatal(err)
	}
}

func deliveryQueue(t *testing.T, r *keyRig) (*webhooks.Service, *queue.Queue) {
	t.Helper()

	mailer, err := r.app.Mailer()
	if err != nil {
		t.Fatal(err)
	}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc, err := webhooks.Open(r.acc.DB, r.app.Crypt, mailer, time.Now, log, webhooks.Options{
		AllowHTTP: true,
		Allow:     []netip.Prefix{netip.MustParsePrefix("127.0.0.0/8")},
		Retries:   -1,
	})
	if err != nil {
		t.Fatal(err)
	}

	q := queue.New(queue.Sync{}, log)
	if err := q.Register(svc.Registration()); err != nil {
		t.Fatal(err)
	}

	return svc, q
}

func cursor(t *testing.T, r *keyRig) int64 {
	t.Helper()

	var id int64
	if err := r.acc.DB.QueryRowContext(t.Context(), `SELECT last_change_id FROM webhook_cursor WHERE id = 1`).Scan(&id); err != nil {
		t.Fatal(err)
	}

	return id
}

func endpointDisabled(t *testing.T, r *keyRig, id string) bool {
	t.Helper()

	var disabled sql.NullTime
	if err := r.acc.DB.QueryRowContext(t.Context(), `SELECT disabled_at FROM webhook_endpoints WHERE id = $1`, id).Scan(&disabled); err != nil {
		t.Fatal(err)
	}

	return disabled.Valid
}

func assertDeliveryIndex(t *testing.T, r *keyRig) {
	t.Helper()

	var def string
	err := r.acc.DB.QueryRowContext(t.Context(), `
SELECT indexdef FROM pg_indexes WHERE indexname = 'webhook_deliveries_endpoint_at'`).Scan(&def)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(def, "endpoint") || !strings.Contains(def, "attempted_at") {
		t.Fatalf("indexdef = %s", def)
	}
}

func assertFanoutSchedule(t *testing.T, r *keyRig) {
	t.Helper()

	sched, err := r.app.Schedule()
	if err != nil {
		t.Fatal(err)
	}

	want := "job:" + jobs.FanoutWebhooks{}.Kind()
	for _, task := range sched.Tasks() {
		if task.Name() == want && task.Expression() == "* * * * *" {
			return
		}
	}

	t.Fatalf("schedule missing %s every minute", want)
}
