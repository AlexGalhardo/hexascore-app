//go:build integration

package tests

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	vwebhook "github.com/AlexGalhardo/vukem/webhook"
)

// TestFanoutWebhooks_RetrySendsOnce redelivers a job whose first run already got a 2xx (the crash between the
// response and the queue ack) and expects a single request to the endpoint.
func TestFanoutWebhooks_RetrySendsOnce(t *testing.T) {
	r := newKeyRig(t)
	const email = "hooks-retry@example.com"
	r.signIn(email)

	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	id, _ := insertHTTPEndpointAt(t, r, email, srv.URL+"/hook")
	_, q := deliveryQueue(t, r)
	d := vwebhook.Delivery{ID: "msg_retry_once", Endpoint: id, Payload: []byte(`{"type":"item.batch","events":[]}`)}
	for range 2 {
		if err := q.DispatchKind(t.Context(), vwebhook.DeliverKind, d); err != nil {
			t.Fatal(err)
		}
	}

	if hits.Load() != 1 {
		t.Fatalf("requests = %d, want 1", hits.Load())
	}
}
