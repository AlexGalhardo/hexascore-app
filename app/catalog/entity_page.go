package catalog

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"
)

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
	page := EntityPage{Counts: map[string]int{}, AvgScore: map[string]float64{}}
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
