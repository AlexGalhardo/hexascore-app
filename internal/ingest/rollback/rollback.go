// Package rollback undoes a bad ingestion run while its history is still kept (SDD 06 §7.8 "Bad run"). Beyond the
// history window the path is the dump restore runbook (SDD 11), not code.
package rollback

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
	"github.com/AlexGalhardo/hexascore-app/internal/search"
)

// Job is the job_runs.job of a rollback run.
const Job = "rollback"

// HistoryDays is how long source_record_history is kept (SDD 06 §5), so how old a run may be and still roll back.
const HistoryDays = 30

// AuditField is the catalog_audit.field of the entries a rollback writes for the items it removed or hid.
const AuditField = "*ingest-rollback*"

// Errors the admin route maps to statuses.
var (
	ErrRunNotFound       = errors.New("rollback: no such run")
	ErrNotRollbackable   = errors.New("rollback: only a finished fetch run can be rolled back")
	ErrHistoryExpired    = errors.New("rollback: the history of this run has expired")
	ErrAlreadyRolledBack = errors.New("rollback: this run was already rolled back")
)

// Params is job_runs.params of a rollback run.
type Params struct {
	TargetRun int64  `json:"target_run"`
	Actor     int64  `json:"actor"`
	Reason    string `json:"reason"`
}

// Check reports why run id cannot be rolled back, or nil.
func Check(ctx context.Context, ex database.Executor, id int64) error {
	var (
		job, status string
		started     bool
		expired     bool
	)

	err := ex.QueryRowContext(ctx, `SELECT job, status, started_at IS NOT NULL,
  created_at < now() - make_interval(days => $2) FROM job_runs WHERE id = $1`, id, HistoryDays).
		Scan(&job, &status, &started, &expired)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrRunNotFound
	}

	if err != nil {
		return fmt.Errorf("rollback: load run %d: %w", id, err)
	}

	if !strings.HasPrefix(job, "fetch.") && job != "tmdb.expire" {
		return fmt.Errorf("%w: %s", ErrNotRollbackable, job)
	}

	if !started || status == run.StatusQueued || status == run.StatusRunning {
		return fmt.Errorf("%w: run %d is %s", ErrNotRollbackable, id, status)
	}

	if expired {
		return ErrHistoryExpired
	}

	var done int

	err = ex.QueryRowContext(ctx, `SELECT count(*) FROM job_runs WHERE job = $1 AND params->>'target_run' = $2::text
AND status IN ('queued', 'running', 'succeeded', 'partial')`, Job, strconv.FormatInt(id, 10)).Scan(&done)
	if err != nil {
		return fmt.Errorf("rollback: earlier rollbacks of %d: %w", id, err)
	}

	if done > 0 {
		return ErrAlreadyRolledBack
	}

	return nil
}

// userData is true for an item that a user rated, tracked, listed or attached an edition to.
const userData = `EXISTS (SELECT 1 FROM ratings WHERE item_id = i.id) OR EXISTS (SELECT 1 FROM item_status WHERE item_id = i.id)
OR EXISTS (SELECT 1 FROM time_to_beat WHERE item_id = i.id) OR EXISTS (SELECT 1 FROM list_items WHERE item_id = i.id)
OR EXISTS (SELECT 1 FROM user_item_editions WHERE item_id = i.id)`

// laterChange is true when a run after $1 superseded the same record again, so the record no longer belongs to
// the target run.
const laterChange = `EXISTS (SELECT 1 FROM source_record_history l WHERE l.source = h.source
AND l.external_id = h.external_id AND l.superseded_by_run > $1)`

// Step rolls the target run back in one transaction. Records the run changed go back to their previous version
// and become pending for the normaliser; records it inserted are deleted, with the items it created unless user
// data references them (those are hidden instead). A record a later run changed again is left alone and counted
// as skipped. Items older than the run are never deleted.
func Step(ctx context.Context, tx *database.Tx, r run.Run) (run.Result, error) {
	var p Params
	if err := json.Unmarshal(r.Params, &p); err != nil || p.TargetRun < 1 {
		return run.Result{}, fmt.Errorf("rollback run %d: bad params: %w", r.ID, queue.ErrPermanent)
	}

	var started sql.NullTime

	err := tx.QueryRowContext(ctx, `SELECT started_at FROM job_runs WHERE id = $1`, p.TargetRun).Scan(&started)
	if err != nil || !started.Valid {
		return run.Result{}, fmt.Errorf(
			"rollback run %d: target %d is not a started run: %w",
			r.ID,
			p.TargetRun,
			queue.ErrPermanent,
		)
	}

	var res run.Result

	err = tx.QueryRowContext(ctx, `SELECT count(DISTINCT (h.source, h.external_id)) FROM source_record_history h
WHERE h.superseded_by_run = $1 AND h.fetched_at < $2 AND `+laterChange, p.TargetRun, started.Time).Scan(&res.Skipped)
	if err != nil {
		return res, fmt.Errorf("rollback: count skipped: %w", err)
	}

	var restored int

	err = tx.QueryRowContext(ctx, `WITH pre AS (
  SELECT DISTINCT ON (h.source, h.external_id) h.source, h.external_id, h.hash, h.payload, h.fetched_at
  FROM source_record_history h
  WHERE h.superseded_by_run = $1 AND h.fetched_at < $2 AND NOT `+laterChange+`
  ORDER BY h.source, h.external_id, h.superseded_at
), upd AS (
  UPDATE source_records s SET payload = p.payload, hash = p.hash, fetched_at = p.fetched_at, normalized_at = NULL,
    last_run_id = $3
  FROM pre p WHERE s.source = p.source AND s.external_id = p.external_id RETURNING s.source, s.external_id
), del AS (
  DELETE FROM source_record_history h USING upd
  WHERE h.source = upd.source AND h.external_id = upd.external_id AND h.superseded_by_run = $1 RETURNING 1
)
SELECT count(*) FROM upd`, p.TargetRun, started.Time, r.ID).Scan(&restored)
	if err != nil {
		return res, fmt.Errorf("rollback: restore records: %w", err)
	}

	qids, removed, err := deleteInserted(ctx, tx, p.TargetRun, started.Time)
	if err != nil {
		return res, err
	}

	if err := removeItems(ctx, tx, p, qids, started.Time); err != nil {
		return res, err
	}

	res.Done = true
	res.Targets = restored + removed
	res.Seen = res.Targets
	res.Changed = res.Targets

	return res, nil
}

