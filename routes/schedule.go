package routes

import (
	"context"
	"errors"
	"time"

	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/schedule"

	"github.com/AlexGalhardo/hexascore-app/app/jobs"
	"github.com/AlexGalhardo/hexascore-app/app/recommend/primitives"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

// Schedule declares the scheduled tasks. "vukem serve" runs them in process;
// with SCHEDULE_IN_SERVE=false run "vukem schedule:work" or a cron entry
// calling "vukem schedule:run" every minute instead. List them with
// "vukem schedule:list". Tasks must honor ctx: on shutdown it is cancelled
// after SHUTDOWN_TIMEOUT. Name the parameters (s, app) with the first task
// you add. Examples:
//
//	s.Command("reports:send").DailyAt("08:00").WithoutOverlapping()
//	s.Job(jobs.PruneSessions{}).Hourly()
func Schedule(s *schedule.Schedule, app *vukem.App) {
	s.Command("search:purge").DailyAt("03:30").WithoutOverlapping()
	// A crash loses at most this minute of views and clicks (SDD 08 NFR-08-4); app shutdown flushes too.
	// workaround(vukem): VF-021, no buffered counter in the framework.
	s.Func("trending:flush", primitives.For(app).Flush).EveryMinute()
	s.Job(jobs.ComputeTrending{}).Cron("5 * * * *").WithoutOverlapping()
	// 04:45 stays clear of the 04:40 and 04:50 prunes.
	s.Job(jobs.PruneTrending{}).DailyAt("04:45").WithoutOverlapping()
	s.Job(jobs.ComputeSimilar{}).DailyAt("02:00").WithoutOverlapping()
	s.Job(jobs.RecoEvalMonthly{}).MonthlyOn(1, "05:30").WithoutOverlapping()
	s.Job(jobs.RefreshEntityStats{}).Hourly().WithoutOverlapping()
	s.Job(jobs.SyncSearchSuggest{}).EveryMinute().WithoutOverlapping()
	s.Job(jobs.DeletionPurge{}).DailyAt("03:00").WithoutOverlapping()
	s.Job(jobs.RetainPersonalData{}).DailyAt("03:15").WithoutOverlapping()
	s.Job(jobs.RecomputeScores{}).DailyAt("04:00").WithoutOverlapping()
	s.Job(jobs.ReleaseBursts{}).Hourly().WithoutOverlapping()
	s.Job(jobs.PruneRatingEvents{}).DailyAt("04:30").WithoutOverlapping()
	// api_changes is kept 30 days (SDD 05 §5). 04:50 stays clear of the 04:40 webhook-delivery prune.
	s.Job(jobs.PruneAPIChanges{}).DailyAt("04:50").WithoutOverlapping()
	// Fan-out reads api_changes after webhook_cursor (SDD 05 §7.3).
	s.Job(jobs.FanoutWebhooks{}).EveryMinute().WithoutOverlapping()
	// webhook_deliveries is kept 30 days (SDD 05 §5). 04:40 stays clear of the 04:50 api_changes prune.
	s.Job(jobs.PruneWebhookDeliveries{}).DailyAt("04:40").WithoutOverlapping()
	s.Job(jobs.ExpireExports{}).Hourly().WithoutOverlapping()
	s.Job(jobs.CommunityHourly{}).Hourly().WithoutOverlapping()
	// After the 04:00 score recompute, clear of the 04:30-04:50 prunes.
	s.Job(jobs.CommunityAsserts{}).DailyAt("05:10").WithoutOverlapping()
	s.Job(jobs.ExportIDs{}).DailyAt("07:00").WithoutOverlapping()
	// backup.daily (SDD 11 §7.3 step 1): moves after the SDD 06 asserts once they exist (P6).
	s.Command("backup:run", "--kind", "daily").DailyAt("06:00").In(saoPaulo).WithoutOverlapping()

	// Ingestion (SDD 06 §7.6). The weekly full runs of Wed-Sat sources join when their connectors exist.
	// A task starts only the first stage; each finished stage queues the next (ingest.Advance).
	s.Func("ingest:wikidata-incremental", startChain(app, ingest.WikidataIncremental)).
		DailyAt("00:30").In(saoPaulo).WithoutOverlapping()
	s.Func("ingest:wikidata-weekly", startChain(app, ingest.WikidataWeekly)).
		WeeklyOn(time.Sunday, "01:00").In(saoPaulo).WithoutOverlapping()
	s.Func("ingest:wikipedia-weekly", startChain(app, ingest.WikipediaWeekly)).
		WeeklyOn(time.Monday, "01:00").In(saoPaulo).WithoutOverlapping()
	s.Func("ingest:openlibrary-weekly", startChain(app, ingest.OpenLibraryWeekly)).
		WeeklyOn(time.Tuesday, "01:00").In(saoPaulo).WithoutOverlapping()
	s.Func("ingest:tmdb-weekly", startChain(app, ingest.TMDBWeekly)).
		WeeklyOn(time.Friday, "01:00").In(saoPaulo).WithoutOverlapping()
	s.Func("ingest:tmdb-expire", startChain(app, ingest.TMDBExpire)).
		DailyAt("03:00").In(saoPaulo).WithoutOverlapping()

	// backup.drill (SDD 11 §7.4): staging only. Staging also runs APP_ENV=production, so the
	// signal is the source bucket only staging is given. Cron ORs day-of-month with
	// day-of-week, so the command itself skips Sundays after the 7th.
	if app.Source.Get("BACKUP_SOURCE_S3_BUCKET") != "" {
		s.Command("backup:drill", "--first-sunday").WeeklyOn(time.Sunday, "07:00").In(saoPaulo).WithoutOverlapping()
	}
}

// saoPaulo is the operator's zone; Brazil has had no DST since 2019, so the
// fixed fallback is exact when the image lacks tzdata.
var saoPaulo = func() *time.Location {
	if loc, err := time.LoadLocation("America/Sao_Paulo"); err == nil {
		return loc
	}

	return time.FixedZone("BRT", -3*60*60)
}()

// startChain queues the first stage of c unless its kill switch is off or the stage is still active from the
// previous period; both are normal, not errors.
func startChain(app *vukem.App, c ingest.Chain) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		ff, err := app.Flags()
		if err != nil {
			return err
		}

		if !run.Allowed(ctx, ff, c.Stages[0].Source) {
			return nil
		}

		db, err := app.DB()
		if err != nil {
			return err
		}

		q, err := app.Queue()
		if err != nil {
			return err
		}

		if _, err := c.Start(ctx, db, q, "schedule", nil); err != nil && !errors.Is(err, run.ErrRunActive) {
			return err
		}

		return nil
	}
}
