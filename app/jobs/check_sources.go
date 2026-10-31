package jobs

import (
	"context"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/app/community"
)

// CheckSources probes a suggestion's source URLs. A bad URL is stored in
// meta and does not fail the job, so the submission is never blocked on it.
type CheckSources struct {
	SuggestionID string `json:"suggestion_id"`
}

// Kind names the job on the queue.
func (CheckSources) Kind() string { return "community.check_sources" }

func handleCheckSources(app *vukem.App) queue.Registration {
	return queue.Handle(func(ctx context.Context, job CheckSources) error {
		db, err := app.DB()
		if err != nil {
			return err
		}

		return community.CheckSources(ctx, db, job.SuggestionID)
	})
}
