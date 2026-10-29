package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029100012CreateAffiliateLinksTable is SDD 01 §5.2 `affiliate_links`: optional store links behind the `affiliates` feature flag (default off).
func m20261029100012CreateAffiliateLinksTable() schema.Migration {
	return schema.Migration{
		ID: "20261029100012_create_affiliate_links_table",
		Up: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil // scaffold tests run on SQLite (VF-033)
			}
			return s.Raw(`CREATE TABLE affiliate_links (
  item_id uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  store text NOT NULL CHECK (store IN ('amazon','mercadolivre','steam','gog')),
  url text NOT NULL,
  active boolean NOT NULL DEFAULT true,
  PRIMARY KEY (item_id, store)
)`)
		},
		Down: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil
			}
			return s.Raw(`DROP TABLE IF EXISTS affiliate_links`)
		},
	}
}
