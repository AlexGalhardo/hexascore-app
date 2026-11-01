// Package expire enforces the cache terms of a display-only source (SDD 06 §7.6): records past RefreshAfter are
// fetched again, records past DeleteAfter are deleted with the source's cover, and their items are re-normalised.
// tmdb.expire and igdb.refresh are its two policies.
package expire

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/store"
)

// Batch is how many records one Step call deletes and how many it refreshes.
const Batch = 200

// Policy is the cache term of one source.
type Policy struct {
	Source       string
	RefreshAfter time.Duration
	DeleteAfter  time.Duration
	// Target rebuilds the fetch target of a stored record from its QID, item type and payload.
	Target func(qid, itemType string, payload []byte) ingest.Target
}

type cursor struct {
	After string `json:"after"`
}

// Step deletes the records past DeleteAfter, drops the source's cover of their items and marks the items for the
// normaliser, then refreshes the records past RefreshAfter through c. The deletion set shrinks as it is processed;
// the refresh set is walked with a cursor because a record the source no longer knows keeps its old age.
func Step(p Policy, c ingest.Connector, client *ingest.Client, now func() time.Time) run.Step {
	return func(ctx context.Context, tx *database.Tx, r run.Run) (run.Result, error) {
		var cur cursor
		if len(r.Cursor) > 0 && string(r.Cursor) != "null" {
			if err := json.Unmarshal(r.Cursor, &cur); err != nil {
				return run.Result{}, fmt.Errorf("%s expire: cursor: %w: %w", p.Source, err, queue.ErrPermanent)
			}
		}

		t := now()
		deleteBefore := t.Add(-p.DeleteAfter)

		deleted, err := deleteExpired(ctx, tx, p.Source, deleteBefore)
		if err != nil {
			return run.Result{}, err
		}

		targets, err := refreshTargets(ctx, tx, p, t.Add(-p.RefreshAfter), deleteBefore, cur.After)
		if err != nil {
			return run.Result{}, err
		}

		res := run.Result{Targets: len(deleted) + len(targets), Changed: len(deleted)}
		res.Done = len(deleted) < Batch && len(targets) < Batch

		if len(targets) > 0 {
			cur.After = targets[len(targets)-1].ID
		}

		if res.Cursor, err = json.Marshal(cur); err != nil {
			return run.Result{}, err
		}

		if len(targets) == 0 {
			return res, nil
		}

		records, failures := c.Fetch(ctx, client, targets)
		for _, f := range failures {
			if errors.Is(f.Err, queue.ErrPermanent) {
				return run.Result{}, f.Err
			}

			res.Failed++
		}

		for _, rec := range records {
			changed, err := store.Upsert(ctx, tx, r.ID, rec)
			if err != nil {
				return run.Result{}, err
			}

			res.Seen++

			if changed {
				res.Changed++
			}
		}

		return res, nil
	}
}

// deleteExpired removes up to Batch records of source older than before, drops the source's cover columns of their
// items (the objects are collected by covers.gc after the deferred-delete window) and clears normalized_at on the
// Wikidata records so the normaliser rebuilds the items without that source's fields.
func deleteExpired(ctx context.Context, tx *database.Tx, source string, before time.Time) ([]string, error) {
	rows, err := tx.QueryContext(ctx, `
DELETE FROM source_records WHERE source = $3 AND external_id IN (
  SELECT external_id FROM source_records WHERE source = $3 AND checked_at < $1 ORDER BY external_id LIMIT $2)
RETURNING external_id`, before, Batch, source)
	if err != nil {
		return nil, fmt.Errorf("%s expire: delete: %w", source, err)
	}

	defer func() { _ = rows.Close() }()

	var qids []string

	for rows.Next() {
		var q string
		if err := rows.Scan(&q); err != nil {
			return nil, fmt.Errorf("%s expire: scan: %w", source, err)
		}

		qids = append(qids, q)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	_ = rows.Close()

	if len(qids) == 0 {
		return nil, nil
	}

	list, err := json.Marshal(qids)
	if err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, `
UPDATE items SET cover_key = NULL, cover_thumb_key = NULL, cover_source = NULL, cover_attribution = NULL
WHERE cover_source = $2 AND wikidata_qid IN (SELECT jsonb_array_elements_text($1::jsonb))`, list, source); err != nil {
		return nil, fmt.Errorf("%s expire: drop covers: %w", source, err)
	}

	if _, err := tx.ExecContext(ctx, `
UPDATE source_records SET normalized_at = NULL
WHERE source = 'wikidata' AND external_id IN (SELECT jsonb_array_elements_text($1::jsonb))`, list); err != nil {
		return nil, fmt.Errorf("%s expire: requeue normalise: %w", source, err)
	}

	return qids, nil
}

// refreshTargets lists up to Batch stored records with refreshBefore > checked_at >= deleteBefore after the cursor.
func refreshTargets(
	ctx context.Context, tx *database.Tx, p Policy, refreshBefore, deleteBefore time.Time, after string,
) ([]ingest.Target, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT s.external_id, i.type, s.payload
FROM source_records s JOIN items i ON i.wikidata_qid = s.external_id
WHERE s.source = $5 AND s.checked_at < $1 AND s.checked_at >= $2 AND s.external_id > $3
ORDER BY s.external_id LIMIT $4`, refreshBefore, deleteBefore, after, Batch, p.Source)
	if err != nil {
		return nil, fmt.Errorf("%s expire: targets: %w", p.Source, err)
	}

	defer func() { _ = rows.Close() }()

	var out []ingest.Target

	for rows.Next() {
		var (
			qid, itemType string
			payload       []byte
		)
		if err := rows.Scan(&qid, &itemType, &payload); err != nil {
			return nil, fmt.Errorf("%s expire: scan target: %w", p.Source, err)
		}

		out = append(out, p.Target(qid, itemType, payload))
	}

	return out, rows.Err()
}
