package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/AlexGalhardo/vukem/database"
)

// ChunkSize is how many records share one transaction in Store (legacy store_records commit_every).
const ChunkSize = 200

// Record is one fetched source payload, keyed like source_records.
type Record struct {
	Source     string
	ExternalID string
	Payload    json.RawMessage
}

// Querier is the row-query half of *database.DB and *database.Tx.
type Querier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// upsertSQL writes the previous version to source_record_history in the same statement as the update.
// $1 source, $2 external_id, $3 payload, $4 hash, $5 run id. An unchanged hash only moves checked_at and
// last_run_id; the changed and touched CTEs are mutually exclusive on the hash, so no row is hit twice.
// old takes no FOR UPDATE: the lock skips a row that upd rewrites in the same command, so hist would read
// nothing. One active run per job (job_runs_one_active) is the single-writer guarantee.
// superseded_at uses clock_timestamp() so two changes of one record in one transaction do not collide.
const upsertSQL = `
WITH old AS (
  SELECT source, external_id, hash, payload, fetched_at FROM source_records
  WHERE source = $1 AND external_id = $2
), upd AS (
  INSERT INTO source_records (source, external_id, payload, hash, fetched_at, checked_at, last_run_id)
  VALUES ($1, $2, $3::jsonb, $4, now(), now(), $5)
  ON CONFLICT (source, external_id) DO UPDATE
  SET payload = excluded.payload, hash = excluded.hash, fetched_at = excluded.fetched_at,
      checked_at = excluded.checked_at, last_run_id = excluded.last_run_id
  WHERE source_records.hash <> excluded.hash
  RETURNING 1
), touch AS (
  UPDATE source_records SET checked_at = now(), last_run_id = $5
  WHERE source = $1 AND external_id = $2 AND hash = $4
  RETURNING 1
), hist AS (
  INSERT INTO source_record_history (source, external_id, hash, payload, fetched_at, superseded_at, superseded_by_run)
  SELECT o.source, o.external_id, o.hash, o.payload, o.fetched_at, clock_timestamp(), $5
  FROM old o WHERE o.hash <> $4
  RETURNING 1
)
SELECT (SELECT count(*) FROM upd), (SELECT count(*) FROM hist), (SELECT count(*) FROM touch)`

// Upsert stores rec under run runID and reports whether the row was inserted or its payload changed.
// An unchanged payload only refreshes checked_at (the TMDB expiry clock) and last_run_id.
func Upsert(ctx context.Context, q Querier, runID int64, rec Record) (bool, error) {
	if rec.Source == "" || rec.ExternalID == "" {
		return false, errors.New("upsert source record: source and external id are required")
	}

	hash, err := CanonicalHash(rec.Payload)
	if err != nil {
		return false, fmt.Errorf("upsert %s/%s: %w", rec.Source, rec.ExternalID, err)
	}

	var changed, history, touched int
	if err := q.QueryRowContext(ctx, upsertSQL, rec.Source, rec.ExternalID, []byte(rec.Payload), hash, runID).
		Scan(&changed, &history, &touched); err != nil {
		return false, fmt.Errorf("upsert %s/%s: %w", rec.Source, rec.ExternalID, err)
	}

	return changed > 0, nil
}

// Store upserts records in transactions of ChunkSize and returns how many it saw and how many were new or
// changed. A failing chunk rolls back alone; earlier chunks stay committed.
func Store(ctx context.Context, db *database.DB, runID int64, records []Record) (seen, changed int, err error) {
	for start := 0; start < len(records); start += ChunkSize {
		chunk := records[start:min(start+ChunkSize, len(records))]
		n := 0

		err = db.Tx(ctx, func(tx *database.Tx) error {
			n = 0

			for _, rec := range chunk {
				ok, err := Upsert(ctx, tx, runID, rec)
				if err != nil {
					return err
				}

				if ok {
					n++
				}
			}

			return nil
		})
		if err != nil {
			return seen, changed, fmt.Errorf("store records: %w", err)
		}

		seen += len(chunk)
		changed += n
	}

	return seen, changed, nil
}
