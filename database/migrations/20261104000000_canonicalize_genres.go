package migrations

import (
	"github.com/AlexGalhardo/vukem/schema"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

// m20261104000000CanonicalizeGenres makes the canonical genres permanent (owner 2026-11-04). genres:canonicalize was
// a console command run by hand in production; the 2026-11-03 data cleanup lost its result, and every new ingest
// brought the Wikidata labels back. This migration applies the same repair as SQL: every source genre link moved to its
// canonical genre (created with its three names when some item needs it) or hidden. The normaliser now does the same
// for each item it saves (catalog.CanonicalizeItemGenres), so nothing needs to run again. Idempotent.
func m20261104000000CanonicalizeGenres() schema.Migration {
	return schema.Migration{
		ID: "20261104000000_canonicalize_genres",
		Up: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil
			}

			for _, st := range catalog.GenreRepairStatements() {
				if err := s.Raw(st.SQL, st.Args...); err != nil {
					return err
				}
			}

			return nil
		},
		// Rolling back leaves the canonical genres in place: the source links are kept in item_genre_sources.
		Down: func(*schema.Schema) error { return nil },
	}
}
