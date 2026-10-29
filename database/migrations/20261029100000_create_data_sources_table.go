package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029100000CreateDataSourcesTable: provenance targets for every
// catalogue field (SDD 01 §5.2). Raw DDL: the catalogue is PostgreSQL-only.
func m20261029100000CreateDataSourcesTable() schema.Migration {
	return schema.Migration{
		ID: "20261029100000_create_data_sources_table",
		Up: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil // scaffold tests run on SQLite (VF-033)
			}
			return s.Raw(`CREATE TABLE data_sources (
  code text PRIMARY KEY,
  name text NOT NULL,
  licence text NOT NULL,
  attribution text,
  url text NOT NULL
)`)
		},
		Down: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil
			}
			return s.Raw(`DROP TABLE IF EXISTS data_sources`)
		},
	}
}
