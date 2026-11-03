// Package refresh is the per-item freshness check of the catalogue worker (RFC-001 section 5 A). It walks the public
// items, the least recently verified first, confirms that each cover is still in the bucket and stores a quality score.
// Re-fetching a source record past its TTL stays with the per-source chains, which re-read every record each week;
// selecting stale records item by item becomes necessary near 250 thousand items (milestone M1) and is not built here.
package refresh

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

// Job is the job_runs.job of the check; it touches no outside host, so only the master switch gates it.
const Job = "refresh.existing"

const (
	// Every is how long a verified item is left alone.
	Every = 14 * 24 * time.Hour
	// BatchSize is how many items share one transaction.
	BatchSize = 200
)

// Config wires the step. Stat reports whether a bucket object exists; nil (no bucket in this setup) counts every
// cover as present.
type Config struct {
	Stat  func(ctx context.Context, key string) (bool, error)
	Batch int
	Now   func() time.Time
}

type item struct{ id, qid, cover string }

// dueSQL: $1 verified before this time or never, $2 limit. Never verified first, then items with a cover, then the most
// popular (the index items_verify_idx serves exactly this order).
const dueSQL = `SELECT id::text, wikidata_qid, coalesce(cover_key, '') FROM items
WHERE visibility = 'public' AND (last_verified_at IS NULL OR last_verified_at < $1)
ORDER BY last_verified_at NULLS FIRST, (cover_key IS NULL), popularity_prior DESC, id LIMIT $2`

// scoreSQL stores the verification time and the quality score of the page. $1 item ids, $2 now, $3 ids whose cover is
// not in the bucket (they earn nothing for it).
const scoreSQL = `UPDATE items i SET last_verified_at = $2, quality_score = q.score
FROM (
  SELECT x.id, (
    25 * (x.cover_key IS NOT NULL AND NOT (x.id = ANY($3::uuid[])))::int
    + 15 * EXISTS (SELECT 1 FROM item_translations t WHERE t.item_id = x.id AND t.locale = 'en' AND t.synopsis IS NOT NULL)::int
    + 5 * EXISTS (SELECT 1 FROM item_translations t WHERE t.item_id = x.id AND t.locale = 'pt-BR')::int
    + 5 * EXISTS (SELECT 1 FROM item_translations t WHERE t.item_id = x.id AND t.locale = 'es')::int
    + 10 * (x.release_date IS NOT NULL)::int
    + 10 * ((SELECT count(*) FROM item_entities ie WHERE ie.item_id = x.id AND ie.role IN
        ('director', 'writer', 'cast', 'creator', 'author', 'developer', 'publisher', 'production_company')) >= 2)::int
    + 10 * (jsonb_array_length(x.trailers) >= 1)::int
    + 20 * coalesce((SELECT avg((v ->> 'confidence')::numeric) >= 80 FROM jsonb_each(x.provenance_v2) e(k, v)), false)::int
  ) AS score
  FROM items x WHERE x.id = ANY($1::uuid[])
) q
WHERE i.id = q.id`

// Step verifies one page of due items. Verified items leave the due set, so no cursor is needed.
func Step(c Config) run.Step {
	return func(ctx context.Context, tx *database.Tx, r run.Run) (run.Result, error) {
		if err := ctx.Err(); err != nil {
			return run.Result{}, err
		}

		batch := c.Batch
		if batch <= 0 {
			batch = BatchSize
		}

		now := c.Now().UTC()

		items, err := due(ctx, tx, now.Add(-Every), batch)
		if err != nil {
			return run.Result{}, err
		}

		res := run.Result{Done: len(items) < batch, Targets: len(items), Seen: len(items)}
		if len(items) == 0 {
			return res, nil
		}

		ids := make([]string, len(items))
		missing := []string{}

		for i, it := range items {
			if err := ctx.Err(); err != nil {
				return run.Result{}, err
			}

			ids[i] = it.id

			if it.cover == "" || c.Stat == nil {
				continue
			}

			ok, err := c.Stat(ctx, it.cover)
			if err != nil {
				return run.Result{}, fmt.Errorf("refresh: stat %s: %w", it.cover, err)
			}

			if ok {
				if err := run.ClearFailure(ctx, tx, Job, it.qid); err != nil {
					return run.Result{}, err
				}

				continue
			}

			// The cover columns belong to the cover job; here the gap is only reported, and the job re-fetches it.
			missing = append(missing, it.id)
			res.Failed++

			if _, err := run.RecordFailure(ctx, tx, r.ID, Job, it.qid, "cover_missing",
				fmt.Errorf("cover %s is not in the bucket", it.cover)); err != nil {
				return run.Result{}, err
			}
		}

		if _, err := tx.ExecContext(ctx, scoreSQL, array(ids), now, array(missing)); err != nil {
			return run.Result{}, fmt.Errorf("refresh: store scores: %w", err)
		}

		res.Changed = len(items)

		return res, nil
	}
}

func due(ctx context.Context, tx *database.Tx, before time.Time, limit int) ([]item, error) {
	rows, err := tx.QueryContext(ctx, dueSQL, before, limit)
	if err != nil {
		return nil, fmt.Errorf("refresh: list due: %w", err)
	}

	defer func() { _ = rows.Close() }()

	var out []item

	for rows.Next() {
		var it item
		if err := rows.Scan(&it.id, &it.qid, &it.cover); err != nil {
			return nil, err
		}

		out = append(out, it)
	}

	return out, rows.Err()
}

// array renders ids as a Postgres uuid[] literal.
func array(ids []string) string { return "{" + strings.Join(ids, ",") + "}" }
