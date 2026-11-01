package config_test

import (
	"io/fs"
	"testing"
	"testing/fstest"

	vukem "github.com/AlexGalhardo/vukem"
	vukemconfig "github.com/AlexGalhardo/vukem/config"

	"github.com/AlexGalhardo/hexascore-app/config"
)

func queueLimits(t *testing.T, env map[string]string) map[string]int {
	t.Helper()

	getenv := config.WithQueueLimits(func(k string) (string, bool) {
		v, ok := env[k]
		return v, ok
	})

	noFile := func(string) ([]byte, error) { return nil, fs.ErrNotExist }

	src, err := vukemconfig.LoadSource(fstest.MapFS{}, getenv, noFile)
	if err != nil {
		t.Fatal(err)
	}

	qc, err := vukemconfig.Load[vukem.QueueConfig](src)
	if err != nil {
		t.Fatal(err)
	}

	limits, err := qc.QueueLimits()
	if err != nil {
		t.Fatal(err)
	}

	return limits
}

// NFR-06-2: ingestion is capped at 2 jobs whatever QUEUE_CONCURRENCY is, even when the operator sets nothing.
func TestQueueLimitsDefaultCapsIngest(t *testing.T) {
	for name, env := range map[string]map[string]string{
		"unset": {"QUEUE_CONCURRENCY": "16"},
		"empty": {"QUEUE_LIMITS": ""},
	} {
		t.Run(name, func(t *testing.T) {
			got := queueLimits(t, env)
			if len(got) != 1 || got["ingest"] != 2 {
				t.Fatalf("limits = %v, want map[ingest:2]", got)
			}
		})
	}
}

func TestQueueLimitsOperatorOverrideWins(t *testing.T) {
	got := queueLimits(t, map[string]string{"QUEUE_LIMITS": "ingest=1,import=3"})
	if got["ingest"] != 1 || got["import"] != 3 || len(got) != 2 {
		t.Fatalf("limits = %v, want ingest=1 import=3", got)
	}
}
