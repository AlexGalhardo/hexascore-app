package ssr

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSuperviseRestartsUntilCancelled(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "runs")
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		// Each run appends one byte and exits at once, like a crashing server.
		Supervise(ctx, 10*time.Millisecond, os.Args[0], "-test.run=TestHelperAppend", "--", marker)
	}()

	deadline := time.Now().Add(5 * time.Second)
	for {
		b, _ := os.ReadFile(marker)
		if len(b) >= 3 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("restarted %d times, want >= 3", len(b))
		}
		time.Sleep(20 * time.Millisecond)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Supervise did not return after cancel")
	}
}

// TestHelperAppend is the child process for TestSuperviseRestartsUntilCancelled.
func TestHelperAppend(t *testing.T) {
	args := os.Args
	for i, a := range args {
		if a == "--" && i+1 < len(args) {
			f, err := os.OpenFile(args[i+1], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
			if err != nil {
				t.Fatal(err)
			}
			_, _ = f.Write([]byte{1})
			_ = f.Close()
			return
		}
	}
}
