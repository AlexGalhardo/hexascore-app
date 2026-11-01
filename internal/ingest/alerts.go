package ingest

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/AlexGalhardo/vukem/database"
)

// a7Checks are the conditions of alert A7 (SDD 12 §7.5, SDD 06 §10) the app can read from its own tables. Each query
// returns how many rows offend. "Breaker open over 1 h" is not here: the breaker lives in the outbound client's memory
// and has no durable record yet.
var a7Checks = []struct {
	name, msg, query string
	window           time.Duration
}{
	{"failed_run", "ingestion run failed in the last 24 h", `SELECT count(*) FROM job_runs
WHERE status = 'failed' AND finished_at > $1::timestamptz`, 24 * time.Hour},
	{"no_incremental", "no incremental Wikidata run in 26 h", `SELECT CASE WHEN EXISTS (SELECT 1 FROM job_runs
WHERE job = 'fetch.wikidata' AND mode = 'incremental' AND status IN ('succeeded', 'partial')
AND finished_at > $1::timestamptz) THEN 0 ELSE 1 END`, 26 * time.Hour},
	{"error_assert", "the latest dq.asserts run has failing error asserts", `SELECT count(*) FROM dq_results
WHERE run_id = (SELECT max(d.run_id) FROM dq_results d JOIN job_runs j ON j.id = d.run_id WHERE j.job = 'dq.asserts')
AND NOT passed AND severity = 'error' AND $1::timestamptz IS NOT NULL`, 0},
	{"tmdb_age", "TMDB records older than 170 days", `SELECT count(*) FROM source_records
WHERE source = 'tmdb' AND checked_at < $1::timestamptz`, 170 * 24 * time.Hour},
}

// CheckA7 writes one alert=A7 line per condition that holds and returns how many did (SDD 12 §6.2: the app exports
// no metrics, an in-app check logs a stable key). enabled is the master kill switch: with ingestion off nothing
// runs, so nothing alerts. A check that cannot run logs state=unknown at ERROR and the rest still run; their errors are joined.
func CheckA7(ctx context.Context, ex database.Executor, log *slog.Logger, now time.Time, enabled bool) (int, error) {
	if !enabled {
		return 0, nil
	}

	var errs []error

	raised := 0

	for _, c := range a7Checks {
		var n int
		if err := ex.QueryRowContext(ctx, c.query, now.Add(-c.window)).Scan(&n); err != nil {
			log.LogAttrs(ctx, slog.LevelError, "ingestion alert check failed",
				slog.String("alert", "A7"), slog.String("check", c.name), slog.String("state", "unknown"))

			errs = append(errs, fmt.Errorf("a7 check %s: %w", c.name, err))

			continue
		}

		if n > 0 {
			raised++

			log.LogAttrs(ctx, slog.LevelWarn, c.msg,
				slog.String("alert", "A7"), slog.String("check", c.name), slog.Int("count", n))
		}
	}

	return raised, errors.Join(errs...)
}
