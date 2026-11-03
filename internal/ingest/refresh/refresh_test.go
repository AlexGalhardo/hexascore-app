package refresh_test

import (
	"context"
	"errors"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/refresh"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

func TestCancelledRefreshDoesNotStartBatch(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := refresh.Step(refresh.Config{})(ctx, nil, run.Run{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled refresh = %v, want context.Canceled", err)
	}
}
