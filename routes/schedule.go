package routes

import (
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
func Schedule(s *schedule.Schedule, _ *vukem.App) {
	s.Command("search:purge").DailyAt("03:30").WithoutOverlapping()
	s.Job(jobs.RefreshEntityStats{}).Hourly().WithoutOverlapping()
	s.Job(jobs.SyncSearchSuggest{}).EveryMinute().WithoutOverlapping()
	s.Job(jobs.DeletionPurge{}).DailyAt("03:00").WithoutOverlapping()
	s.Job(jobs.RecomputeScores{}).DailyAt("04:00").WithoutOverlapping()
	s.Job(jobs.ReleaseBursts{}).Hourly().WithoutOverlapping()
	s.Job(jobs.PruneRatingEvents{}).DailyAt("04:30").WithoutOverlapping()
}
