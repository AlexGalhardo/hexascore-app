package tests

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/cache"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/challenge"
	"github.com/AlexGalhardo/hexascore-app/app/forms"
	appmiddleware "github.com/AlexGalhardo/hexascore-app/app/http/middleware"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// TestSecurity_A07_HexaHoneypotAndFillTime: a filled honeypot, a form sent under 3 s after render, or a
// missing or forged form token all get the normal answer and store nothing (SDD 13 §7.2).
func TestSecurity_A07_HexaHoneypotAndFillTime(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)
	c := app.Client(t).Inertia("")

	c.Get("/signup").AssertOK().AssertInertia("Auth/Signup").HasProp("formToken")

	key, err := app.DeriveKey("signup.form", 32)
	if err != nil {
		t.Fatal(err)
	}

	tok := forms.Token{Key: key, Now: app.Now}

	for name, edit := range map[string]func(map[string]any){
		"honeypot":   func(f map[string]any) { f["website"] = "http://spam.example" },
		"too fast":   func(f map[string]any) { f["form_token"] = tok.IssueAt("signup", app.Now().Add(-time.Second)) },
		"no token":   func(f map[string]any) { delete(f, "form_token") },
		"other form": func(f map[string]any) { f["form_token"] = tok.IssueAt("contact", app.Now().Add(-time.Minute)) },
		"forged":     func(f map[string]any) { f["form_token"] = "1.x" },
	} {
		// One address per case: the per-address sign-up limit is 5 an hour.
		email := strings.ReplaceAll(name, " ", "-") + "@example.com"
		form := signUpForm(app, "Bot Net", email, testPassword)
		edit(form)
		// Six sign-ups from one IP pass the PoW threshold: solve each time.
		c.WithHeader(middlewareHeader, solved(t, c, "signup")).PostJSON("/signup", form).
			AssertRedirect("/signup/check-email")

		vukemtest.AssertDatabaseMissing(t, app, "users", map[string]any{"email": email})

		if t.Failed() {
			t.Fatalf("case %q stored the user", name)
		}
	}

	c.WithHeader(middlewareHeader, solved(t, c, "signup")).
		PostJSON("/signup", signUpForm(app, "Bot Net", "honey@example.com", testPassword)).
		AssertRedirect("/signup/check-email")
	vukemtest.AssertDatabaseHas(t, app, "users", map[string]any{"email": "honey@example.com"})
}

// TestSecurity_A07_PowFallback: after 10 s a slow client may ask once per hour for bits-2.
func TestSecurity_A07_PowFallback(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)
	c := app.Client(t).Inertia("")

	var quiet powChallenge
	c.GetJSON("/challenge?route=signup&fallback=1").AssertOK().JSON(&quiet)

	if quiet.Required || quiet.Bits != challenge.DefaultBits {
		t.Fatalf("a quiet client spent the fallback: %+v", quiet)
	}

	for i := range 3 {
		signupAs(app, c, i).AssertRedirect("/signup/check-email")
	}

	var easy, again powChallenge
	c.GetJSON("/challenge?route=signup&fallback=1").AssertOK().JSON(&easy)
	c.GetJSON("/challenge?route=signup&fallback=1").AssertOK().JSON(&again)

	if !easy.Required || easy.Bits != challenge.DefaultBits-appmiddleware.FallbackDrop ||
		again.Bits != challenge.DefaultBits {
		t.Fatalf("fallback = %+v, second = %+v", easy, again)
	}

	proof := easy.C + ":" + challenge.Solve(challenge.Challenge{C: easy.C, Bits: easy.Bits, Exp: easy.Exp})
	signupAs(app, c.WithHeader(middlewareHeader, proof), 3).AssertRedirect("/signup/check-email")
}

// ttlStore records the TTL of every write, so the 1 h trip hold is visible without waiting.
type ttlStore struct {
	*cache.Memory

	mu   sync.Mutex
	ttls map[string]time.Duration
}

func (s *ttlStore) Set(ctx context.Context, key string, v []byte, ttl time.Duration) error {
	s.mu.Lock()
	s.ttls[key] = ttl
	s.mu.Unlock()

	return s.Memory.Set(ctx, key, v, ttl)
}

