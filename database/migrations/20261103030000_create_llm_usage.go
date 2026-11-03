package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261103030000CreateLLMUsage adds the ledger of the LLM normaliser (RFC-001 section 5 F). One row per call: the
// tokens and the price in US dollars (internal/ingest/llm.Cost), summed per UTC month against LLM_MONTHLY_BUDGET_USD.
// outcome says what became of the answer: queued (a value went to field_reviews), empty (nothing found), invalid (it
// failed the schema; the tokens were still spent) or refused (the month's budget was spent; cost 0, one row per month,
// the marker that makes the budget alert fire once). item_id is kept NULL if the item is deleted, so the spend history
// survives. Written by internal/ingest/llm; rows older than 400 days are pruned by the ingest prune task.
func m20261103030000CreateLLMUsage() schema.Migration {
	return postgresMigration("20261103030000_create_llm_usage",
		[]string{
			`CREATE TABLE llm_usage (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  ts timestamptz NOT NULL DEFAULT now(),
  job text NOT NULL,
  model text NOT NULL,
  item_id uuid REFERENCES items(id) ON DELETE SET NULL,
  tokens_in int NOT NULL CHECK (tokens_in >= 0),
  tokens_out int NOT NULL CHECK (tokens_out >= 0),
  cost_usd numeric(10, 6) NOT NULL CHECK (cost_usd >= 0),
  outcome text NOT NULL CHECK (outcome IN ('queued', 'empty', 'invalid', 'refused'))
)`,
			// serves the month's spend and the once-a-month refusal marker
			`CREATE INDEX llm_usage_ts_idx ON llm_usage (ts)`,
			// serves "was this item asked about already"
			`CREATE INDEX llm_usage_item_idx ON llm_usage (item_id, job, ts) WHERE item_id IS NOT NULL`,
		},
		[]string{`DROP TABLE IF EXISTS llm_usage`},
	)
}
