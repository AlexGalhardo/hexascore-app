package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029100011CreateEditionsTable is SDD 01 §5.2 `editions`: book editions under a book work. Language `und` and a null publisher are valid (open issue 01-1, DQ-10). `items.primary_edition_id` gets its composite foreign key here so I-3 (the primary edition belongs to the same work) holds in the database.
func m20261029100011CreateEditionsTable() schema.Migration {
	return schema.Migration{
		ID: "20261029100011_create_editions_table",
		Up: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil // scaffold tests run on SQLite (VF-033)
			}
			return s.Raw(`CREATE TABLE editions (
  id uuid PRIMARY KEY,
  work_id uuid NOT NULL,
  work_type text NOT NULL DEFAULT 'book' CHECK (work_type = 'book'),
  FOREIGN KEY (work_id, work_type) REFERENCES items(id, type) ON DELETE CASCADE,
  isbn13 char(13) UNIQUE CHECK (isbn13 IS NULL OR isbn13_is_valid(isbn13)),
  wikidata_qid text UNIQUE,
  openlibrary_id text UNIQUE,
  title text NOT NULL,
  subtitle text,
  language text NOT NULL,
  market_country char(2),
  is_brazilian boolean GENERATED ALWAYS AS (market_country = 'BR' OR language = 'pt-BR') STORED,
  publisher_id uuid REFERENCES entities(id),
  imprint_id uuid REFERENCES entities(id),
  publication_year smallint CHECK (publication_year BETWEEN 1400 AND 2100),
  page_count int,
  format text CHECK (format IN ('hardcover','paperback','ebook','audiobook','other')),
  cover_key text,
  cover_thumb_key text,
  cover_source text REFERENCES data_sources(code),
  source text NOT NULL REFERENCES data_sources(code),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (id, work_id),
  CHECK (isbn13 IS NOT NULL OR wikidata_qid IS NOT NULL OR openlibrary_id IS NOT NULL)
);
CREATE INDEX editions_work_idx ON editions (work_id, is_brazilian DESC, publication_year DESC);
ALTER TABLE items ADD CONSTRAINT items_primary_edition_fk
  FOREIGN KEY (primary_edition_id, id) REFERENCES editions(id, work_id)
  ON DELETE SET NULL (primary_edition_id)`)
		},
		Down: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil
			}
			return s.Raw(`ALTER TABLE items DROP CONSTRAINT IF EXISTS items_primary_edition_fk;
DROP TABLE IF EXISTS editions`)
		},
	}
}
