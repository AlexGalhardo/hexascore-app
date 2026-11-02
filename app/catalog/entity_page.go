package catalog

import (
	"cmp"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// Sections maps an item type to its web section, the first path segment of its pages.
var Sections = map[string]string{"game": "games", "movie": "movies", "series": "series", "book": "books"}

// EntityPaths maps an entity kind to its path segment under a section: /{section}/{entity path}/{key}
// (owner round 2026-11-01 item 19). Countries and languages are keyed by code, the others by slug.
var EntityPaths = map[string]string{
	"person": "people", "organization": "companies", "franchise": "franchises", "platform": "platforms",
	"genre": "genres", "game_mode": "game-modes", "country": "countries", "language": "languages",
}

var itemTypes = []string{"game", "movie", "series", "book"}

// EntityHome is where an entity page lives without a category in the URL: the item type with the most public
// items of the entity (ties in game, movie, series, book order; "game" when it has none) and its URL key, the
// canonical code for countries and languages. ErrItemNotFound when there is no such entity.
func EntityHome(ctx context.Context, db Querier, kind, key string) (typ, canonicalKey string, err error) {
	match := "e.slug = $2"
	if kind == "country" || kind == "language" {
		match = "lower(e.code) = lower($2)"
	}

	var id string
	// match is one of two constants, never input.
	err = db.QueryRowContext(ctx, `SELECT e.id::text, coalesce(e.code, e.slug) FROM entities e
WHERE e.kind = $1 AND `+match, kind, key).Scan(&id, &canonicalKey)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", ErrItemNotFound
	}
	if err != nil {
		return "", "", fmt.Errorf("load entity: %w", err)
	}

	types, err := EntityTypes(ctx, db, []string{id})
	if err != nil {
		return "", "", err
	}

	return cmp.Or(types[id], itemTypes[0]), canonicalKey, nil
}

// EntityTypes maps each entity id to the item type with the most public items linked to it (ties in game,
// movie, series, book order); ids without public items are absent.
func EntityTypes(ctx context.Context, db Querier, ids []string) (map[string]string, error) {
	out := make(map[string]string, len(ids))
	if len(ids) == 0 {
		return out, nil
	}

	rows, err := db.QueryContext(ctx, `SELECT DISTINCT ON (ie.entity_id) ie.entity_id::text, i.type::text
FROM item_entities ie JOIN items i ON i.id = ie.item_id AND i.visibility = 'public'
WHERE ie.entity_id = ANY($1::uuid[])
GROUP BY ie.entity_id, i.type
ORDER BY ie.entity_id, count(DISTINCT i.id) DESC, array_position(ARRAY['game','movie','series','book'], i.type::text)`,
		"{"+strings.Join(ids, ",")+"}")
	if err != nil {
		return nil, fmt.Errorf("entity types: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, typ string
		if err := rows.Scan(&id, &typ); err != nil {
			return nil, fmt.Errorf("scan entity type: %w", err)
		}
		out[id] = typ
	}

	return out, rows.Err()
}
