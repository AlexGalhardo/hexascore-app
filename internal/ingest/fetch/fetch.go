// Package fetch turns an ingest.Connector into a run.Step: it pages through the connector's targets, stores the
// records and keeps the dead-letter table current (SDD 06 §7.3, §7.6).
package fetch

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

// BatchSize is how many targets one Step call fetches and stores in its transaction (NFR-06-3).
const BatchSize = 200

// IncrementalWindow is how far back an incremental run looks when its params carry no since (SDD 06 §7.6).
const IncrementalWindow = 36 * time.Hour

type runParams struct {
	Limit int       `json:"limit"`
	Since time.Time `json:"since"`
	Types []string  `json:"types"`
}

type cursor struct {
	After string `json:"after"`
}

// Step fetches up to BatchSize targets of c per call. The cursor is the last target id handed to Fetch, so a
// resumed run restarts after it; a parked target is skipped and a failed one is recorded for the dead letter.
func Step(c ingest.Connector, client *ingest.Client, now func() time.Time) run.Step {
	return func(ctx context.Context, tx *database.Tx, r run.Run) (run.Result, error) {
		var p runParams
		if len(r.Params) > 0 {
			if err := json.Unmarshal(r.Params, &p); err != nil {
				return run.Result{}, fmt.Errorf("fetch %s: params: %w: %w", r.Job, err, queue.ErrPermanent)
			}
		}

		var cur cursor
		if len(r.Cursor) > 0 && string(r.Cursor) != "null" {
			if err := json.Unmarshal(r.Cursor, &cur); err != nil {
				return run.Result{}, fmt.Errorf("fetch %s: cursor: %w: %w", r.Job, err, queue.ErrPermanent)
			}
		}

		if r.Mode == run.ModeIncremental && p.Since.IsZero() {
			p.Since = now().Add(-IncrementalWindow)
		}

		batch, exhausted, err := collect(ctx, tx, c, ingest.RunParams{
			RunID: r.ID, Mode: r.Mode, Since: p.Since, Types: p.Types, Limit: p.Limit,
		}, cur.After)
		if err != nil {
			return run.Result{}, err
		}

		res := run.Result{Done: exhausted, Targets: len(batch.targets)}
		if batch.last != "" {
			cur.After = batch.last
		}

		res.Skipped = batch.parked

		raw, err := json.Marshal(cur)
		if err != nil {
			return run.Result{}, err
		}

		res.Cursor = raw

		if len(batch.targets) == 0 {
			return res, nil
		}

		records, failures := c.Fetch(ctx, client, batch.targets)
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

		return res, settle(ctx, tx, r, c.Source(), batch.targets, failures, &res)
	}
}

type page struct {
	// last is the final target read, parked or not: the cursor moves past it.
	last    string
	targets []ingest.Target
	parked  int
}

// collect reads one page of BatchSize targets after the cursor. exhausted is true when the connector ran out
// before the page was full. Parked targets are filtered after the read: the connector may still hold a cursor
// open on tx, so no other query can run while it iterates.
func collect(
	ctx context.Context, tx *database.Tx, c ingest.Connector, p ingest.RunParams, after string,
) (page, bool, error) {
	var (
		raw       []ingest.Target
		exhausted = true
	)

	for t, err := range c.Targets(ctx, tx, p, after) {
		if err != nil {
			return page{}, false, err
		}

		raw = append(raw, t)

		if len(raw) == BatchSize {
			exhausted = false

			break
		}
	}

	var out page

	for _, t := range raw {
		out.last = t.ID

		parked, err := run.IsParked(ctx, tx, c.Source(), t.ID)
		if err != nil {
			return page{}, false, err
		}

		if parked {
			out.parked++

			continue
		}

		out.targets = append(out.targets, t)
	}

	return out, exhausted, nil
}

// settle records the failed targets and forgets the ones that went through. A rate-limit or permanent error of
// the source ends the run instead of parking every target behind it.
func settle(
	ctx context.Context,
	tx *database.Tx,
	r run.Run,
	source string,
	targets []ingest.Target,
	failures []ingest.TargetError,
	res *run.Result,
) error {
	failed := make(map[string]bool, len(failures))

	for _, f := range failures {
		if errors.Is(f.Err, queue.ErrPermanent) {
			return f.Err
		}

		failed[f.Target.ID] = true

		if _, err := run.RecordFailure(ctx, tx, r.ID, source, f.Target.ID, "fetch", f.Err); err != nil {
			return err
		}

		res.Failed++
	}

	for _, t := range targets {
		if !failed[t.ID] {
			if err := run.ClearFailure(ctx, tx, source, t.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

// Unavailable is the Step of a source whose connector does not exist yet. It fails permanently, so a switch
// turned on by mistake is loud; with the switch off the Runner answers skipped before reaching it.
func Unavailable(source, why string) run.Step {
	return func(_ context.Context, _ *database.Tx, r run.Run) (run.Result, error) {
		return run.Result{}, fmt.Errorf(
			"fetch %s: no connector for %s (%s): %w",
			r.Job,
			source,
			why,
			queue.ErrPermanent,
		)
	}
}
