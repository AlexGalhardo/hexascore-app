package scoring

import "github.com/AlexGalhardo/hexascore-app/app/catalog"

// DQAsserts returns the scoring data-quality asserts (SDD 02 §10), run by `catalog:check` after DQ-01..DQ-14.
// SC-02 is the drift count: it must be 0 right after the nightly job.
func DQAsserts() []catalog.DQAssert {
	return []catalog.DQAssert{
		{
			ID: "SC-01", Level: catalog.LevelError, Description: "score outside [5,100]",
			Query: `SELECT id::text FROM items WHERE score IS NOT NULL AND (score < 5 OR score > 100)`,
		},
		{
			ID:          "SC-02",
			Level:       catalog.LevelError,
			Description: "rating_count differs from a recount of counted ratings",
			Query:       driftQuery,
		},
		{
			ID: "SC-03", Level: catalog.LevelWarn, Description: "rated item not recomputed for more than 36 hours",
			Query: `SELECT id::text FROM items WHERE rating_count > 0
AND (score_computed_at IS NULL OR score_computed_at < now() - interval '36 hours')`,
		},
	}
}

// driftQuery reuses the recompute weights evaluated at now(), so it cannot disagree with the job.
var driftQuery = replaceAt(weightedRatingsSQL) + `
), n AS (
  SELECT item_id, count(*) AS n FROM w WHERE weight > 0 GROUP BY item_id
)
SELECT i.id::text FROM items i LEFT JOIN n ON n.item_id = i.id WHERE i.rating_count <> coalesce(n.n, 0)`
