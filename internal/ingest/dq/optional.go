package dq

import (
	"context"
	"fmt"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

// optional runs the asserts that need a collaborator: DQ-I-04 (Stat on cover objects) and DQ-I-12 (serialiser).
func (a Asserts) optional(ctx context.Context, tx *database.Tx, r run.Run) ([]result, error) {
	var out []result

	if a.Stat != nil {
		x, err := a.covers(ctx, tx, r)
		if err != nil {
			return nil, err
		}

		out = append(out, x)
	}

	if a.Serialise != nil {
		ids, err := a.Serialise(ctx, 100)
		if err != nil {
			return nil, fmt.Errorf("dq: DQ-I-12: %w", err)
		}

		out = append(out, result{
			code: "DQ-I-12", severity: "error", passed: len(ids) == 0, observed: float64(len(ids)),
			hasThreshold: true, sample: ids[:min(len(ids), SampleLimit)],
		})
	}

	return out, nil
}

// covers stats a random CoverSample of the stored covers (every cover in a full run).
func (a Asserts) covers(ctx context.Context, tx *database.Tx, r run.Run) (result, error) {
	limit := a.CoverSample
	if limit <= 0 {
		limit = 1000
	}

	if r.Mode == run.ModeFull {
		limit = -1
	}

	rows, err := tx.QueryContext(ctx, `SELECT id::text, key FROM (
  SELECT id, cover_key AS key FROM items WHERE cover_key IS NOT NULL
  UNION ALL SELECT id, cover_thumb_key FROM items WHERE cover_thumb_key IS NOT NULL
) c ORDER BY random() LIMIT CASE WHEN $1 < 0 THEN NULL ELSE $1 END`, limit)
	if err != nil {
		return result{}, fmt.Errorf("dq: list covers: %w", err)
	}

	type cover struct{ id, key string }

	var list []cover

	for rows.Next() {
		var c cover
		if err := rows.Scan(&c.id, &c.key); err != nil {
			_ = rows.Close()

			return result{}, err
		}

		list = append(list, c)
	}

	if err := rows.Close(); err != nil {
		return result{}, err
	}

	res := result{code: "DQ-I-04", severity: "error", hasThreshold: true, sample: []string{}}

	for _, c := range list {
		ok, err := a.Stat(ctx, c.key)
		if err != nil {
			return result{}, fmt.Errorf("dq: stat %s: %w", c.key, err)
		}

		if !ok {
			res.observed++

			if len(res.sample) < SampleLimit {
				res.sample = append(res.sample, c.id+":"+c.key)
			}
		}
	}

	res.passed = res.observed == 0

	return res, nil
}
