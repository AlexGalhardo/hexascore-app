package tmdb

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

// JobName is the job_runs.job of the expiry run.
const JobName = "tmdb.expire"

// RefreshAfter and DeleteAfter are the TMDB cache terms (SDD 06 §7.6, FR-06-5): no value may be older than 180
// days by checked_at, so refresh starts 30 days earlier.
const (
	RefreshAfter = 150 * 24 * time.Hour
	DeleteAfter  = 180 * 24 * time.Hour
)

// ExpireBatch is how many records one Step call deletes and how many it refreshes.
const ExpireBatch = 200

type expireCursor struct {
	After string `json:"after"`
}

// ExpireStep deletes the TMDB records past DeleteAfter, drops the TMDB cover of their items and marks the items
// for the normaliser, then refreshes the records past RefreshAfter through c. The deletion set shrinks as it is
// processed; the refresh set is walked with a cursor because a record TMDB no longer knows keeps its old age.
func ExpireStep(c *Connector, client *ingest.Client, now func() time.Time) run.Step {
	return func(ctx context.Context, tx *database.Tx, r run.Run) (run.Result, error) {
		var cur expireCursor
		if len(r.Cursor) > 0 && string(r.Cursor) != "null" {
			if err := json.Unmarshal(r.Cursor, &cur); err != nil {
				return run.Result{}, fmt.Errorf("tmdb expire: cursor: %w: %w", err, queue.ErrPermanent)
			}
		}

		t := now()
		deleteBefore := t.Add(-DeleteAfter)

		deleted, err := deleteExpired(ctx, tx, deleteBefore)
		if err != nil {
			return run.Result{}, err
		}

		targets, err := refreshTargets(ctx, tx, t.Add(-RefreshAfter), deleteBefore, cur.After)
		if err != nil {
			return run.Result{}, err
		}

		res := run.Result{Targets: len(deleted) + len(targets), Changed: len(deleted)}
		res.Done = len(deleted) < ExpireBatch && len(targets) < ExpireBatch

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

// deleteExpired removes up to ExpireBatch records older than before, drops the TMDB cover columns of their items
// (the objects are collected by covers.gc after the deferred-delete window) and clears normalized_at on the
// Wikidata records so the normaliser rebuilds the items without TMDB fields.
func deleteExpired(ctx context.Context, tx *database.Tx, before time.Time) ([]string, error) {
	rows, err := tx.QueryContext(ctx, `
DELETE FROM source_records WHERE source = 'tmdb' AND external_id IN (
  SELECT external_id FROM source_records WHERE source = 'tmdb' AND checked_at < $1 ORDER BY external_id LIMIT $2)
RETURNING external_id`, before, ExpireBatch)
	if err != nil {
		return nil, fmt.Errorf("tmdb expire: delete: %w", err)
	}

	defer func() { _ = rows.Close() }()

	var qids []string

	for rows.Next() {
		var q string
		if err := rows.Scan(&q); err != nil {
			return nil, fmt.Errorf("tmdb expire: scan: %w", err)
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
WHERE cover_source = 'tmdb' AND wikidata_qid IN (SELECT jsonb_array_elements_text($1::jsonb))`, list); err != nil {
		return nil, fmt.Errorf("tmdb expire: drop covers: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
UPDATE source_records SET normalized_at = NULL
WHERE source = 'wikidata' AND external_id IN (SELECT jsonb_array_elements_text($1::jsonb))`, list); err != nil {
		return nil, fmt.Errorf("tmdb expire: requeue normalise: %w", err)
	}

	return qids, nil
}

// refreshTargets lists up to ExpireBatch stored records with refreshBefore > checked_at >= deleteBefore after the
// cursor. The item supplies the type; the stored payload supplies the TMDB id.
func refreshTargets(
	ctx context.Context, tx *database.Tx, refreshBefore, deleteBefore time.Time, after string,
) ([]ingest.Target, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT s.external_id, i.type, coalesce(s.payload->>'id', '')
FROM source_records s JOIN items i ON i.wikidata_qid = s.external_id
WHERE s.source = 'tmdb' AND s.checked_at < $1 AND s.checked_at >= $2 AND s.external_id > $3
ORDER BY s.external_id LIMIT $4`, refreshBefore, deleteBefore, after, ExpireBatch)
	if err != nil {
		return nil, fmt.Errorf("tmdb expire: targets: %w", err)
	}

	defer func() { _ = rows.Close() }()

	var out []ingest.Target

	for rows.Next() {
		var qid, itemType, id string
		if err := rows.Scan(&qid, &itemType, &id); err != nil {
			return nil, fmt.Errorf("tmdb expire: scan target: %w", err)
		}

		out = append(out, ingest.Target{ID: qid, Type: Ref(itemType, id, "")})
	}

	return out, rows.Err()
}
