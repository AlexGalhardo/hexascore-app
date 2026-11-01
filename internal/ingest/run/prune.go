package run

import (
	"context"
	"fmt"

	"github.com/AlexGalhardo/vukem/database"
)

// Retention windows in days (SDD 06 §5): history 30, job_runs 180, dq_results and field_coverage 365.
const (
	HistoryRetention = 30
	RunRetention     = 180
	ResultRetention  = 365
)

// Pruned counts the rows Prune deleted.
type Pruned struct{ History, Results, Coverage, Runs int }

// Prune applies the retention windows. dq_results and field_coverage reference job_runs and live longer, so a run
// older than 180 days stays until its results age out too.
func Prune(ctx context.Context, ex database.Executor) (Pruned, error) {
	var p Pruned

	steps := []struct {
		into *int
		sql  string
		days int
	}{
		{
			&p.History, `DELETE FROM source_record_history WHERE superseded_at < now() - make_interval(days => $1)`,
			HistoryRetention,
		},
		{&p.Results, `DELETE FROM dq_results d USING job_runs j WHERE j.id = d.run_id
AND j.created_at < now() - make_interval(days => $1)`, ResultRetention},
		{&p.Coverage, `DELETE FROM field_coverage f USING job_runs j WHERE j.id = f.run_id
AND j.created_at < now() - make_interval(days => $1)`, ResultRetention},
		{&p.Runs, `DELETE FROM job_runs j WHERE j.created_at < now() - make_interval(days => $1)
AND NOT EXISTS (SELECT 1 FROM dq_results WHERE run_id = j.id)
AND NOT EXISTS (SELECT 1 FROM field_coverage WHERE run_id = j.id)`, RunRetention},
	}

	for _, s := range steps {
		res, err := ex.ExecContext(ctx, s.sql, s.days)
		if err != nil {
			return p, fmt.Errorf("ingest prune: %w", err)
		}

		n, err := res.RowsAffected()
		if err != nil {
			return p, err
		}

		*s.into = int(n)
	}

	return p, nil
}
