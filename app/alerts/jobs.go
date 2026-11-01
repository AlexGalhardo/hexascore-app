package alerts

import (
	"context"
	"log/slog"
	"sync"

	"github.com/AlexGalhardo/vukem/queue"
)

// failedRunsToAlert is how many failed runs of one kind in a row raise A6.
const failedRunsToAlert = 2

// JobFailures is a queue.Observer that raises A6 when a job kind fails twice in a row. A run fails when its
// retries are exhausted (JobEvent.Final); a success of the kind resets the streak. The error text is not
// logged: it may carry user data.
type JobFailures struct {
	Log *slog.Logger

	mu      sync.Mutex
	streaks map[string]int
}

// ObserveJob implements queue.Observer.
func (j *JobFailures) ObserveJob(ctx context.Context, e queue.JobEvent) {
	j.mu.Lock()
	defer j.mu.Unlock()

	switch {
	case e.Status == queue.JobProcessed:
		delete(j.streaks, e.Kind)
	case e.Status == queue.JobFailed && e.Final:
		if j.streaks == nil {
			j.streaks = map[string]int{}
		}

		j.streaks[e.Kind]++
		if n := j.streaks[e.Kind]; n >= failedRunsToAlert {
			Error(ctx, j.Log, A6, "job failed twice in a row", "kind", e.Kind, "queue", e.Queue, "failed_runs", n)
		}
	}
}
