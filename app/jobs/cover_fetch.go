package jobs

import (
	"context"
	"errors"
	"sync"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/queue"
	"github.com/AlexGalhardo/vukem/storage"
	"github.com/AlexGalhardo/vukem/upload"

	"github.com/AlexGalhardo/hexascore-app/app/community"
)

// CoverFetch downloads one https cover and stores it with the cover rules.
type CoverFetch struct {
	SuggestionID string `json:"suggestion_id"`
	URL          string `json:"url"`
}

// Kind names the job on the queue.
func (CoverFetch) Kind() string { return "community.cover_fetch" }

func handleCoverFetch(app *vukem.App) queue.Registration {
	get := sync.OnceValues(func() (*storage.Local, error) { return communityCovers(app) })

	return queue.Handle(func(ctx context.Context, job CoverFetch) error {
		db, err := app.DB()
		if err != nil {
			return err
		}

		disk, err := get()
		if err != nil {
			return err
		}

		cover, err := community.FetchCover(ctx, disk, job.URL)
		status, terminal := coverOutcome(err)
		if !terminal {
			return err
		}

		return community.RememberCover(ctx, db, job.SuggestionID, status, cover)
	})
}

// coverOutcome maps a terminal cover failure onto a meta status. A transport
// error is not terminal, so the queue retries it.
func coverOutcome(err error) (string, bool) {
	switch {
	case err == nil:
		return "ok", true
	case errors.Is(err, community.ErrCoverBlocked):
		return "blocked", true
	case errors.Is(err, community.ErrCoverRefused),
		errors.Is(err, upload.ErrFileTooLarge),
		errors.Is(err, errs.ErrRequestTooLarge):
		return "refused", true
	}

	if _, ok := errors.AsType[*errs.ValidationError](err); ok {
		return "refused", true
	}

	return "", false
}
