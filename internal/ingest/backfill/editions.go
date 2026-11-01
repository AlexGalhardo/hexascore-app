// Package backfill holds the one-off jobs that fill a new field over the existing catalogue (SDD 06 §7.8).
package backfill

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/editions"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/normalize"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/store"
)

// EditionsJob is the job_runs.job of the edition backfill.
const EditionsJob = "backfill.editions"

// BatchSize is how many items share one transaction (SDD 06 §7.8).
const BatchSize = 500

// Editions walks the books by items.id and saves the editions their stored Wikidata record lists. Saving is an
// upsert, so a rerun or an item the normaliser already handled changes nothing.
type Editions struct {
	// Batch overrides BatchSize (tests).
	Batch int
}

type cursor struct {
	After string `json:"after"`
}

// Step processes one keyset page. The cursor holds the last id seen.
func (e Editions) Step(ctx context.Context, tx *database.Tx, r run.Run) (run.Result, error) {
	var cur cursor
	if len(r.Cursor) > 0 && string(r.Cursor) != "null" {
		if err := json.Unmarshal(r.Cursor, &cur); err != nil {
			return run.Result{}, fmt.Errorf("backfill editions: cursor: %w: %w", err, queue.ErrPermanent)
		}
	}

	batch := e.Batch
	if batch <= 0 {
		batch = BatchSize
	}

	books, err := listBooks(ctx, tx, cur.After, batch)
	if err != nil {
		return run.Result{}, err
	}

	res := run.Result{Done: len(books) < batch, Targets: len(books), Seen: len(books)}

	for _, b := range books {
		cur.After = b.id

		if err := e.one(ctx, tx, r, b.id, b.qid, &res); err != nil {
			return run.Result{}, err
		}
	}

	res.Cursor, err = json.Marshal(cur)

	return res, err
}

func (Editions) one(ctx context.Context, tx *database.Tx, r run.Run, id, qid string, res *run.Result) error {
	var payload json.RawMessage

	err := tx.QueryRowContext(ctx, `SELECT payload FROM source_records WHERE source = 'wikidata' AND external_id = $1`, qid).
		Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		res.Skipped++

		return nil
	}

	if err != nil {
		return fmt.Errorf("backfill editions: read record of %s: %w", qid, err)
	}

	cands, _, err := normalize.ExtractEditions(
		qid,
		[]store.Record{{Source: "wikidata", ExternalID: qid, Payload: payload}},
	)
	if err != nil {
		res.Failed++

		_, ferr := run.RecordFailure(ctx, tx, r.ID, EditionsJob, qid, "invalid_payload", err)

		return ferr
	}

	conflicts, err := editions.SaveAll(ctx, tx, id, cands)
	if err != nil {
		return err
	}

	for _, c := range conflicts {
		res.Failed++

		if _, err := run.RecordFailure(ctx, tx, r.ID, editions.FailureSource, c.ID, c.Class,
			fmt.Errorf("edition %s belongs to another work", c.ID)); err != nil {
			return err
		}
	}

	res.Changed += len(cands) - len(conflicts)

	return nil
}

type book struct{ id, qid string }

func listBooks(ctx context.Context, tx *database.Tx, after string, limit int) ([]book, error) {
	rows, err := tx.QueryContext(ctx, `SELECT id::text, wikidata_qid FROM items
WHERE type = 'book' AND wikidata_qid IS NOT NULL AND ($1 = '' OR id > $1::uuid) ORDER BY id LIMIT $2`, after, limit)
	if err != nil {
		return nil, fmt.Errorf("backfill editions: list books: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []book

	for rows.Next() {
		var b book
		if err := rows.Scan(&b.id, &b.qid); err != nil {
			return nil, fmt.Errorf("backfill editions: list books: %w", err)
		}

		out = append(out, b)
	}

	return out, rows.Err()
}
