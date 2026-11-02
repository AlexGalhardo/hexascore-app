package jobs

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/database"
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

		return runCoverFetch(ctx, db, disk, community.FetchCover, job)
	}, queue.Retries(3), queue.Timeout(time.Minute))
}

// runCoverFetch is idempotent on retry: a suggestion whose meta already carries a cover_status is skipped, and an
// object stored for an outcome that could not be saved is deleted so the retry leaves no orphan.
func runCoverFetch(
	ctx context.Context,
	db *database.DB,
	disk storage.Disk,
	fetch func(context.Context, storage.Disk, string) (community.Cover, error),
	job CoverFetch,
) error {
	var done bool
	err := db.QueryRowContext(ctx, `SELECT meta->>'cover_status' IS NOT NULL FROM suggestions WHERE id = $1`,
		job.SuggestionID).Scan(&done)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && done) {
		return nil
	}
	if err != nil {
		return err
	}

	cover, err := fetch(ctx, disk, job.URL)
	status, terminal := coverOutcome(err)
	if !terminal {
		return err
	}

	if err := community.RememberCover(ctx, db, job.SuggestionID, status, cover); err != nil {
		if cover.Key != "" {
			_ = disk.Delete(context.WithoutCancel(ctx), cover.Key)
		}

		return err
	}

	return nil
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
