package related

import "testing"

func TestMetrics(t *testing.T) {
	t.Parallel()

	m := NewMetrics()
	m.Impression("similar", "game")
	m.Impression("similar", "game")
	m.Click("similar", "game")
	m.Click("bogus", "game")

	if m.Impressions("similar", "game") != 2 || m.Clicks("similar", "game") != 1 || m.Clicks("bogus", "game") != 0 {
		t.Fatalf("impressions=%d clicks=%d", m.Impressions("similar", "game"), m.Clicks("similar", "game"))
	}
}

func TestFromSurface(t *testing.T) {
	t.Parallel()

	for in, want := range map[string]string{"similar": "similar", "crossmedia": "crossmedia", "trending": "", "": ""} {
		if got := FromSurface(in); got != want {
			t.Errorf("FromSurface(%q) = %q, want %q", in, got, want)
		}
	}
}
