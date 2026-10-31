package similar

import (
	"context"

	"github.com/AlexGalhardo/vukem/database"
)

// LoadParents returns the franchise parents (entity_links part_of) of every entity, for ExtractFeatures.
func LoadParents(ctx context.Context, db *database.DB) (map[string][]string, error) {
	return loadParents(ctx, db)
}

// LoadEntries returns the item_entities rows of the given public items of typ, keyed by item id.
func LoadEntries(ctx context.Context, db *database.DB, typ string, ids []string) (map[string][]Entry, error) {
	e, _, err := loadEntries(ctx, db, typ, ids, Options{})

	return e, err
}
