package scheduler

import (
	"context"

	"github.com/AlexGalhardo/hexascore-app/app/metrics"
)

// Counted wraps a scheduled task so each run increments hexa.schedule.runs.
// vukem's Command and Job build their closures inside the framework, so those
// tasks are counted only when registered through a function this wrapper can see.
func Counted(task string, fn func(context.Context) error) func(context.Context) error {
	return func(ctx context.Context) error {
		err := fn(ctx)
		outcome := "ok"
		if err != nil {
			outcome = "error"
		}

		metrics.RecordSchedule(ctx, task, outcome)

		return err
	}
}
