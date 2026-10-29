package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029100003CreateItemAltTitlesTable: aliases, transliterations and
// former titles (open issue 07-6); normalised search columns come in P1.E.1.1.
func m20261029100003CreateItemAltTitlesTable() schema.Migration {
	return schema.Migration{
		ID: "20261029100003_create_item_alt_titles_table",
		Up: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil // scaffold tests run on SQLite (VF-033)
			}
			return s.Raw(`CREATE TABLE item_alt_titles (
  item_id uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  title text NOT NULL CHECK (length(title) BETWEEN 1 AND 500),
  lang text,
  kind text NOT NULL CHECK (kind IN ('alias','transliteration','original','former')),
  source text NOT NULL REFERENCES data_sources(code),
  PRIMARY KEY (item_id, title)
)`)
		},
		Down: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil
			}
			return s.Raw(`DROP TABLE IF EXISTS item_alt_titles`)
		},
	}
}
