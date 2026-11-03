package migrations

import (
	"fmt"

	"github.com/AlexGalhardo/vukem/schema"
)

// uniqueExternalIDs are the external ids that identify one item (RFC-001 section 5 C): one value maps to at most one
// item. The TMDB id is per media type (a film and a series may share a number), so its index includes the type.
var uniqueExternalIDs = []struct {
	key    string
	byType bool
}{
	{"epic", false},
	{"gog", false},
	{"igdb", false},
	{"imdb", false},
	{"openlibrary", false},
	{"steam", false},
	{"tmdb", true},
}

// externalIDIndex builds the DO block that creates the unique partial index of one external id, but only when the
// table has no duplicate yet: a unique index over data that already breaks the rule would fail the deploy, so the
// migration notes it and leaves the index to be created after the duplicates are merged (the normaliser enforces
// the rule in code meanwhile). The index is not CONCURRENT because the check and the create must share one
// transaction; items holds about 100 thousand rows, which build in well under a second.
func externalIDIndex(key string, byType bool) string {
	scope, cols := "", fmt.Sprintf("(external_ids->>'%s')", key)
	if byType {
		scope, cols = "type, ", "type, "+cols
	}

	name := "items_ext_" + key + "_uidx"

	return fmt.Sprintf(`DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM items WHERE (external_ids->>'%[1]s') <> '' GROUP BY %[2]s(external_ids->>'%[1]s') HAVING count(*) > 1) THEN
    RAISE NOTICE '%[3]s not created: items share a %[1]s id; merge them, then create the index';
  ELSE
    CREATE UNIQUE INDEX IF NOT EXISTS %[3]s ON items (%[4]s) WHERE (external_ids->>'%[1]s') <> '';
  END IF;
END $$`, key, scope, name, cols)
}

// m20261103020000CreateItemConflicts adds the review queue for items the normaliser will not create on its own
// (RFC-001 section 5 C) and the unique indexes on external ids.
//
// item_conflicts has one row per (incoming Wikidata QID, kind, key, existing item). kind external_id means the QID
// carries an external id (key, value) that the existing item owns; kind title_year means its English title (value) and
// release year are within one year of the existing item's and the trigram similarity is at least 0.92. status open
// waits for an admin: allowed lets the QID in (without the contested id), dismissed keeps it out. Written by the
// normaliser (internal/ingest/resolve), resolved on /admin/ingest/conflicts; resolved_by is an admin's users.id kept as
// a bare id (the purge keeps the users row). One row per contested QID, so the table stays small and is not pruned.
func m20261103020000CreateItemConflicts() schema.Migration {
	up := []string{
		`CREATE TABLE item_conflicts (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  wikidata_qid text NOT NULL CHECK (wikidata_qid ~ '^Q[1-9][0-9]*$'),
  item_type text NOT NULL CHECK (item_type IN ('game', 'movie', 'series', 'book')),
  kind text NOT NULL CHECK (kind IN ('external_id', 'title_year')),
  key text NOT NULL DEFAULT '',
  value text NOT NULL DEFAULT '',
  other_item_id uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  similarity real,
  status text NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'allowed', 'dismissed')),
  created_at timestamptz NOT NULL DEFAULT now(),
  resolved_at timestamptz,
  resolved_by bigint,
  UNIQUE (wikidata_qid, kind, key, other_item_id)
)`,
		// serves the admin queue: the oldest open conflicts first
		`CREATE INDEX item_conflicts_open_idx ON item_conflicts (created_at) WHERE status = 'open'`,
	}

	down := []string{`DROP TABLE IF EXISTS item_conflicts`}

	for _, u := range uniqueExternalIDs {
		up = append(up, externalIDIndex(u.key, u.byType))
		down = append([]string{fmt.Sprintf("DROP INDEX IF EXISTS items_ext_%s_uidx", u.key)}, down...)
	}

	return postgresMigration("20261103020000_create_item_conflicts", up, down)
}
