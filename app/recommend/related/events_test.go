package related

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func newEvents(buf *bytes.Buffer, every, pick int) Events {
	return Events{
		Log:             slog.New(slog.NewTextHandler(buf, nil)),
		ImpressionEvery: every,
		pick:            func(int) int { return pick },
	}
}

func TestEventsImpressionLogsStructuredLine(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	newEvents(&buf, 1, 0).Impression(context.Background(), "similar", "game", "item-1")

	for _, want := range []string{"event=reco_impression", "surface=similar", "type=game", "item=item-1", "level=INFO"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("line %q lacks %q", buf.String(), want)
		}
	}
}

func TestEventsImpressionSampling(t *testing.T) {
	t.Parallel()

	var kept, dropped bytes.Buffer
	newEvents(&kept, 10, 0).Impression(context.Background(), "similar", "game", "i")
	newEvents(&dropped, 10, 3).Impression(context.Background(), "similar", "game", "i")

	if !strings.Contains(kept.String(), "sample_rate=10") || dropped.Len() != 0 {
		t.Fatalf("kept=%q dropped=%q", kept.String(), dropped.String())
	}
}

func TestEventsClick(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	e := newEvents(&buf, 10, 5)
	e.Click(context.Background(), "similar", "game", "item-1")
	e.Click(context.Background(), "trending", "game", "item-1")

	if got := strings.Count(buf.String(), "event=reco_click"); got != 1 {
		t.Fatalf("clicks logged = %d, want 1 (unsampled, other surfaces ignored): %q", got, buf.String())
	}
}
