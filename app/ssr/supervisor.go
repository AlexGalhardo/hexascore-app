package ssr

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"time"
)

// Supervise runs name with args until ctx ends, restarting it delay after
// every exit. One container runs Go and the Node SSR server (ADR 0004), so
// the Go process owns Node's lifecycle instead of a second init system.
func Supervise(ctx context.Context, delay time.Duration, name string, args ...string) {
	for ctx.Err() == nil {
		//nolint:gosec // G702: name and args come from the app's bootstrap, never from a request
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		err := cmd.Run()
		if ctx.Err() != nil {
			return
		}
		slog.WarnContext(ctx, "ssr server exited, restarting", "error", err, "delay", delay)
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}
	}
}
