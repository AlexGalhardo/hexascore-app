//go:build integration

package offers_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/offers"
)

type fakeProber map[string]offers.Probe

func (f fakeProber) Probe(_ context.Context, url string) offers.Probe {
	if p, ok := f[url]; ok {
		return p
	}

	return offers.Probe{Err: errors.New("unexpected probe of " + url)}
}

type state struct {
	status                  string
	broken                  int
	active                  bool
	verified, checked, http bool
}

func (r *rig) state(item, store string) state {
	r.t.Helper()

	var (
		s                 state
		verified, checked *time.Time
		httpStatus        *int
	)

	err := r.db.QueryRowContext(r.ctx, `SELECT verify_status, broken_runs, active, last_verified_at, last_checked_at,
verify_http_status FROM offers WHERE item_id = $1 AND store = $2`, item, store).
		Scan(&s.status, &s.broken, &s.active, &verified, &checked, &httpStatus)
	if err != nil {
		r.t.Fatal(err)
	}

	s.verified, s.checked, s.http = verified != nil, checked != nil, httpStatus != nil

	return s
}

func (r *rig) offer(item, store, kind, url, source string) {
	r.t.Helper()
	r.exec(`INSERT INTO offers (item_id, country, store, kind, url, source, licence)
VALUES ($1, 'ZZ', $2, $3, $4, $5, 'x')`, item, store, kind, url, source)
}

func TestVerifySweep(t *testing.T) {
	r := newRig(t)
	game := r.item("game", "public", `{}`)

	const (
		steam = "https://store.steampowered.com/app/1"
		gog   = "https://www.gog.com/en/game/x"
		epic  = "https://store.epicgames.com/en-US/p/y"
	)

	r.offer(game, "steam", "buy_digital", steam, "wikidata")
	r.offer(game, "gog", "buy_digital", gog, "wikidata")
	r.offer(game, "epic", "buy_digital", epic, "wikidata")
	r.offer(game, "amazon_br", "buy", "https://www.amazon.com.br/s?k=1", "editions") // never probed
	r.offer(game, "mercadolivre", "buy", "https://example.com/not-a-store", "admin") // host outside the list
	r.offer(game, "google_play", "mobile", "https://play.google.com/store/apps/details?id=a.b", "tmdb_providers")

	pr := fakeProber{
		steam: {Status: 200, FinalURL: steam},
		gog:   {Status: 404},
		epic:  {Status: 503},
	}

	start := time.Date(2026, 11, 3, 2, 45, 0, 0, time.UTC)

	st, err := offers.Verify(r.ctx, r.db, pr, start, start, 100)
	if err != nil {
		t.Fatal(err)
	}

	if want := (offers.Stats{Checked: 3, OK: 1, Broken: 1, Transient: 1}); st != want {
		t.Fatalf("stats = %+v, want %+v", st, want)
	}

	if s := r.state(game, "steam"); s.status != "ok" || !s.verified || !s.checked || !s.active {
		t.Fatalf("steam = %+v", s)
	}

	if s := r.state(game, "gog"); s.status != "broken" || s.broken != 1 || !s.active || !s.verified {
		t.Fatalf("gog after one 404 = %+v", s)
	}

	// a timeout or 5xx proves nothing: attempted, never verified
	if s := r.state(game, "epic"); s.status != "unverified" || s.verified || !s.checked || s.http {
		t.Fatalf("epic after a 503 = %+v", s)
	}

	for _, store := range []string{"amazon_br", "mercadolivre", "google_play"} {
		if s := r.state(game, store); s.checked || s.verified {
			t.Fatalf("%s must not be probed: %+v", store, s)
		}
	}

	// the same run asks for nothing twice, even after a 503
	if st, err := offers.Verify(r.ctx, r.db, pr, start, start.Add(time.Minute), 100); err != nil || st.Checked != 0 {
		t.Fatalf("second call in the run: %+v, %v", st, err)
	}

	// the next day: a second 404 in a row deactivates the offer
	next := start.AddDate(0, 0, 1)
	pr[epic] = offers.Probe{Status: 200, FinalURL: epic}

	if _, err := offers.Verify(r.ctx, r.db, pr, next, next, 100); err != nil {
		t.Fatal(err)
	}

	if s := r.state(game, "gog"); s.status != "gone" || s.broken != 2 || s.active {
		t.Fatalf("gog after two 404s = %+v", s)
	}

	if s := r.state(game, "epic"); s.status != "ok" || !s.verified {
		t.Fatalf("epic once it answers = %+v", s)
	}

	// a deactivated offer is not asked about again
	if st, err := offers.Verify(
		r.ctx,
		r.db,
		pr,
		next.AddDate(0, 0, 1),
		next.AddDate(0, 0, 1),
		100,
	); err != nil ||
		st.Checked != 2 {
		t.Fatalf("third day: %+v, %v (steam and epic only)", st, err)
	}
}

func TestVerifyHonoursTheLimitAndOrdersByLeastRecentlyChecked(t *testing.T) {
	r := newRig(t)
	game := r.item("game", "public", `{}`)

	r.offer(game, "steam", "buy_digital", "https://store.steampowered.com/app/1", "wikidata")
	r.offer(game, "gog", "buy_digital", "https://www.gog.com/en/game/x", "wikidata")
	r.exec(`UPDATE offers SET last_checked_at = '2026-10-01' WHERE store = 'gog'`)

	pr := fakeProber{
		"https://store.steampowered.com/app/1": {Status: 200, FinalURL: "https://store.steampowered.com/app/1"},
		"https://www.gog.com/en/game/x":        {Status: 200, FinalURL: "https://www.gog.com/en/game/x"},
	}

	now := time.Date(2026, 11, 3, 2, 45, 0, 0, time.UTC)
	if st, err := offers.Verify(r.ctx, r.db, pr, now, now, 1); err != nil || st.Checked != 1 {
		t.Fatalf("%+v, %v", st, err)
	}

	// the never-checked offer goes first
	if !r.state(game, "steam").checked || r.state(game, "gog").verified {
		t.Fatal("the offer never checked must be probed before the one checked in October")
	}
}
