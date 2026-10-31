package jobs

import (
	"context"
	"errors"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/app/community"
)

// errNotYet keeps the job in the failed set for a replay instead of acking work nobody did.
var errNotYet = errors.New("jobs: handler lands in a later ROADMAP row")

// handleCatalogChanged registers `catalog.changed` so the apply transaction can dispatch it.
// shortcut: no purge, reindex or sitemap yet; P4.B.2.2 replaces this body.
func handleCatalogChanged(_ *vukem.App) queue.Registration {
	return queue.Handle(func(context.Context, community.CatalogChanged) error { return errNotYet }, queue.Retries(0))
}

// handleCommunityNotify registers `community.notify` so the apply transaction can dispatch it.
// shortcut: no inbox row or mail yet; P4.B.3.1 replaces this body, failed jobs can be replayed then.
func handleCommunityNotify(_ *vukem.App) queue.Registration {
	return queue.Handle(func(context.Context, community.Notify) error { return errNotYet }, queue.Retries(0))
}
