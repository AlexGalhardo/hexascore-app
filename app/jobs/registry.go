// Package jobs holds the queued jobs.
package jobs

import (
	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/queue"
)

// All lists the job handlers, built from the app so they reach its
// database, clock and mailer; vukem make:job appends here and names the
// parameter "app" with the first handler that needs it.
func All(app *vukem.App) []queue.Registration {
	return []queue.Registration{
		handleRefreshEntityStats(app),
		handleMerged(app),
		handleRecomputeScores(app),
		handleRecomputeItemScore(app),
		handleReleaseBursts(app),
		handlePruneRatingEvents(app),
		handleSyncSearchSuggest(app),
		handleDeletionPurge(app),
		handleLibraryExport(app),
		handleExpireExports(app),
		handleRetainPersonalData(app),
		handlePruneAPIChanges(app),
		handleFanoutWebhooks(app),
		handleWebhookDeliver(app),
	}
}
