package normalize

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/uuid"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/store"
)

// JobName is the job_runs.job of the normalise run.
const JobName = "normalize"

// BatchSize is how many QIDs share one transaction (NFR-06-3).
const BatchSize = 200

// Normalizer is the run.Step of the normalise job. The hooks keep this
// package free of the app layer: production wires ItemsChanged to the
// catalog.ItemsChanged job and ScoreChanged to scoring.RecomputeItemScore, both
// dispatched with DispatchTx so they exist only if the batch commits.
type Normalizer struct {
	// Batch overrides BatchSize (tests).
	Batch     int
	Community Community
	Now       func() time.Time
	// ItemsChanged is called once per batch with the ids of the saved items.
	ItemsChanged func(ctx context.Context, ex database.Executor, ids []string) error
	// ScoreChanged is called for a rated item whose release date, precision or status changed.
	ScoreChanged func(ctx context.Context, ex database.Executor, itemID string) error
}

// Step normalises up to BatchSize pending QIDs. A processed QID leaves the
// pending index (normalized_at is set even when the item is skipped or its
// payload is invalid), so the index itself is the cursor.
func (n Normalizer) Step(ctx context.Context, tx *database.Tx, r run.Run) (run.Result, error) {
	if err := SeedDataSources(ctx, tx); err != nil {
		return run.Result{}, err
	}

	batch := n.Batch
	if batch <= 0 {
		batch = BatchSize
	}

	qids, err := pending(ctx, tx, batch)
	if err != nil {
		return run.Result{}, err
	}

	res := run.Result{Cursor: nil, Done: len(qids) < batch, Targets: len(qids), Seen: len(qids)}
	if len(qids) == 0 {
		return res, nil
	}

	recs, err := loadRecords(ctx, tx, qids)
	if err != nil {
		return res, err
	}

	known, err := loadExisting(ctx, tx, qids)
	if err != nil {
		return res, err
	}

	now := n.Now()
	ids := map[string]string{}
	saved := []string{}
	items := map[string]*Item{}

	// Pass 1 writes items so that pass 2 can link items of the same batch.
	for _, qid := range qids {
		item, err := BuildItem(qid, recs[qid], now)

		switch {
		case errors.Is(err, ErrNoWikidata), errors.Is(err, ErrNoEnglishTitle):
			res.Skipped++

			continue
		case err != nil:
			res.Failed++

			if _, ferr := run.RecordFailure(ctx, tx, r.ID, JobName, qid, "invalid_payload", err); ferr != nil {
				return res, ferr
			}

			continue
		}

		old := known[qid]
		id, slug := "", ""

		if old != nil {
			id, slug = old.ID, old.Slug

			if _, err := ApplyCommunity(ctx, n.Community, id, item, now); err != nil {
				return res, err
			}
		} else {
			id = uuid.NewV7().String()

			if slug, err = chooseSlug(ctx, tx, item); err != nil {
				return res, err
			}
		}

		dateChanged, err := saveItem(ctx, tx, id, slug, old, item, now)
		if err != nil {
			return res, err
		}

		if dateChanged && n.ScoreChanged != nil {
			if err := n.dispatchScore(ctx, tx, id); err != nil {
				return res, err
			}
		}

		ids[qid], items[qid] = id, item
		saved = append(saved, id)
		res.Changed++
	}

	if err := n.saveRelations(ctx, tx, qids, recs, ids, items); err != nil {
		return res, err
	}

	if err := markDone(ctx, tx, qids, ids); err != nil {
		return res, err
	}

	if len(saved) > 0 && n.ItemsChanged != nil {
		if err := n.ItemsChanged(ctx, tx, saved); err != nil {
			return res, fmt.Errorf("items changed: %w", err)
		}
	}

	return res, nil
}

func (n Normalizer) dispatchScore(ctx context.Context, tx *database.Tx, id string) error {
	var rated bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM ratings WHERE item_id = $1)`, id).
		Scan(&rated); err != nil {
		return fmt.Errorf("check ratings of %s: %w", id, err)
	}

	if !rated {
		return nil
	}

	return n.ScoreChanged(ctx, tx, id)
}

func (n Normalizer) saveRelations(
	ctx context.Context,
	tx *database.Tx,
	qids []string,
	recs map[string][]store.Record,
	ids map[string]string,
	items map[string]*Item,
) error {
	// based_on may point at an item stored by an earlier batch.
	targets := map[string]bool{}

	for qid := range items {
		wd, _, err := decode(qid, recs[qid])
		if err != nil {
			return err
		}

		for _, t := range values(wd, "P144") {
			if _, inBatch := ids[t]; !inBatch {
				targets[t] = true
			}
		}
	}

	extra := make([]string, 0, len(targets))
	for t := range targets {
		extra = append(extra, t)
	}

	older, err := loadExisting(ctx, tx, extra)
	if err != nil {
		return err
	}

	all := map[string]string{}
	known := map[string]bool{}

	for q, id := range ids {
		all[q], known[q] = id, true
	}

	for q, e := range older {
		all[q], known[q] = e.ID, true
	}

	for _, qid := range qids {
		id, ok := ids[qid]
		if !ok {
			continue
		}

		rel, err := ExtractRelations(qid, recs[qid], known)
		if err != nil {
			return err
		}

		if err := saveRelations(ctx, tx, id, rel, all); err != nil {
			return err
		}
	}

	return nil
}

func pending(ctx context.Context, tx *database.Tx, limit int) ([]string, error) {
	rows, err := tx.QueryContext(ctx, `SELECT DISTINCT external_id FROM source_records
WHERE source IN (SELECT jsonb_array_elements_text($1::jsonb))
  AND (normalized_at IS NULL OR normalized_at < fetched_at)
ORDER BY external_id LIMIT $2`, jsonList(Sources), limit)
	if err != nil {
		return nil, fmt.Errorf("read pending qids: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []string

	for rows.Next() {
		var q string
		if err := rows.Scan(&q); err != nil {
			return nil, fmt.Errorf("read pending qids: %w", err)
		}

		out = append(out, q)
	}

	return out, rows.Err()
}

func loadRecords(ctx context.Context, tx *database.Tx, qids []string) (map[string][]store.Record, error) {
	rows, err := tx.QueryContext(ctx, `SELECT source, external_id, payload FROM source_records
WHERE source IN (SELECT jsonb_array_elements_text($1::jsonb))
  AND external_id IN (SELECT jsonb_array_elements_text($2::jsonb))`, jsonList(Sources), jsonList(qids))
	if err != nil {
		return nil, fmt.Errorf("read source records: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := map[string][]store.Record{}

	for rows.Next() {
		var r store.Record
		if err := rows.Scan(&r.Source, &r.ExternalID, &r.Payload); err != nil {
			return nil, fmt.Errorf("read source records: %w", err)
		}

		out[r.ExternalID] = append(out[r.ExternalID], r)
	}

	return out, rows.Err()
}

func markDone(ctx context.Context, tx *database.Tx, qids []string, ids map[string]string) error {
	for _, qid := range qids {
		var id any
		if v, ok := ids[qid]; ok {
			id = v
		}

		if _, err := tx.ExecContext(ctx, `UPDATE source_records SET normalized_at = now(), item_id = $2
WHERE source IN (SELECT jsonb_array_elements_text($1::jsonb)) AND external_id = $3`,
			jsonList(Sources), id, qid); err != nil {
			return fmt.Errorf("mark %s normalised: %w", qid, err)
		}
	}

	return nil
}
