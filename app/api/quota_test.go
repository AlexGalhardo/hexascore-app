package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// memStore is a QuotaStore in memory; fail makes the next Save fail.
type memStore struct {
	mu      sync.Mutex
	used    map[string]int // "user/period"
	credits map[int64]int
	days    map[string][2]int // "user/day" -> count, refunded
	saves   int
	fail    bool
}

func newMemStore() *memStore {
	return &memStore{used: map[string]int{}, credits: map[int64]int{}, days: map[string][2]int{}}
}

func key(id int64, s string) string { return fmt.Sprintf("%d/%s", id, s) }

func (s *memStore) Load(_ context.Context, id int64, period string) (int, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.used[key(id, period)], s.credits[id], nil
}

func (s *memStore) Save(_ context.Context, batch []UserDelta) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.fail {
		s.fail = false
		return errors.New("disk on fire")
	}

	s.saves++

	for _, d := range batch {
		for p, n := range d.Used {
			s.used[key(d.UserID, p)] += n
		}

		s.credits[d.UserID] -= d.Credits

		for day, n := range d.Count {
			v := s.days[key(d.UserID, day)]
			v[0] += n
			s.days[key(d.UserID, day)] = v
		}

		for day, n := range d.Refunded {
			v := s.days[key(d.UserID, day)]
			v[1] += n
			s.days[key(d.UserID, day)] = v
		}
	}

	return nil
}

type clock struct{ t time.Time }

func (c *clock) now() time.Time { return c.t }

func free(t *testing.T) Plan {
	t.Helper()

	p, ok := PlanByID(PlanFree)
	if !ok {
		t.Fatal("no free plan")
	}

	return p
}

// metered serves h behind Quota for user 7, like Guard would have authenticated it.
func metered(m *Meter, plans PlanResolver, h http.Handler) http.Handler {
	q := Quota(m, plans, "https://hexascore.app/problems/")(h)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := web.NewCtx(w, r, nil)
		auth.SetUser(c, GuardName, models.User{ID: 7})
		q.ServeHTTP(w, c.Request())
	})
}

func get(h http.Handler) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/games", nil))

	return rec
}

func TestPlans(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		id                   string
		quota, pageSize, rps int
		lifetime             bool
	}{
		{PlanFree, 100, 10, 1, true},
		{PlanPro, 50_000, 100, 10, false},
		{PlanBusiness, 500_000, 100, 50, false},
	} {
		p, ok := PlanByID(tc.id)
		if !ok || p.Quota != tc.quota || p.MaxPageSize != tc.pageSize || p.RPS != tc.rps || p.Lifetime != tc.lifetime {
			t.Errorf("PlanByID(%q) = %+v, %v", tc.id, p, ok)
		}
	}

	if _, ok := PlanByID("enterprise"); ok {
		t.Error("unknown plan accepted")
	}

	pro, _ := PlanByID(PlanPro)
	now := time.Date(2026, 10, 31, 23, 59, 0, 0, time.UTC)

	if got := pro.Period(now); got != "202610" {
		t.Errorf("Period = %q", got)
	}

	if got := pro.ResetIn(now); got != 60 {
		t.Errorf("ResetIn = %d, want 60", got)
	}

	if got := free(t).Period(now); got != "life" || free(t).ResetIn(now) != 0 {
		t.Errorf("free period %q", got)
	}
}

func TestQuotaFreeKeyStopsAt100(t *testing.T) {
	t.Parallel()

	m := &Meter{Store: newMemStore(), Now: (&clock{time.Now()}).now}
	h := metered(m, FreeForAll, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 1; i <= 100; i++ {
		if rec := get(h); rec.Code != http.StatusOK {
			t.Fatalf("request %d = %d, want 200", i, rec.Code)
		}
	}

	rec := get(h)
	if rec.Code != http.StatusPaymentRequired {
		t.Fatalf("request 101 = %d, want 402", rec.Code)
	}

	if got := rec.Header().Get("X-Quota-Remaining"); got != "0" {
		t.Errorf("X-Quota-Remaining = %q", got)
	}

	if body := rec.Body.String(); !strings.Contains(body, `"code":"quota-exhausted"`) ||
		!strings.Contains(body, `"plan":"free"`) ||
		!strings.Contains(body, `"limit":100`) {
		t.Errorf("402 body = %s", body)
	}
}

func TestQuotaRefundsBadRequestsAndServerErrors(t *testing.T) {
	t.Parallel()

	st := newMemStore()
	clk := &clock{time.Date(2026, 10, 29, 12, 0, 0, 0, time.UTC)}
	m := &Meter{Store: st, Now: clk.now}

	status := http.StatusOK
	h := metered(
		m,
		FreeForAll,
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(status) }),
	)

	for _, s := range []int{
		http.StatusOK, http.StatusNotFound, http.StatusNotModified, http.StatusBadRequest,
		http.StatusInternalServerError, http.StatusBadGateway,
	} {
		status = s
		get(h)
	}

	if err := m.Flush(t.Context()); err != nil {
		t.Fatal(err)
	}

	// 200, 404 and 304 count; 400 and the two 5xx do not, and the 5xx are recorded as refunded.
	if got := st.used[key(7, "life")]; got != 3 {
		t.Errorf("used = %d, want 3", got)
	}

	if got := st.days[key(7, "2026-10-29")]; got != [2]int{3, 2} {
		t.Errorf("daily = %v, want [3 2]", got)
	}
}

