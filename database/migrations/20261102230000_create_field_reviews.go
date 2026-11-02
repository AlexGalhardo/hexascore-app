package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261102230000CreateFieldReviews adds provenance v2 and the low-confidence review queue (RFC-001 section 5 D).
// items.provenance_v2 holds {field: {source, fetched_at, licence, confidence}} beside the legacy items.provenance
// ({field: source}, which DQ-11, the normaliser and the community edits still read as plain strings); readers use
// v2 only to drop fields whose confidence is under 60. field_reviews holds the values that did not reach items;
// resolved_by is the admin's users.id, kept as a bare id (the purge keeps the users row, SDD 03). Resolved rows are
// pruned after 365 days (run.Prune).
func m20261102230000CreateFieldReviews() schema.Migration {
	return postgresMigration("20261102230000_create_field_reviews",
		[]string{
			`ALTER TABLE items ADD COLUMN provenance_v2 jsonb NOT NULL DEFAULT '{}'::jsonb`,
			`CREATE TABLE field_reviews (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  item_id uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  field text NOT NULL,
  value jsonb NOT NULL,
  source text NOT NULL,
  confidence smallint NOT NULL CHECK (confidence BETWEEN 0 AND 100),
  reason text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now(),
  resolved_at timestamptz,
  resolved_by bigint,
  resolution text CHECK (resolution IN ('accepted', 'rejected'))
)`,
		},
		[]string{
			`DROP TABLE IF EXISTS field_reviews`,
			`ALTER TABLE items DROP COLUMN IF EXISTS provenance_v2`,
		},
	)
}

// m20261102230100AddFieldReviewsOpenIdx keeps one open review per item field and serves the queue page. One
// concurrent statement per NoTransaction migration.
func m20261102230100AddFieldReviewsOpenIdx() schema.Migration {
	m := postgresMigration("20261102230100_add_field_reviews_open_idx",
		[]string{
			`CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS field_reviews_open_uidx ON field_reviews (item_id, field) WHERE resolved_at IS NULL`,
		},
		[]string{`DROP INDEX CONCURRENTLY IF EXISTS field_reviews_open_uidx`},
	)
	m.NoTransaction = true

	return m
}
