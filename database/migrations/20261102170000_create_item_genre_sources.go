package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261102170000CreateItemGenreSources keeps the source genres (Wikidata labels such as "comedy film") that
// genres:canonicalize replaces with canonical genres (owner 2026-11-02). Nothing on the site or the API reads it;
// it is the record of where each canonical link came from and the input for widening the mapping.
func m20261102170000CreateItemGenreSources() schema.Migration {
	return postgresMigration("20261102170000_create_item_genre_sources",
		[]string{
			`CREATE TABLE item_genre_sources (
  item_id uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  entity_id uuid NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
  source text NOT NULL REFERENCES data_sources(code),
  position smallint NOT NULL DEFAULT 0,
  code text,
  PRIMARY KEY (item_id, entity_id)
)`,
			`CREATE INDEX item_genre_sources_unmapped_idx ON item_genre_sources (entity_id) WHERE code IS NULL`,
		},
		// Rolling back puts the source genres back on the items, beside the canonical ones.
		[]string{
			`INSERT INTO item_entities (item_id, entity_id, role, position, source)
SELECT item_id, entity_id, 'genre', position, source FROM item_genre_sources ON CONFLICT DO NOTHING`,
			`DROP TABLE IF EXISTS item_genre_sources`,
		},
	)
}
