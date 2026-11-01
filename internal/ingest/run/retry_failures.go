package run

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/flags"
)

// RetryBatch caps the QIDs one retry run takes per source (the admin trigger caps qids at 500 too).
const RetryBatch = 500

// RetryTrigger is the triggered_by of the runs RetryFailures queues.
const RetryTrigger = "schedule:retry"

// retryJob maps a dead-letter source to the job that can retry it. Editions failures wait for a review, and the
// editions backfill re-runs by hand, so neither retries here.
func retryJob(source string) (job, runSource string, ok bool) {
	switch source {
	case "normalize":
		return "normalize", "", true
	case "wikidata:entity":
		return "fetch.entities", source, true
	case "covers":
		return "covers", source, true
	case "editions", "backfill.editions":
		return "", "", false
	}

	return "fetch." + source, source, true
}

type retryGroup struct {
	source string
	ids    []string
}

// RetryFailures queues one run per source with non-parked dead-letter targets, restricted to those QIDs through
// the qids param (SDD 06 §7.6 "ingest.retry_failures"). Normalise failures go back to the pending index first.
// A source whose kill switch is off, or whose job is already active, is skipped. It returns how many runs it queued.
func RetryFailures(ctx context.Context, db *database.DB, d Dispatcher, ff *flags.Flags) (int, error) {
	groups, err := retryGroups(ctx, db)
	if err != nil {
		return 0, err
	}

	queued := 0

	for _, g := range groups {
		job, src, ok := retryJob(g.source)
		if !ok || !Allowed(ctx, ff, src) {
			continue
		}

		if job == "normalize" {
			if _, err := db.ExecContext(ctx, `UPDATE source_records SET normalized_at = NULL
WHERE source = 'wikidata' AND external_id = ANY($1::text[])`, g.ids); err != nil {
				return queued, fmt.Errorf("ingest retry: requeue normalise: %w", err)
			}
		}

		params, err := json.Marshal(map[string]any{"qids": g.ids})
		if err != nil {
			return queued, err
		}

		_, err = Create(
			ctx,
			db,
			d,
			Spec{Job: job, Source: src, Mode: ModeFull, Params: params, TriggeredBy: RetryTrigger},
		)
		if errors.Is(err, ErrRunActive) {
			continue
		}

		if err != nil {
			return queued, fmt.Errorf("ingest retry %s: %w", job, err)
		}

		queued++
	}

	return queued, nil
}

func retryGroups(ctx context.Context, db *database.DB) ([]retryGroup, error) {
	rows, err := db.QueryContext(ctx, `SELECT source, json_agg(external_id ORDER BY last_seen) FROM (
  SELECT source, external_id, last_seen, row_number() OVER (PARTITION BY source ORDER BY last_seen) AS rn
  FROM ingest_failures WHERE NOT parked
) f WHERE rn <= $1 GROUP BY source ORDER BY source`, RetryBatch)
	if err != nil {
		return nil, fmt.Errorf("ingest retry: list failures: %w", err)
	}

	defer func() { _ = rows.Close() }()

	var groups []retryGroup

	for rows.Next() {
		var (
			g   retryGroup
			raw []byte
		)

		if err := rows.Scan(&g.source, &raw); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(raw, &g.ids); err != nil {
			return nil, err
		}

		groups = append(groups, g)
	}

	return groups, rows.Err()
}
