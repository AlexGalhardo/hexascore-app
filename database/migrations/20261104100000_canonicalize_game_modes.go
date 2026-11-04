package migrations

import (
	"github.com/AlexGalhardo/vukem/schema"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

// m20261104100000CanonicalizeGameModes folds the Wikidata game mode labels ("single-player video game", "multiplayer
// online game", "shared/splitscreen multiplayer") into a few canonical modes named with a capital first letter in
// three locales (owner 2026-11-04). Like the genres migration it is plain SQL with the Go rules, creates a canonical
// mode only when some game needs it, keeps the source link in item_genre_sources and hides a label no rule maps. The
// normaliser does the same for every item it saves, so nothing needs to run again. Idempotent.
func m20261104100000CanonicalizeGameModes() schema.Migration {
	return schema.Migration{
		ID: "20261104100000_canonicalize_game_modes",
		Up: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil
			}

			for _, st := range catalog.GameModeRepairStatements() {
				if err := s.Raw(st.SQL, st.Args...); err != nil {
					return err
				}
			}

			return nil
		},
		// Rolling back leaves the canonical modes in place: the source links are kept in item_genre_sources.
		Down: func(*schema.Schema) error { return nil },
	}
}