func TestQuotaCreditsBeforePlan(t *testing.T) {
	t.Parallel()

	st := newMemStore()
	st.credits[7] = 2
	m := &Meter{Store: st, Now: (&clock{time.Now()}).now}

	u, err := m.Admit(t.Context(), 7, "", free(t))
	if err != nil || u.Remaining != 101 {
		t.Fatalf("first: %+v, %v (want remaining 100 + 1 credit)", u, err)
	}

	_, _ = m.Admit(t.Context(), 7, "", free(t))
	u, _ = m.Admit(t.Context(), 7, "", free(t))

	if u.Remaining != 99 {
		t.Fatalf("third: remaining %d, want 99", u.Remaining)
	}

	if err := m.Flush(t.Context()); err != nil {
		t.Fatal(err)
	}

	if st.credits[7] != 0 || st.used[key(7, "life")] != 1 {
		t.Fatalf("credits %d used %d, want 0 and 1", st.credits[7], st.used[key(7, "life")])
	}
}

func TestQuotaFlushFailureKeepsDeltas(t *testing.T) {
	t.Parallel()

	st := newMemStore()
	clk := &clock{time.Now()}
	m := &Meter{Store: st, Now: clk.now}

	for range 3 {
		_, _ = m.Admit(t.Context(), 7, "", free(t))
	}

	st.fail = true
	if err := m.Flush(t.Context()); err == nil {
		t.Fatal("flush succeeded")
	}

	_, _ = m.Admit(t.Context(), 7, "", free(t))

	if err := m.Flush(t.Context()); err != nil {
		t.Fatal(err)
	}

	if got := st.used[key(7, "life")]; got != 4 {
		t.Fatalf("used = %d, want 4 (written once)", got)
	}

	// Unflushed for longer than MaxUnflushed: fail closed.
	_, _ = m.Admit(t.Context(), 7, "", free(t))
	clk.t = clk.t.Add(MaxUnflushed + time.Second)

	if _, err := m.Admit(t.Context(), 7, "", free(t)); !errors.Is(err, ErrMeterStale) {
		t.Fatalf("stale meter: err = %v", err)
	}

	h := metered(m, FreeForAll, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	if rec := get(h); rec.Code != http.StatusServiceUnavailable || rec.Header().Get("Retry-After") != "30" {
		t.Fatalf(
			"stale meter answered %d (Retry-After %q), want 503 with Retry-After: 30",
			rec.Code,
			rec.Header().Get("Retry-After"),
		)
	}
}

func TestQuotaMonthRollover(t *testing.T) {
	t.Parallel()

	st := newMemStore()
	clk := &clock{time.Date(2026, 10, 31, 23, 59, 59, 0, time.UTC)}
	m := &Meter{Store: st, Now: clk.now}
	pro, _ := PlanByID(PlanPro)

	_, _ = m.Admit(t.Context(), 7, "", pro)
	clk.t = clk.t.Add(2 * time.Second)
	u, _ := m.Admit(t.Context(), 7, "", pro)

	if u.Remaining != pro.Quota-1 {
		t.Fatalf("November remaining %d, want %d", u.Remaining, pro.Quota-1)
	}

	if err := m.Flush(t.Context()); err != nil {
		t.Fatal(err)
	}

	if st.used[key(7, "202610")] != 1 || st.used[key(7, "202611")] != 1 {
		t.Fatalf("counters %v", st.used)
	}
}

// slowStore holds Save until release is closed.
type slowStore struct {
	*memStore
	started, release chan struct{}
}

func (s slowStore) Save(ctx context.Context, batch []UserDelta) error {
	close(s.started)
	<-s.release

	return s.memStore.Save(ctx, batch)
}

func TestQuotaNoOverAdmissionDuringFlush(t *testing.T) {
	t.Parallel()

	st := slowStore{memStore: newMemStore(), started: make(chan struct{}), release: make(chan struct{})}
	m := &Meter{Store: st, Now: (&clock{time.Now()}).now}

	for range 100 {
		if _, err := m.Admit(t.Context(), 7, "", free(t)); err != nil {
			t.Fatal(err)
		}
	}

	done := make(chan error)
	go func() { done <- m.Flush(t.Context()) }()
	<-st.started

	if u, _ := m.Admit(t.Context(), 7, "", free(t)); !u.Exhausted {
		t.Fatalf("request 101 admitted while the flush was in flight: %+v", u)
	}

	close(st.release)

	if err := <-done; err != nil {
		t.Fatal(err)
	}
}
