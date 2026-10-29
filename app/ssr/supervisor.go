package ssr

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// StopGrace is how long Node gets to exit after SIGTERM before it is killed.
const StopGrace = 5 * time.Second

// Supervise runs name with args until ctx ends, restarting it delay after
// every exit. One container runs Go and the Node SSR server (ADR 0004), so
// the Go process owns Node's lifecycle instead of a second init system.
// When ctx ends the child gets SIGTERM, then SIGKILL after StopGrace; the
// caller cancels ctx only after the HTTP server drained, so in-flight
// renders finish first.
func Supervise(ctx context.Context, delay time.Duration, name string, args ...string) {
	for ctx.Err() == nil {
		//nolint:gosec // G702: name and args come from the app's bootstrap, never from a request
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		// SIGTERM is a no-op error on Windows (dev only); WaitDelay then kills.
		cmd.Cancel = func() error { return cmd.Process.Signal(syscall.SIGTERM) }
		cmd.WaitDelay = StopGrace
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
