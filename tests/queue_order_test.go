//go:build integration

package tests

import (
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/queue"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/jobs"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// TestQueue_ObservedPriorityOrder records how the database driver picks across
// QUEUE_NAMES=default,ingest (open issue 11-7: undocumented in
// v0.32.0). No app code may rely on it; this test only flags a change.
func TestQueue_ObservedPriorityOrder(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{
		"DATABASE_URL": pgtest.Database(t, ""), "QUEUE_CONNECTION": "database",
	}})

	q, err := app.Queue()
	if err != nil {
		t.Fatal(err)
	}

	// ingest is dispatched first, so FIFO across queues would pop it first.
	for _, name := range []string{"ingest", "default"} {
		if err := q.Dispatch(t.Context(), jobs.RefreshEntityStats{}, queue.OnQueue(name)); err != nil {
			t.Fatal(err)
		}
	}

	var got []string
	for range 2 {
		m, err := q.Driver().Pop(t.Context(), []string{"default", "ingest"})
		if err != nil || m == nil {
			t.Fatalf("pop: %v, %v", m, err)
		}
		got = append(got, m.Queue)
	}

	t.Logf("observed order for QUEUE_NAMES=default,ingest: %v", got)
	if got[0] != "default" || got[1] != "ingest" {
		t.Errorf("order = %v, want [default ingest] (strict priority by QUEUE_NAMES position)", got)
	}
}
