package catalog

import (
	"cmp"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

// Sections maps an item type to its web section, the first path segment of its pages.
var Sections = map[string]string{"game": "games", "movie": "movies", "series": "series", "book": "books"}

// EntityPaths maps an entity kind to its path segment under a section: /{section}/{entity path}/{key}
// (owner round 2026-11-01 item 19). Countries and languages are keyed by code, the others by slug.
var EntityPaths = map[string]string{
	"person": "people", "organization": "companies", "franchise": "franchises", "platform": "platforms",
	"genre": "genres", "game_mode": "game-modes", "country": "countries", "language": "languages",
}

// EntityGroupSize is how many items one type group lists per page on Entity/Show.
const EntityGroupSize = 12

// EntityView is the site-audience entity header (SDD 01 §6.2).
type EntityView struct {
	ID          string    `json:"id"`
	Kind        string    `json:"kind"`
	Slug        string    `json:"slug"`
	Code        string    `json:"code,omitempty"`
	Name        string    `json:"name"`
	WikidataQID string    `json:"wikidataQid,omitempty"`
	Description *Synopsis `json:"description,omitempty"`
	BornOn      string    `json:"bornOn,omitempty"`
	DiedOn      string    `json:"diedOn,omitempty"`
	FoundedYear *int      `json:"foundedYear,omitempty"`
}

// RoleCount is how many public items the entity is linked to under one role.
type RoleCount struct {
	Role  string `json:"role"`
	Count int    `json:"count"`
}

// EntityGroup is the entity's public items of one type. Next is the offset of the next page, -1 at the end.
type EntityGroup struct {
	Type  string      `json:"type"`
	Total int         `json:"total"`
	Items []Highlight `json:"items"`
	Next  int         `json:"next"`
}

// EntityPage is everything Entity/Show renders.
type EntityPage struct {
	Entity EntityView
	Roles  []RoleCount
	// Counts is public items per type; AvgScore is the mean visible score per type, present only once
	// entity_stats saw 5 scored items (SDD 01 §7.4).
	Counts   map[string]int
	AvgScore map[string]float64
	Groups   []EntityGroup
	// Indexable is false below MinIndexableItems linked public items (noindex,follow, open issues 01-8, 09-8).
	Indexable bool
}

// LoadEntityPage reads the site page of the entity of kind found by key (slug, or code for countries and
// languages). typ empty lists the first page of every type; typ set pages that type from offset.
// ErrItemNotFound when there is no such entity.
func LoadEntityPage(ctx context.Context, db Querier, kind, key, locale, typ string, offset int) (EntityPage, error) {
	page := EntityPage{Groups: []EntityGroup{}, Counts: map[string]int{}, AvgScore: map[string]float64{}}
	// Codes match case-insensitively so the controller can 301 /countries/br to the canonical /countries/BR.
	match := "e.slug = $2"
	if kind == "country" || kind == "language" {
		match = "lower(e.code) = lower($2)"
	}

	var (
		qid, code, descText, descLic, descURL sql.NullString
		born, died                            sql.NullTime
		founded                               sql.NullInt64
	)
	// match is one of two constants, never input.
	err := db.QueryRowContext(ctx, `SELECT e.id::text, e.slug, e.code, e.wikidata_qid,
  coalesce(t.name, te.name, e.default_name),
  CASE WHEN t.description IS NOT NULL THEN t.description ELSE te.description END,
  CASE WHEN t.description IS NOT NULL THEN t.description_licence ELSE te.description_licence END,
  CASE WHEN t.description IS NOT NULL THEN t.description_source_url ELSE te.description_source_url END,
  e.born_on, e.died_on, e.founded_year
FROM entities e
LEFT JOIN entity_translations t ON t.entity_id = e.id AND t.locale = $3
LEFT JOIN entity_translations te ON te.entity_id = e.id AND te.locale = 'en'
WHERE e.kind = $1 AND `+match+``, kind, key, locale).Scan(
		&page.Entity.ID, &page.Entity.Slug, &code, &qid, &page.Entity.Name,
		&descText, &descLic, &descURL, &born, &died, &founded)
	if errors.Is(err, sql.ErrNoRows) {
		return page, ErrItemNotFound
	}
	if err != nil {
		return page, fmt.Errorf("load entity: %w", err)
	}

	page.Entity.Kind, page.Entity.Code, page.Entity.WikidataQID = kind, code.String, qid.String
	if descText.Valid {
		page.Entity.Description = &Synopsis{Text: descText.String, Licence: descLic.String, SourceURL: descURL.String}
	}
	if born.Valid {
		page.Entity.BornOn = born.Time.Format(time.DateOnly)
	}
	if died.Valid {
		page.Entity.DiedOn = died.Time.Format(time.DateOnly)
	}
	if founded.Valid {
		y := int(founded.Int64)
		page.Entity.FoundedYear = &y
	}

	if err := page.loadCounts(ctx, db); err != nil {
		return page, err
	}
	// A category page (owner round 2026-11-01 item 19) is indexable on its own items, not the entity's total.
	if typ != "" {
		page.Indexable = page.Counts[typ] >= MinIndexableItems
	}

	types := make([]string, 0, len(page.Counts))
	for _, t := range itemTypes {
		if n := page.Counts[t]; n > 0 && (typ == "" || typ == t) {
			types = append(types, t)
		}
	}
	for _, t := range types {
		off := 0
		if typ != "" {
			off = offset
		}
		g, err := loadEntityGroup(ctx, db, page.Entity.ID, t, locale, off, page.Counts[t])
		if err != nil {
			return page, err
		}
		page.Groups = append(page.Groups, g)
	}

	return page, nil
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

func (p *EntityPage) loadCounts(ctx context.Context, db Querier) error {
	rows, err := db.QueryContext(ctx, `SELECT i.type, count(DISTINCT i.id)
FROM item_entities ie JOIN items i ON i.id = ie.item_id AND i.visibility = 'public'
WHERE ie.entity_id = $1 GROUP BY i.type`, p.Entity.ID)
	if err != nil {
		return fmt.Errorf("entity counts: %w", err)
	}
	defer rows.Close()

	total := 0
	for rows.Next() {
		var t string
		var n int
		if err := rows.Scan(&t, &n); err != nil {
			return fmt.Errorf("scan entity count: %w", err)
		}
		p.Counts[t] = n
		total += n
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("entity counts: %w", err)
	}
	p.Indexable = total >= MinIndexableItems

	roles, err := db.QueryContext(ctx, `SELECT ie.role, count(DISTINCT ie.item_id)
FROM item_entities ie JOIN items i ON i.id = ie.item_id AND i.visibility = 'public'
WHERE ie.entity_id = $1 GROUP BY ie.role ORDER BY count(DISTINCT ie.item_id) DESC, ie.role`, p.Entity.ID)
	if err != nil {
		return fmt.Errorf("entity roles: %w", err)
	}
	defer roles.Close()

	p.Roles = []RoleCount{}
	for roles.Next() {
		var r RoleCount
		if err := roles.Scan(&r.Role, &r.Count); err != nil {
			return fmt.Errorf("scan entity role: %w", err)
		}
		p.Roles = append(p.Roles, r)
	}
	if err := roles.Err(); err != nil {
		return fmt.Errorf("entity roles: %w", err)
	}

	stats, err := db.QueryContext(ctx,
		`SELECT type, avg_score FROM entity_stats WHERE entity_id = $1 AND scored_count >= 5 AND avg_score IS NOT NULL`,
		p.Entity.ID)
	if err != nil {
		return fmt.Errorf("entity stats: %w", err)
	}
	defer stats.Close()

	for stats.Next() {
		var t string
		var avg float64
		if err := stats.Scan(&t, &avg); err != nil {
			return fmt.Errorf("scan entity stats: %w", err)
		}
		p.AvgScore[t] = math.Round(avg) / 10 // stored 0-100, shown 0-10 like the item rating summary
	}

	return stats.Err()
}

func loadEntityGroup(
	ctx context.Context,
	db Querier,
	entityID, typ, locale string,
	offset, total int,
) (EntityGroup, error) {
	rows, err := db.QueryContext(ctx, `SELECT i.id::text FROM items i
WHERE i.type = $2 AND i.visibility = 'public'
  AND EXISTS (SELECT 1 FROM item_entities ie WHERE ie.item_id = i.id AND ie.entity_id = $1)
ORDER BY CASE WHEN `+visibleScore+` THEN i.score END DESC NULLS LAST, i.id
OFFSET $3 LIMIT $4`, entityID, typ, offset, EntityGroupSize)
	if err != nil {
		return EntityGroup{}, fmt.Errorf("entity items: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return EntityGroup{}, fmt.Errorf("scan entity item: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return EntityGroup{}, fmt.Errorf("entity items: %w", err)
	}

	cards, err := Summaries(ctx, db, ids, locale)
	if err != nil {
		return EntityGroup{}, err
	}

	next := offset + len(ids)
	if next >= total {
		next = -1
	}

	return EntityGroup{Type: typ, Total: total, Items: cards, Next: next}, nil
}
