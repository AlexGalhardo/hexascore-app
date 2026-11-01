package alerts

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/queue"
)

func capture() (*slog.Logger, *bytes.Buffer) {
	var b bytes.Buffer

	return slog.New(slog.NewTextHandler(&b, &slog.HandlerOptions{Level: slog.LevelDebug})), &b
}

func failed(kind string, final bool) queue.JobEvent {
	return queue.JobEvent{Kind: kind, Status: queue.JobFailed, Final: final, Err: errors.New("secret user@example.com")}
}

func TestAlert_JobFailedTwice(t *testing.T) {
	log, out := capture()
	o := &JobFailures{Log: log}
	ctx := context.Background()

	o.ObserveJob(ctx, failed("a", false))
	o.ObserveJob(ctx, failed("a", true))
	if out.Len() != 0 {
		t.Fatalf("one failed run must not alert: %s", out)
	}

	o.ObserveJob(ctx, failed("a", true))
	line := out.String()
	for _, want := range []string{"level=ERROR", "alert=A6", "kind=a"} {
		if !strings.Contains(line, want) {
			t.Errorf("log %q lacks %q", line, want)
		}
	}
	if strings.Contains(line, "user@example.com") {
		t.Errorf("log leaks the error text: %s", line)
	}
}

func TestAlert_JobFailedTwiceResetsOnSuccess(t *testing.T) {
	log, out := capture()
	o := &JobFailures{Log: log}
	ctx := context.Background()

	o.ObserveJob(ctx, failed("a", true))
	o.ObserveJob(ctx, queue.JobEvent{Kind: "a", Status: queue.JobProcessed})
	o.ObserveJob(ctx, failed("a", true))
	o.ObserveJob(ctx, failed("b", true))
	if out.Len() != 0 {
		t.Fatalf("no alert expected, got %s", out)
	}
}