func (s *ttlStore) ttl(key string) time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.ttls[key]
}

func challengeFrom(ip, proof string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/signup", nil)
	r.RemoteAddr = ip + ":4000"

	if proof != "" {
		r.Header.Set(middlewareHeader, proof)
	}

	return r
}

func testChallenger(store cache.Store, powStore func() (*cache.Cache, error)) *appmiddleware.Challenger {
	c := cache.New(store, "t:")
	counters := func() (*cache.Cache, error) { return c, nil }

	if powStore == nil {
		powStore = counters
	}

	return &appmiddleware.Challenger{
		PoW:   challenge.New([]byte("0123456789abcdef0123456789abcdef"), powStore, time.Now),
		Cache: counters,
	}
}

var okHandler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

// TestSecurity_A07_PowGlobalTrip: past 30 sign-ups an hour from anywhere every client needs 21 bits,
// and the trip holds for 1 h after the counter window ends.
func TestSecurity_A07_PowGlobalTrip(t *testing.T) {
	t.Parallel()

	store := &ttlStore{Memory: cache.NewMemory(0), ttls: map[string]time.Duration{}}
	h := testChallenger(store, nil).Require("signup")(okHandler)

	post := func(ip string) (code, bits int) {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, challengeFrom(ip, ""))

		var p struct{ Challenge struct{ Bits int } }
		_ = json.Unmarshal(rec.Body.Bytes(), &p)

		return rec.Code, p.Challenge.Bits
	}

	for i := range 30 {
		if code, _ := post(fmt.Sprintf("198.51.100.%d", i)); code != http.StatusOK {
			t.Fatalf("sign-up %d: status %d", i+1, code)
		}
	}

	if code, bits := post("203.0.113.1"); code != http.StatusPreconditionRequired || bits != challenge.AttackBits {
		t.Fatalf("31st sign-up: status %d, bits %d; want 428 with %d bits", code, bits, challenge.AttackBits)
	}

	if got := store.ttl("t:challenge:signup:all:trip"); got != appmiddleware.TripHold {
		t.Fatalf("trip flag TTL = %v, want %v", got, appmiddleware.TripHold)
	}

	// The counter window ends; the flag keeps the challenge on.
	if err := store.Delete(t.Context(), "t:challenge:signup:all"); err != nil {
		t.Fatal(err)
	}

	if code, bits := post("203.0.113.2"); code != http.StatusPreconditionRequired || bits != challenge.AttackBits {
		t.Fatalf("after the window: status %d, bits %d", code, bits)
	}

	// The hold ends: a quiet client passes again.
	if err := store.Delete(t.Context(), "t:challenge:signup:all", "t:challenge:signup:all:trip"); err != nil {
		t.Fatal(err)
	}

	if code, _ := post("203.0.113.3"); code != http.StatusOK {
		t.Fatalf("after the hold: status %d", code)
	}
}

// TestSecurity_A10_HexaPowStoreDown503: a broken single-use store fails closed with 503; a broken
// counter store lets requests through (rate limits still apply, NFR-03-9).
func TestSecurity_A10_HexaPowStoreDown503(t *testing.T) {
	t.Parallel()

	down := func() (*cache.Cache, error) { return nil, errors.New("store down") }
	ch := testChallenger(cache.NewMemory(0), down)
	h := ch.Require("signup")(okHandler)

	fresh, err := ch.PoW.Issue("signup", challenge.DefaultBits)
	if err != nil {
		t.Fatal(err)
	}

	for range 3 {
		h.ServeHTTP(httptest.NewRecorder(), challengeFrom("192.0.2.9", ""))
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, challengeFrom("192.0.2.9", fresh.C+":"+challenge.Solve(fresh)))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("solution store down: status %d, want 503", rec.Code)
	}

	rec = httptest.NewRecorder()
	(&appmiddleware.Challenger{PoW: ch.PoW, Cache: down}).Require("signup")(okHandler).
		ServeHTTP(rec, challengeFrom("192.0.2.9", ""))

	if rec.Code != http.StatusOK {
		t.Fatalf("counter store down: status %d, want 200", rec.Code)
	}
}
