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

	// No wall-clock deadline: each restart spawns a test binary, whose start-up time
	// depends on machine load, not on Supervise. `go test -timeout` bounds a real hang.
	for {
		b, _ := os.ReadFile(marker)
		if len(b) >= 3 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	cancel()
	<-done
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
