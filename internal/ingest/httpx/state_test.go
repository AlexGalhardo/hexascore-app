package httpx

import (
	"context"
	"errors"
	"maps"
	"sync"
	"testing"

	"github.com/AlexGalhardo/vukem/httpclient"
)

type memStore struct {
	mu sync.Mutex
	m  map[string]HostState
}

func (s *memStore) Load(context.Context) (map[string]HostState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := map[string]HostState{}
	maps.Copy(out, s.m)

	return out, nil
}

func (s *memStore) Save(_ context.Context, host string, st HostState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.m == nil {
		s.m = map[string]HostState{}
	}

	s.m[host] = st

	return nil
}

func withStore(st StateStore, clk *fakeClock, sc *script) *Client {
	return New(Config{
		ContactURL: "https://hexascore.app/bot",
		Hosts:      map[string]HostLimit{"api.test": {PerSecond: 5}},
		Clock:      clk, Base: sc, State: st,
	})
}

func TestStateSavedOnBreakerOpenAndRestored(t *testing.T) {
	st := &memStore{}
	clk := newClock()
	c := withStore(st, clk, &script{steps: []step{{status: 500}}, clock: clk})

	for range 12 {
		_, _ = get(t, c, "api.test")
	}

	saved := st.m["api.test"]
	if saved.OpenFor == 0 || !saved.OpenUntil.After(clk.Now()) {
		t.Fatalf("breaker not stored: %+v", saved)
	}

	next := withStore(st, clk, &script{steps: []step{{status: 200}}, clock: clk})
	if err := next.Restore(t.Context()); err != nil {
		t.Fatal(err)
	}

	if _, err := get(t, next, "api.test"); !errors.Is(err, httpclient.ErrCircuitOpen) {
		t.Fatalf("restored client err = %v, want ErrCircuitOpen", err)
	}
}

func TestStateStoresHalvedRateAndCalls(t *testing.T) {
	st := &memStore{}
	clk := newClock()
	c := withStore(st, clk, &script{steps: []step{{status: 200}}, clock: clk})

	for range 3 {
		_, _ = get(t, c, "api.test")
	}

	h := c.policyHost("api.test")
	for range 20 {
		h.rateLimited(clk.Now())
	}

	if err := c.Flush(t.Context()); err != nil {
		t.Fatal(err)
	}

	if got := st.m["api.test"]; got.Rate != 2.5 || got.CallsToday != 3 || got.Last429At.IsZero() {
		t.Fatalf("stored %+v, want rate 2.5, 3 calls, a last 429", got)
	}

	next := withStore(st, clk, &script{steps: []step{{status: 200}}, clock: clk})
	if err := next.Restore(t.Context()); err != nil {
		t.Fatal(err)
	}

	if got := next.policyHost("api.test"); got.rate != 2.5 || got.calls != 3 {
		t.Fatalf("restored rate %v calls %d, want 2.5 and 3", got.rate, got.calls)
	}
}
