package ssr

import (
	"context"
	"testing"

	"github.com/AlexGalhardo/vukem/inertia"
)

func TestFailuresCountsEachFallback(t *testing.T) {
	var f Failures
	for range 3 {
		f.Observe(context.Background(), inertia.SSRError{Type: "connection"})
	}
	if got := f.Count(); got != 3 {
		t.Fatalf("Count() = %d, want 3", got)
	}
}
