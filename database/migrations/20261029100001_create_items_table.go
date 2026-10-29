package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029100001CreateItemsTable is SDD 01 §5.2 `items`: one table for the
// four media types, uuid v7 keys set in Go, jsonb and CHECK enums as raw DDL
// (VF-006, VF-035). Score columns belong to SDD 02, search columns to SDD 07;
// `primary_edition_id` gets its foreign key with `editions` (P1.B.1.3).
func m20261029100001CreateItemsTable() schema.Migration {
	return schema.Migration{
		ID: "20261029100001_create_items_table",
		Up: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil // scaffold tests run on SQLite (VF-033)
			}
			return s.Raw(`CREATE TABLE items (
  id uuid PRIMARY KEY,
  type text NOT NULL CHECK (type IN ('game','movie','series','book')),
  slug text NOT NULL CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$' AND length(slug) <= 120),
  wikidata_qid text NOT NULL UNIQUE CHECK (wikidata_qid ~ '^Q[1-9][0-9]*$'),
  original_title text,
  release_date date,
  release_precision text NOT NULL DEFAULT 'unknown'
    CHECK (release_precision IN ('day','month','year','unknown')),
  release_status text NOT NULL DEFAULT 'released'
    CHECK (release_status IN ('announced','released','cancelled')),
  visibility text NOT NULL DEFAULT 'public'
    CHECK (visibility IN ('public','hidden','takedown')),
  runtime_minutes int CHECK (runtime_minutes BETWEEN 1 AND 6000),
  season_count smallint,
  episode_count int,
  page_count int CHECK (page_count BETWEEN 1 AND 20000),
  primary_edition_id uuid,
  trailers jsonb NOT NULL DEFAULT '[]',
  external_ids jsonb NOT NULL DEFAULT '{}',
  provenance jsonb NOT NULL DEFAULT '{}',
  cover_key text,
  cover_thumb_key text,
  cover_source text REFERENCES data_sources(code),
  cover_attribution jsonb,
  popularity_prior real NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (type, slug),
  UNIQUE (id, type)
);
CREATE INDEX items_type_popularity_idx ON items (type, popularity_prior DESC, id);
CREATE INDEX items_type_release_idx ON items (type, release_date, id);
CREATE INDEX items_external_ids_gin ON items USING gin (external_ids jsonb_path_ops);
CREATE INDEX items_public_idx ON items (type, popularity_prior DESC, id) WHERE visibility = 'public'`)
		},
		Down: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil
			}
			return s.Raw(`DROP TABLE IF EXISTS items`)
		},
	}
}
