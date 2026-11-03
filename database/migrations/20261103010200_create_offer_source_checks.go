package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261103010200CreateOfferSourceChecks remembers when an item's offers were last fetched from a source (RFC-003
// section 3). A film no provider carries has no offer row to hold that date, so without this table the TMDB stage
// would ask about it on every run. One row per item and source, upserted in place by the stage that fetched it, so it
// needs no prune (it follows the item).
func m20261103010200CreateOfferSourceChecks() schema.Migration {
	return postgresMigration("20261103010200_create_offer_source_checks",
		[]string{
			`CREATE TABLE offer_source_checks (
  item_id uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  source text NOT NULL CHECK (source IN ('tmdb_providers', 'watchmode')),
  checked_at timestamptz NOT NULL,
  PRIMARY KEY (item_id, source)
)`,
		},
		[]string{`DROP TABLE IF EXISTS offer_source_checks`},
	)
}
