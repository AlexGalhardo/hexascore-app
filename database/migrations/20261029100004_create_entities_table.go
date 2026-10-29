package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029100004CreateEntitiesTable is SDD 01 §5.2 `entities`: people, organisations and the shared taxonomies in one table; `wikidata_qid` is null for HexaScore-native rows.
func m20261029100004CreateEntitiesTable() schema.Migration {
	return schema.Migration{
		ID: "20261029100004_create_entities_table",
		Up: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil // scaffold tests run on SQLite (VF-033)
			}
			return s.Raw(`CREATE TABLE entities (
  id uuid PRIMARY KEY,
  kind text NOT NULL CHECK (kind IN ('person','organization','franchise','platform',
    'genre','game_mode','country','language')),
  slug text NOT NULL CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$' AND length(slug) <= 120),
  wikidata_qid text UNIQUE CHECK (wikidata_qid ~ '^Q[1-9][0-9]*$'),
  code text,
  default_name text NOT NULL,
  external_ids jsonb NOT NULL DEFAULT '{}',
  image_key text,
  image_source text REFERENCES data_sources(code),
  born_on date,
  died_on date,
  founded_year smallint,
  provenance jsonb NOT NULL DEFAULT '{}',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (kind, slug),
  UNIQUE (kind, code)
)`)
		},
		Down: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil
			}
			return s.Raw(`DROP TABLE IF EXISTS entities`)
		},
	}
}
