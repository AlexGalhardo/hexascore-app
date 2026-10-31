//go:build integration

package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/router"
	vwebhook "github.com/AlexGalhardo/vukem/webhook"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/AlexGalhardo/hexascore-app/app/jobs"
	"github.com/AlexGalhardo/hexascore-app/app/metrics"
	"github.com/AlexGalhardo/hexascore-app/app/webhooks"
)

func TestAPIWebhookTest(t *testing.T) {
	r, client := webhookClient(t, "hooks-ping@example.com", "pro")
	created := client.PostJSON("/api/v1/webhooks", map[string]any{
		"url": "https://1.1.1.1/hook", "events": []string{"item.updated"}, "itemTypes": []string{"games"},
	}).AssertCreated()
	var one struct {
		ID     string `json:"id"`
		Secret string `json:"secret"`
	}
	created.JSON(&one)

	dropHTTPSCheck(t, r)
	var hits atomic.Int32
	var kind string
	mw, err := vwebhook.Standard(one.Secret, vwebhook.StandardOptions{})
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(mw(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		hits.Add(1)
		var body struct {
			Type string `json:"type"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Errorf("decode: %v", err)
			http.Error(w, "bad", http.StatusBadRequest)

			return
		}

		kind = body.Type
		w.WriteHeader(http.StatusOK)
	})))
	t.Cleanup(srv.Close)
	pointEndpoint(t, r, one.ID, srv.URL+"/hook")

	before := quotaUsed(t, client)
	res := client.PostJSON("/api/v1/webhooks/"+one.ID+"/test", map[string]any{}).AssertOK().AssertMatchesOpenAPI()
	var got struct {
		Delivered bool `json:"delivered"`
		Status    int  `json:"status"`
	}
	res.JSON(&got)
	if !got.Delivered || got.Status != http.StatusOK || hits.Load() != 1 || kind != "item.batch" {
		t.Fatalf("ping = %+v hits=%d type=%s", got, hits.Load(), kind)
	}

	if after := quotaUsed(t, client); after != before+1 {
		t.Fatalf("quota %d → %d, want +1 from the usage read only", before, after)
	}

	_, other := webhookClient(t, "hooks-ping-other@example.com", "pro")
	other.PostJSON("/api/v1/webhooks/"+one.ID+"/test", map[string]any{}).AssertNotFound()

	if _, err := r.acc.DB.ExecContext(
		t.Context(),
		`UPDATE webhook_endpoints SET disabled_at = now() WHERE id = $1`,
		one.ID,
	); err != nil {
		t.Fatal(err)
	}

	client.PostJSON("/api/v1/webhooks/"+one.ID+"/test", map[string]any{}).AssertNotFound()
	if hits.Load() != 1 {
		t.Fatalf("disabled endpoint was called: hits=%d", hits.Load())
	}

	_, free := webhookClient(t, "hooks-ping-free@example.com", "")
	free.PostJSON("/api/v1/webhooks/"+one.ID+"/test", map[string]any{}).
		AssertForbidden().
		AssertJSONPath("code", "plan-required")
}

func TestPruneDeliveries(t *testing.T) {
	r := newKeyRig(t)
	now := time.Now().UTC()
	cutoff := now.Add(-webhooks.DeliveryRetain).UnixMilli()
	old := insertDelivery(t, r, cutoff-1)
	kept := insertDelivery(t, r, cutoff)
	fresh := insertDelivery(t, r, now.UnixMilli())

	n, err := webhooks.PruneDeliveries(t.Context(), r.acc.DB, now)
	if err != nil {
		t.Fatal(err)
	}

	if n != 1 || deliveryCount(t, r, old) != 0 || deliveryCount(t, r, kept) != 1 || deliveryCount(t, r, fresh) != 1 {
		t.Fatalf(
			"deleted=%d old=%d kept=%d fresh=%d",
			n,
			deliveryCount(t, r, old),
			deliveryCount(t, r, kept),
			deliveryCount(t, r, fresh),
		)
	}

	sched, err := r.app.Schedule()
	if err != nil {
		t.Fatal(err)
	}

	want := "job:" + jobs.PruneWebhookDeliveries{}.Kind()
	for _, task := range sched.Tasks() {
		if task.Name() == want && task.Expression() == "40 4 * * *" {
			return
		}
	}

	t.Fatalf("schedule missing %s at 04:40", want)
}

func TestMetrics_WebhookOutcome(t *testing.T) {
	reader := metric.NewManualReader()
	mp := metric.NewMeterProvider(metric.WithReader(reader))
	r := newKeyRigWith(t,
		vukem.WithRateLimiters(func(_ string, _ int, window time.Duration) ratelimit.Limiter {
			return ratelimit.NewMemory(1000, window)
		}),
		vukem.WithRoutes(func(_ *router.Router, built *vukem.App) error {
			metrics.Bind(built, mp)

			return nil
		}),
	)
	token := r.create(r.signIn("hooks-metric@example.com"), "ci")
	setPlan(t, r, "hooks-metric@example.com", "pro")
	client := r.app.Client(t).WithToken(token)
	created := client.PostJSON("/api/v1/webhooks", map[string]any{
		"url": "https://1.1.1.1/hook", "events": []string{"item.updated"}, "itemTypes": []string{"games"},
	}).AssertCreated()
	var one struct {
		ID     string `json:"id"`
		Secret string `json:"secret"`
	}
	created.JSON(&one)
	dropHTTPSCheck(t, r)

	okSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(okSrv.Close)
	failSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(failSrv.Close)

	pointEndpoint(t, r, one.ID, okSrv.URL+"/ok")
	client.PostJSON("/api/v1/webhooks/"+one.ID+"/test", map[string]any{}).AssertOK()
	pointEndpoint(t, r, one.ID, failSrv.URL+"/fail")
	client.PostJSON("/api/v1/webhooks/"+one.ID+"/test", map[string]any{}).AssertOK()

	got := webhookOutcomes(t, reader)
	if got["success"] != 1 || got["failure"] != 1 {
		t.Fatalf("outcomes = %v", got)
	}
}

func insertDelivery(t *testing.T, r *keyRig, at int64) int64 {
	t.Helper()

	var id int64
	err := r.acc.DB.QueryRowContext(t.Context(), `
INSERT INTO webhook_deliveries (endpoint, event_id, status, duration_ms, error, attempted_at)
VALUES ('ep', 'evt', 200, 1, '', $1)
RETURNING id`, at).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}

	return id
}

func deliveryCount(t *testing.T, r *keyRig, id int64) int {
	t.Helper()

	var n int
	if err := r.acc.DB.QueryRowContext(t.Context(), `SELECT count(*) FROM webhook_deliveries WHERE id = $1`, id).
		Scan(&n); err != nil {
		t.Fatal(err)
	}

	return n
}

func webhookOutcomes(t *testing.T, reader *metric.ManualReader) map[string]int {
	t.Helper()

	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatal(err)
	}

	got := map[string]int{}

	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != "api.webhook.deliveries" {
				continue
			}

			sum, ok := m.Data.(metricdata.Sum[int64])
			if !ok {
				t.Fatalf("deliveries data %T", m.Data)
			}

			for _, dp := range sum.DataPoints {
				for _, attr := range dp.Attributes.ToSlice() {
					if string(attr.Key) == "outcome" {
						got[attr.Value.AsString()] += int(dp.Value)
					}
				}
			}
		}
	}

	return got
}
