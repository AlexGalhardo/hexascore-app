// Package related serves the "More like this" and cross-media blocks of the item page (SDD 08 sections 6 and 7.4).
package related

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"strings"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/app/recommend/similar"
)

const (
	// SimilarShown is how many of the stored neighbours the page shows.
	SimilarShown = 8
	// CrossMediaShown caps the cross-media block.
	CrossMediaShown = 6
)

// Reason is the top feature the two items share.
type Reason struct {
	Kind   string `json:"kind"`
	Entity string `json:"entity"`
}

// SimilarItem is one neighbour card. Reason is null when no direct feature explains the match.
type SimilarItem struct {
	catalog.Highlight
	Reason *Reason `json:"reason"`
}

// CrossItem is an item linked by item_links in either direction. Relation is "based_on" when the page item is based
// on it and "adapted_as" when it is based on the page item.
type CrossItem struct {
	catalog.Highlight
	Relation string `json:"relation"`
}

// Related is the pair of blocks; both lists are empty, never nil.
type Related struct {
	Similar    []SimilarItem `json:"similar"`
	CrossMedia []CrossItem   `json:"crossMedia"`
}

// Load builds the blocks for one item of type typ.
// shortcut: the reason is the shared feature with the highest role weight, not the highest idf-weighted g_f; upgrade if
// the two disagree visibly.
func Load(ctx context.Context, db *database.DB, itemID, typ, locale string) (Related, error) {
	out := Related{Similar: []SimilarItem{}, CrossMedia: []CrossItem{}}

	var stored string

	err := db.QueryRowContext(ctx, `SELECT similar_ids::text FROM item_similar WHERE item_id = $1`, itemID).
		Scan(&stored)
	if err != nil && !isNoRows(err) {
		return out, err
	}

	if ids := splitArray(stored); len(ids) > 0 {
		cards, err := catalog.HighlightsByID(ctx, db, ids, locale)
		if err != nil {
			return out, err
		}

		if len(cards) > SimilarShown {
			cards = cards[:SimilarShown]
		}

		reasons, err := reasonsFor(ctx, db, itemID, typ, cards)
		if err != nil {
			return out, err
		}

		for _, c := range cards {
			out.Similar = append(out.Similar, SimilarItem{Highlight: c, Reason: reasons[c.ID]})
		}
	}

	cross, err := crossMedia(ctx, db, itemID, locale)
	if err != nil {
		return out, err
	}

	out.CrossMedia = cross

	return out, nil
}

func crossMedia(ctx context.Context, db *database.DB, itemID, locale string) ([]CrossItem, error) {
	rows, err := db.QueryContext(ctx, `SELECT l.other::text, l.relation FROM (
  SELECT to_item_id AS other, 'based_on' AS relation, from_item_id FROM item_links WHERE from_item_id = $1
  UNION ALL
  SELECT from_item_id, 'adapted_as', to_item_id FROM item_links WHERE to_item_id = $1) l
JOIN items i ON i.id = l.other AND i.visibility = 'public'
ORDER BY i.rating_count DESC, i.id LIMIT $2`, itemID, CrossMediaShown)
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	var ids []string

	rel := map[string]string{}

	for rows.Next() {
		var id, r string
		if err := rows.Scan(&id, &r); err != nil {
			return nil, err
		}

		ids, rel[id] = append(ids, id), r
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	cards, err := catalog.HighlightsByID(ctx, db, ids, locale)
	if err != nil {
		return nil, err
	}

	out := make([]CrossItem, 0, len(cards))
	for _, c := range cards {
		out = append(out, CrossItem{Highlight: c, Relation: rel[c.ID]})
	}

	return out, nil
}

func reasonsFor(
	ctx context.Context,
	db *database.DB,
	itemID, typ string,
	cards []catalog.Highlight,
) (map[string]*Reason, error) {
	out := map[string]*Reason{}
	if len(cards) == 0 {
		return out, nil
	}

	ids := []string{itemID}
	for _, c := range cards {
		ids = append(ids, c.ID)
	}

	entries, err := similar.LoadEntries(ctx, db, typ, ids)
	if err != nil {
		return nil, err
	}

	parents, err := similar.LoadParents(ctx, db)
	if err != nil {
		return nil, err
	}

	self := similar.ExtractFeatures(entries[itemID], parents)
	top := map[string]similar.Feature{}

	var need []string

	for _, c := range cards {
		best, ok := topShared(self, similar.ExtractFeatures(entries[c.ID], parents))
		if ok {
			top[c.ID] = best
			need = append(need, best.Entity)
		}
	}

	names, err := entityNames(ctx, db, need)
	if err != nil {
		return nil, err
	}

	for id, f := range top {
		out[id] = &Reason{Kind: f.Role, Entity: names[f.Entity]}
	}

	return out, nil
}

// topShared picks the shared feature with the highest role weight times multiplier; ties go to the lowest entity id.
func topShared(a, b map[similar.Feature]float64) (similar.Feature, bool) {
	var shared []similar.Feature

	for f := range a {
		if _, ok := b[f]; ok {
			shared = append(shared, f)
		}
	}

	if len(shared) == 0 {
		return similar.Feature{}, false
	}

	w := func(f similar.Feature) float64 { return similar.Omega(f.Role) * min(a[f], b[f]) }

	sort.Slice(shared, func(i, j int) bool {
		if w(shared[i]) != w(shared[j]) {
			return w(shared[i]) > w(shared[j])
		}

		return shared[i].Entity < shared[j].Entity
	})

	return shared[0], true
}

func entityNames(ctx context.Context, db *database.DB, ids []string) (map[string]string, error) {
	out := map[string]string{}
	if len(ids) == 0 {
		return out, nil
	}

	rows, err := db.QueryContext(ctx, `SELECT id::text, default_name FROM entities WHERE id = ANY($1::uuid[])`,
		"{"+strings.Join(ids, ",")+"}")
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}

		out[id] = name
	}

	return out, rows.Err()
}

func splitArray(s string) []string {
	s = strings.Trim(s, "{}")
	if s == "" {
		return nil
	}

	return strings.Split(s, ",")
}

func isNoRows(err error) bool { return errors.Is(err, sql.ErrNoRows) }
