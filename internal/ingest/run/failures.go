package run

import (
	"context"

	"github.com/AlexGalhardo/vukem/database"
)

// ParkAfter is how many failed runs park a target (SDD 06 §7.6 "Dead letter").
const ParkAfter = 5

// RecordFailure notes that target (source, externalID) failed in run runID and reports whether it is now parked.
// Failures inside one run count once; a parked target stays parked until ClearFailure.
func RecordFailure(
	ctx context.Context,
	ex database.Executor,
	runID int64,
	source, externalID, class string,
	cause error,
) (bool, error) {
	var parked bool

	err := ex.QueryRowContext(ctx, `INSERT INTO ingest_failures (source, external_id, last_run_id, error_class, error)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (source, external_id) DO UPDATE SET
  attempts = ingest_failures.attempts + (ingest_failures.last_run_id <> excluded.last_run_id)::int,
  parked = ingest_failures.parked OR
    ingest_failures.attempts + (ingest_failures.last_run_id <> excluded.last_run_id)::int >= $6,
  last_run_id = excluded.last_run_id, error_class = excluded.error_class, error = excluded.error, last_seen = now()
RETURNING parked`, source, externalID, runID, class, cause.Error(), ParkAfter).Scan(&parked)

	return parked, err
}

// ClearFailure forgets a target after it succeeded (or after an operator un-parks it).
func ClearFailure(ctx context.Context, ex database.Executor, source, externalID string) error {
	_, err := ex.ExecContext(
		ctx,
		`DELETE FROM ingest_failures WHERE source = $1 AND external_id = $2`,
		source,
		externalID,
	)

	return err
}

// IsParked reports whether the target is parked and must be skipped by regular runs.
func IsParked(ctx context.Context, ex database.Executor, source, externalID string) (bool, error) {
	var parked bool

	err := ex.QueryRowContext(ctx, `SELECT coalesce(bool_or(parked), false) FROM ingest_failures
WHERE source = $1 AND external_id = $2`, source, externalID).Scan(&parked)

	return parked, err
}
