package routes

import (
	"time"

	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/schedule"

	"github.com/AlexGalhardo/hexascore-app/app/jobs"
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
	// backup.daily (SDD 11 §7.3 step 1): moves after the SDD 06 asserts once they exist (P6).
	s.Command("backup:run", "--kind", "daily").DailyAt("06:00").In(saoPaulo).WithoutOverlapping()

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
