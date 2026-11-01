package dq

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

var itemTypes = []string{"game", "movie", "series", "book"}

// countDrop is DQ-I-01: items per type against the last catalog_counts snapshot. A drop over 1 % fails unless the
// run carries allow_count_drop and a reason. The snapshot only moves when the assert passes, so a failing drop keeps
// failing until it is allowed or recovered.
func (Asserts) countDrop(ctx context.Context, tx *database.Tx, r run.Run) (result, error) {
	now := map[string]int{}
	for _, t := range itemTypes {
		now[t] = 0
	}

	rows, err := tx.QueryContext(ctx, `SELECT type, count(*) FROM items GROUP BY type`)
	if err != nil {
		return result{}, fmt.Errorf("dq: count items: %w", err)
	}

	for rows.Next() {
		var (
			t string
			n int
		)

		if err := rows.Scan(&t, &n); err != nil {
			_ = rows.Close()

			return result{}, err
		}

		now[t] = n
	}

	if err := rows.Close(); err != nil {
		return result{}, err
	}

	prev := map[string]int{}

	rows, err = tx.QueryContext(
		ctx,
		`SELECT DISTINCT ON (type) type, items FROM catalog_counts ORDER BY type, day DESC`,
	)
	if err != nil {
		return result{}, fmt.Errorf("dq: read snapshot: %w", err)
	}

	for rows.Next() {
		var (
			t string
			n int
		)

		if err := rows.Scan(&t, &n); err != nil {
			_ = rows.Close()

			return result{}, err
		}

		prev[t] = n
	}

	if err := rows.Close(); err != nil {
		return result{}, err
	}

	res := result{
		code:         "DQ-I-01",
		severity:     "error",
		threshold:    maxCountDrop * 100,
		hasThreshold: true,
		sample:       []string{},
	}

	for _, t := range itemTypes {
		if p := prev[t]; p > 0 && now[t] < p {
			drop := float64(p-now[t]) / float64(p) * 100
			res.observed = math.Max(res.observed, drop)

			if drop > res.threshold {
				res.sample = append(res.sample, t)
			}
		}
	}

	res.passed = len(res.sample) == 0 || allowsDrop(r.Params)

	if res.passed {
		for _, t := range itemTypes {
			if _, err := tx.ExecContext(
				ctx,
				`INSERT INTO catalog_counts (day, type, items) VALUES (current_date, $1, $2)
ON CONFLICT (day, type) DO UPDATE SET items = excluded.items`,
				t,
				now[t],
			); err != nil {
				return result{}, fmt.Errorf("dq: write snapshot: %w", err)
			}
		}
	}

	return res, nil
}

// allowsDrop reads the allow_count_drop override; it needs a non-empty reason (SDD 06 §7.7).
func allowsDrop(params json.RawMessage) bool {
	var p struct {
		Allow  bool   `json:"allow_count_drop"`
		Reason string `json:"reason"`
	}

	return json.Unmarshal(params, &p) == nil && p.Allow && p.Reason != ""
}
