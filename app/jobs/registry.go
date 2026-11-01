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
	return AllWith(app, IngestSources{})
}

// AllWith is All with the ingestion sources replaced; only tests pass a non-zero value.
func AllWith(app *vukem.App, ingestSrc IngestSources) []queue.Registration {
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
		handlePruneWebhookDeliveries(app),
		handleCommunityHourly(app),
		handleCommunityAsserts(app),
		handleCheckSources(app),
		handleCoverFetch(app),
		handleExportIDs(app),
		handleComputeSimilar(app),
		handleRecommendUser(app),
		handleRecoEvalMonthly(app),
		handleComputeTrending(app),
		handleTrendingRebuild(app),
		handlePruneTrending(app),
		handleCatalogChanged(app),
		handleCommunityNotify(app),
		handleItemsChanged(app),
		handleIngest(app, ingestSrc),
	}
}
