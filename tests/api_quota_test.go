//go:build integration

package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/api"
	"github.com/AlexGalhardo/hexascore-app/app/models"
)

type quotaRig struct {
	*keyRig
	store api.PGQuotaStore
	user  models.User
}

func newQuotaRig(t *testing.T, email string) *quotaRig {
	t.Helper()

	r := newKeyRig(t)

	return &quotaRig{keyRig: r, store: api.PGQuotaStore{DB: r.acc.DB}, user: makeUser(t, r.app, email, "user", true)}
}

func (r *quotaRig) meter() *api.Meter { return &api.Meter{Store: r.store, Now: r.app.Now} }

// serve runs one request of r.user through the quota middleware in front of a handler answering status.
func (r *quotaRig) serve(m *api.Meter, status int) *httptest.ResponseRecorder {
	r.t.Helper()

	q := api.Quota(m, api.FreeForAll, "https://hexascore.app/problems/")(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(status) }))
	rec := httptest.NewRecorder()
	c := web.NewCtx(rec, httptest.NewRequestWithContext(r.t.Context(), http.MethodGet, "/api/v1/games", nil), nil)
	auth.SetUser(c, api.GuardName, r.user)
	q.ServeHTTP(rec, c.Request())

	return rec
}

func (r *quotaRig) flush(m *api.Meter) {
	r.t.Helper()

	if err := m.Flush(r.t.Context()); err != nil {
		r.t.Fatal(err)
	}
}

func (r *quotaRig) used() int {
	return r.count(`SELECT COALESCE(sum(used), 0) FROM api_quota_counters WHERE user_id = $1`, r.user.ID)
}

func TestCreditBalanceConsumedFirst(t *testing.T) {
	r := newQuotaRig(t, "credits@example.com")
	if _, err := r.acc.DB.ExecContext(t.Context(),
		`INSERT INTO api_credit_balances (user_id, remaining) VALUES ($1, 2)`, r.user.ID); err != nil {
		t.Fatal(err)
	}

	m := r.meter()

	if got := r.serve(m, http.StatusOK).Header().Get("X-Quota-Remaining"); got != "101" {
		t.Fatalf("first X-Quota-Remaining = %s, want 101 (100 plan + 1 credit left)", got)
	}

	r.serve(m, http.StatusOK)
	r.serve(m, http.StatusOK)
	r.flush(m)

	if n := r.count(`SELECT remaining FROM api_credit_balances WHERE user_id = $1`, r.user.ID); n != 0 {
		t.Fatalf("credits left %d, want 0", n)
	}

	if n := r.used(); n != 1 {
		t.Fatalf("plan used %d, want 1", n)
	}
}

func TestRefundOn5xx(t *testing.T) {
	r := newQuotaRig(t, "refund@example.com")
	m := r.meter()

	for _, s := range []int{
		http.StatusOK, http.StatusNotFound, http.StatusBadRequest, http.StatusInternalServerError,
		http.StatusServiceUnavailable,
	} {
		r.serve(m, s)
	}

	r.flush(m)

	if n := r.used(); n != 2 {
		t.Fatalf("used %d, want 2 (200 and 404)", n)
	}

	day := r.app.Now().UTC().Format(time.DateOnly)
	if n := r.count(`SELECT count FROM api_usage_daily WHERE user_id = $1 AND day = $2`, r.user.ID, day); n != 2 {
		t.Fatalf("daily count %d, want 2", n)
	}

	if n := r.count(`SELECT refunded FROM api_usage_daily WHERE user_id = $1 AND day = $2`, r.user.ID, day); n != 2 {
		t.Fatalf("daily refunded %d, want 2", n)
	}
}

func TestQuotaFlushCrash(t *testing.T) {
	r := newQuotaRig(t, "crash@example.com")
	m := r.meter()

	for range 5 {
		r.serve(m, http.StatusOK)
	}

	// A flush that cannot commit writes nothing and keeps its deltas.
	cancelled, cancel := context.WithCancel(t.Context())
	cancel()

	if err := m.Flush(cancelled); err == nil {
		t.Fatal("flush on a cancelled context succeeded")
	}

	if n := r.used(); n != 0 {
		t.Fatalf("failed flush wrote %d", n)
	}

	r.flush(m)

	if n := r.used(); n != 5 {
		t.Fatalf("used %d after retry, want 5 (written once)", n)
	}

	// A crash drops the unflushed interval: the restarted meter undercounts it and never overcharges.
	for range 3 {
		r.serve(m, http.StatusOK)
	}

	restarted := r.meter()
	if got := r.serve(restarted, http.StatusOK).Header().Get("X-Quota-Remaining"); got != "94" {
		t.Fatalf("after restart X-Quota-Remaining = %s, want 94", got)
	}

	r.flush(restarted)

	if n := r.used(); n != 6 {
		t.Fatalf("used %d, want 6", n)
	}
}
