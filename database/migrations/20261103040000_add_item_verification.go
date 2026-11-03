package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261103040000AddItemVerification adds the two columns of the refresh job (RFC-001 section 5 A and section 7).
// last_verified_at is when refresh.existing last checked the item (cover in the bucket, quality score); NULL means never.
// quality_score (0-100) adds up cover 25, English synopsis 15, pt-BR and es titles 5 each, release date 10, two
// credits 10, a trailer 10 and a provenance confidence mean of 80 or more 20 (internal/ingest/refresh). Written by
// refresh.existing only; nothing reads quality_score for visibility yet (the trending floor of RFC-002 uses its own
// provisional expression until this column is filled everywhere).
func m20261103040000AddItemVerification() schema.Migration {
	return postgresMigration("20261103040000_add_item_verification",
		[]string{
			`ALTER TABLE items ADD COLUMN last_verified_at timestamptz,
  ADD COLUMN quality_score smallint CHECK (quality_score BETWEEN 0 AND 100)`,
		},
		[]string{`ALTER TABLE items DROP COLUMN IF EXISTS quality_score, DROP COLUMN IF EXISTS last_verified_at`},
	)
}

// m20261103040100AddItemsVerifyIdx serves refresh.existing: the next items to check are the never verified first, then
// the ones with a cover, then the most popular. One concurrent statement per NoTransaction migration.
func m20261103040100AddItemsVerifyIdx() schema.Migration {
	m := postgresMigration("20261103040100_add_items_verify_idx",
		[]string{
			`CREATE INDEX CONCURRENTLY IF NOT EXISTS items_verify_idx ON items
  (last_verified_at NULLS FIRST, (cover_key IS NULL), popularity_prior DESC, id) WHERE visibility = 'public'`,
		},
		[]string{`DROP INDEX CONCURRENTLY IF EXISTS items_verify_idx`},
	)
	m.NoTransaction = true

	return m
}
