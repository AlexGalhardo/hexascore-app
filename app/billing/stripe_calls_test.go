package billing

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"sync"
	"testing"

	"github.com/AlexGalhardo/vukem/contrib/stripe"
)

type stripeStub struct {
	mu      sync.Mutex
	queries []url.Values
}

func newStripeStub(t *testing.T, mux *http.ServeMux) (*stripe.Client, *stripeStub) {
	t.Helper()

	s := &stripeStub{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		s.queries = append(s.queries, r.URL.Query())
		s.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		mux.ServeHTTP(w, r)
	}))
	t.Cleanup(srv.Close)

	c, err := stripe.New("sk_test_stub", stripe.WithBaseURL(srv.URL), stripe.WithRetries(0))
	if err != nil {
		t.Fatal(err)
	}

	return c, s
}

func reply(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()

	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Error(err)
	}
}

func TestCustomerSubscriptionsPages(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("starting_after") == "" {
			reply(
				t,
				w,
				map[string]any{
					"data":     []any{map[string]any{"id": "sub_1"}, map[string]any{"id": "sub_2"}},
					"has_more": true,
				},
			)

			return
		}

		reply(t, w, map[string]any{"data": []any{map[string]any{"id": "sub_3"}}, "has_more": false})
	})

	c, stub := newStripeStub(t, mux)

	subs, err := customerSubscriptions(t.Context(), c, "cus_1")
	if err != nil {
		t.Fatal(err)
	}

	var ids []string
	for _, s := range subs {
		ids = append(ids, s.ID)
	}

	if !slices.Equal(ids, []string{"sub_1", "sub_2", "sub_3"}) {
		t.Fatalf("subscriptions = %v", ids)
	}

	if len(stub.queries) != 2 {
		t.Fatalf("calls = %d, want 2", len(stub.queries))
	}

	first, second := stub.queries[0], stub.queries[1]
	if first.Get("customer") != "cus_1" || first.Get("status") != "all" || first.Get("limit") != "100" ||
		first.Get("starting_after") != "" {
		t.Errorf("first page query = %v", first)
	}

	if second.Get("starting_after") != "sub_2" || second.Get("customer") != "cus_1" {
		t.Errorf("second page query = %v", second)
	}
}

func TestExpireOpenCheckoutsOnlySubscriptionMode(t *testing.T) {
	var (
		mu      sync.Mutex
		expired []string
	)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/checkout/sessions", func(w http.ResponseWriter, _ *http.Request) {
		reply(t, w, map[string]any{"data": []any{
			map[string]any{"id": "cs_test_sub", "mode": "subscription"},
			map[string]any{"id": "cs_test_pay", "mode": "payment"},
			map[string]any{"id": "cs_test_gone", "mode": "subscription"},
		}})
	})
	mux.HandleFunc("POST /v1/checkout/sessions/{id}/expire", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		expired = append(expired, r.PathValue("id"))
		mu.Unlock()

		if r.PathValue("id") == "cs_test_gone" {
			w.WriteHeader(http.StatusBadRequest)
			reply(t, w, map[string]any{"error": map[string]any{"type": "invalid_request_error", "message": "not open"}})

			return
		}

		reply(t, w, map[string]any{"id": r.PathValue("id")})
	})

	c, stub := newStripeStub(t, mux)

	if err := expireOpenCheckouts(t.Context(), c, "cus_1"); err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(expired, []string{"cs_test_sub", "cs_test_gone"}) {
		t.Errorf("expired = %v", expired)
	}

	if q := stub.queries[0]; q.Get("customer") != "cus_1" || q.Get("status") != "open" || q.Get("limit") != "100" {
		t.Errorf("list query = %v", q)
	}
}
