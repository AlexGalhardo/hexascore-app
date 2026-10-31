package webhooks

import (
	"context"
	"errors"
	"testing"
)

func TestPingRejectsMalformedID(t *testing.T) {
	_, err := Ping(context.Background(), nil, 1, "0")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Ping(%q) = %v, want ErrNotFound", "0", err)
	}
}