// deleteInserted removes the records the run inserted (written by it, never superseded from a pre-run version)
// and returns their Wikidata QIDs and how many went.
func deleteInserted(ctx context.Context, tx *database.Tx, target int64, started time.Time) ([]string, int, error) {
	rows, err := tx.QueryContext(ctx, `WITH del AS (
  DELETE FROM source_records s WHERE s.last_run_id = $1 AND s.fetched_at >= $2
    AND NOT EXISTS (SELECT 1 FROM source_record_history h WHERE h.source = s.source AND h.external_id = s.external_id
      AND h.superseded_by_run = $1 AND h.fetched_at < $2)
  RETURNING s.source, s.external_id
), hist AS (
  DELETE FROM source_record_history h USING del
  WHERE h.source = del.source AND h.external_id = del.external_id AND h.superseded_by_run = $1 RETURNING 1
)
SELECT source, external_id FROM del`, target, started)
	if err != nil {
		return nil, 0, fmt.Errorf("rollback: delete inserted records: %w", err)
	}

	defer func() { _ = rows.Close() }()

	var (
		qids []string
		n    int
	)

	for rows.Next() {
		var source, id string
		if err := rows.Scan(&source, &id); err != nil {
			return nil, 0, err
		}

		n++

		if source == "wikidata" {
			qids = append(qids, id)
		}
	}

	return qids, n, rows.Err()
}

type removable struct {
	id, qid, kind, slug string
	used                bool
}

func removeItems(ctx context.Context, tx *database.Tx, p Params, qids []string, started time.Time) error {
	if len(qids) == 0 {
		return nil
	}

	items, err := createdItems(ctx, tx, qids, started)
	if err != nil {
		return err
	}

	for _, it := range items {
		action, query := "deleted", `DELETE FROM items WHERE id = $1`
		if it.used {
			action, query = "hidden", `UPDATE items SET visibility = 'hidden', updated_at = now() WHERE id = $1`
		}

		before, err := json.Marshal(map[string]any{
			"action": action, "run": p.TargetRun, "qid": it.qid, "type": it.kind, "slug": it.slug,
		})
		if err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, `INSERT INTO catalog_audit
(actor_id, actor_role, source, target_type, target_id, field, before, reason)
VALUES (NULLIF($1, 0), 'admin', 'admin', 'item', $2, $3, $4::jsonb, $5)`,
			p.Actor, it.id, AuditField, string(before), p.Reason); err != nil {
			return fmt.Errorf("rollback: audit %s: %w", it.id, err)
		}

		if _, err := tx.ExecContext(ctx, query, it.id); err != nil {
			return fmt.Errorf("rollback: %s item %s: %w", action, it.id, err)
		}

		if err := search.MarkDirty(ctx, tx, "item", it.id); err != nil {
			return err
		}
	}

	return nil
}

func createdItems(ctx context.Context, tx *database.Tx, qids []string, started time.Time) ([]removable, error) {
	rows, err := tx.QueryContext(ctx, `SELECT i.id::text, i.wikidata_qid, i.type, i.slug, `+userData+`
FROM items i WHERE i.wikidata_qid = ANY($1::text[]) AND i.created_at >= $2`, qids, started)
	if err != nil {
		return nil, fmt.Errorf("rollback: find created items: %w", err)
	}

	defer func() { _ = rows.Close() }()

	var items []removable

	for rows.Next() {
		var it removable
		if err := rows.Scan(&it.id, &it.qid, &it.kind, &it.slug, &it.used); err != nil {
			return nil, err
		}

		items = append(items, it)
	}

	return items, rows.Err()
}
