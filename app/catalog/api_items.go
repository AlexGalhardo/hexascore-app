package catalog

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

// APIItem is one public item for the paid API: the licence-filtered read
// model plus the SDD 02 §6 score fields.
type APIItem struct {
	Detail ItemDetail
	Score  RatingSummary
}

// LoadAPIItems reads the public items with ids, in the order given, filtered
// by p. Missing or non-public ids are skipped. Relations load only when
// withEntities is set (detail, not listings). Field sources come from
// items.provenance and the translation columns; a field without a known
// source is hidden by the policy (fail closed, ADR 0006).
func LoadAPIItems(
	ctx context.Context, db Querier, ids []string, locale string, p Policy, withEntities bool,
) ([]APIItem, error) {
	out := []APIItem{}
	if len(ids) == 0 {
		return out, nil
	}

	licences, err := sourceLicences(ctx, db)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, `SELECT i.id::text, i.type, i.slug, i.wikidata_qid, i.release_date,
  i.release_precision, i.runtime_minutes, i.page_count, i.provenance, i.external_ids, i.cover_key IS NOT NULL,
  coalesce(t.title, te.title, i.original_title, i.slug),
  CASE WHEN t.title IS NOT NULL THEN t.title_source WHEN te.title IS NOT NULL THEN te.title_source
    WHEN i.original_title IS NOT NULL THEN i.provenance->>'original_title' ELSE 'hexascore' END,
  coalesce(t.synopsis, te.synopsis),
  CASE WHEN t.synopsis IS NOT NULL THEN t.synopsis_source ELSE te.synopsis_source END,
  CASE WHEN t.synopsis IS NOT NULL THEN t.synopsis_licence ELSE te.synopsis_licence END,
  CASE WHEN t.synopsis IS NOT NULL THEN t.synopsis_source_url ELSE te.synopsis_source_url END,
  i.rating_count, i.score, i.score_c1, i.score_c2, i.score_c3, i.score_c4, i.score_c5, i.score_c6,
  i.score_confidence, i.score_formula_version
FROM unnest($1::uuid[]) WITH ORDINALITY AS o(id, n)
JOIN items i ON i.id = o.id AND i.visibility = 'public'
LEFT JOIN item_translations t ON t.item_id = i.id AND t.locale = $2
LEFT JOIN item_translations te ON te.item_id = i.id AND te.locale = 'en'
ORDER BY o.n`, "{"+strings.Join(ids, ",")+"}", locale)
	if err != nil {
		return nil, fmt.Errorf("load api items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			src                         ItemSource
			release                     sql.NullTime
			runtime, pages              sql.NullInt64
			provRaw, extRaw             []byte
			hasCover                    bool
			title, tsrc                 sql.NullString
			syn, synSrc, synLic, synURL sql.NullString
			count                       int
			score                       sql.NullFloat64
			cats                        [6]sql.NullFloat64
			conf                        sql.NullString
			version                     sql.NullInt64
		)
		if err := rows.Scan(&src.ID, &src.Type, &src.Slug, &src.WikidataQID, &release, &src.ReleasePrecision,
			&runtime, &pages, &provRaw, &extRaw, &hasCover, &title, &tsrc, &syn, &synSrc, &synLic, &synURL,
			&count, &score, &cats[0], &cats[1], &cats[2], &cats[3], &cats[4], &cats[5], &conf, &version); err != nil {
			return nil, fmt.Errorf("scan api item: %w", err)
		}

		prov := map[string]string{}
		if err := json.Unmarshal(provRaw, &prov); err != nil {
			return nil, fmt.Errorf("item %s provenance: %w", src.ID, err)
		}
		if src.ExternalIDs, err = externalIDs(extRaw); err != nil {
			return nil, fmt.Errorf("item %s external ids: %w", src.ID, err)
		}

		src.Title = Sourced[string]{Value: title.String, Source: tsrc.String}
		if syn.Valid {
			src.Synopsis = &SynopsisIn{
				Text:      syn.String,
				Source:    synSrc.String,
				Licence:   synLic.String,
				SourceURL: synURL.String,
			}
		}
		if release.Valid {
			src.ReleaseDate = &Sourced[time.Time]{Value: release.Time, Source: prov["release_date"]}
		}
		if runtime.Valid {
			src.RuntimeMinutes = &Sourced[int]{Value: int(runtime.Int64), Source: prov["runtime_minutes"]}
		}
		if pages.Valid {
			src.PageCount = &Sourced[int]{Value: int(pages.Int64), Source: prov["page_count"]}
		}
		if hasCover {
			// shortcut: covers have no public URL builder yet, so an empty source hides them on the API;
			// pass cover_source and the URL once covers are served (SDD 01 §6.1).
			src.Cover = &Sourced[string]{}
		}

		item := APIItem{Detail: Assemble(src, p), Score: scoreFields(src.Type, count, score, cats, conf, version)}
		for field, fs := range item.Detail.Provenance {
			fs.Licence = licences[fs.Source]
			item.Detail.Provenance[field] = fs
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("load api items: %w", err)
	}

	if withEntities {
		for i := range out {
			if err := addEntities(ctx, db, &out[i], locale, p); err != nil {
				return nil, err
			}
		}
	}

	return out, nil
}

// addEntities reloads the item through Assemble with its relations, so the
// relation rows go through the same policy as every other field.
func addEntities(ctx context.Context, db Querier, item *APIItem, locale string, p Policy) error {
	rows, err := db.QueryContext(
		ctx,
		`SELECT e.id::text, e.kind, ie.role, coalesce(t.name, te.name, e.default_name),
  CASE WHEN t.name IS NOT NULL THEN t.name_source WHEN te.name IS NOT NULL THEN te.name_source
    ELSE e.provenance->>'default_name' END, ie.source
FROM item_entities ie
JOIN entities e ON e.id = ie.entity_id
LEFT JOIN entity_translations t ON t.entity_id = e.id AND t.locale = $2
LEFT JOIN entity_translations te ON te.entity_id = e.id AND te.locale = 'en'
WHERE ie.item_id = $1
ORDER BY ie.role, ie.position, e.slug`,
		item.Detail.ID,
		locale,
	)
	if err != nil {
		return fmt.Errorf("load api entities: %w", err)
	}
	defer rows.Close()

	var src ItemSource
	for rows.Next() {
		var (
			e  EntityRefIn
			ns sql.NullString
		)
		if err := rows.Scan(&e.ID, &e.Kind, &e.Role, &e.Name, &ns, &e.Source); err != nil {
			return fmt.Errorf("scan api entity: %w", err)
		}
		e.NameSource = ns.String
		src.Entities = append(src.Entities, e)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("load api entities: %w", err)
	}

	rel := Assemble(src, p)
	item.Detail.Entities = rel.Entities
	// src carries only relations: keep the relation omission, not the empty title's.
	if rel.Meta != nil {
		for _, f := range rel.Meta.OmittedFields {
			if f == "entities" || f == "entities.name" {
				if item.Detail.Meta == nil {
					item.Detail.Meta = &Meta{}
				}
				item.Detail.Meta.OmittedFields = append(item.Detail.Meta.OmittedFields, f)
				slices.Sort(item.Detail.Meta.OmittedFields)
			}
		}
	}

	return nil
}

// externalIDs decodes items.external_ids; numeric ids become strings.
func externalIDs(raw []byte) (map[string]string, error) {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		switch x := v.(type) {
		case string:
			out[k] = x
		case float64:
			out[k] = fmt.Sprintf("%.0f", x)
		}
	}
	return out, nil
}

// sourceLicences maps data_sources codes to their licence (a small table).
func sourceLicences(ctx context.Context, db Querier) (map[string]string, error) {
	rows, err := db.QueryContext(ctx, `SELECT code, licence FROM data_sources`)
	if err != nil {
		return nil, fmt.Errorf("load data sources: %w", err)
	}
	defer rows.Close()

	out := map[string]string{}
	for rows.Next() {
		var code, licence string
		if err := rows.Scan(&code, &licence); err != nil {
			return nil, fmt.Errorf("scan data source: %w", err)
		}
		out[code] = licence
	}
	return out, rows.Err()
}

// APIEdition is one book edition for the paid API.
type APIEdition struct {
	Edition
	OmittedFields []string
}

// LoadAPIEditions pages the editions of the public book workID by id after
// the edition id `after` (empty = first page). Titles need an
// allowed source, descriptive fields an allowed content source (open issue
// 01-2). ErrItemNotFound when the work is not a public book.
func LoadAPIEditions(
	ctx context.Context, db Querier, workID, after string, limit int, p Policy,
) (editions []APIEdition, hasMore bool, err error) {
	var exists bool
	if err := db.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM items WHERE id = $1 AND type = 'book' AND visibility = 'public')`, workID).
		Scan(&exists); err != nil {
		return nil, false, fmt.Errorf("check book: %w", err)
	}
	if !exists {
		return nil, false, ErrItemNotFound
	}

	q := `SELECT e.id::text, e.title, e.subtitle, e.language, p.default_name, e.publication_year, e.page_count,
  e.format, e.isbn13, e.source
FROM editions e LEFT JOIN entities p ON p.id = e.publisher_id
WHERE e.work_id = $1 AND ($2 = '' OR e.id > NULLIF($2, '')::uuid)
ORDER BY e.id LIMIT $3`
	rows, err := db.QueryContext(ctx, q, workID, after, limit+1)
	if err != nil {
		return nil, false, fmt.Errorf("load api editions: %w", err)
	}
	defer rows.Close()

	editions = []APIEdition{}
	for rows.Next() {
		var source string
		e, err := scanEdition(scanTail{rows, &source})
		if err != nil {
			return nil, false, fmt.Errorf("scan api edition: %w", err)
		}
		editions = append(editions, filterEdition(e, source, p))
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("load api editions: %w", err)
	}

	if len(editions) > limit {
		return editions[:limit], true, nil
	}
	return editions, false, nil
}

// scanTail lets scanEdition read an extra trailing column.
type scanTail struct {
	rows *sql.Rows
	tail *string
}

func (s scanTail) Scan(dest ...any) error { return s.rows.Scan(append(dest, s.tail)...) }

func filterEdition(e Edition, source string, p Policy) APIEdition {
	// isbn13 is an identifier, exposed like the item externalIds identifier keys.
	var out APIEdition
	out.ID, out.Language, out.ISBN13 = e.ID, e.Language, e.ISBN13
	if p.Allows(source) {
		out.Title = e.Title
	} else {
		out.OmittedFields = append(out.OmittedFields, "title")
	}
	if p.AllowsContent(source) {
		out.Subtitle, out.Publisher, out.PublicationYear, out.PageCount, out.Format =
			e.Subtitle, e.Publisher, e.PublicationYear, e.PageCount, e.Format
	} else {
		out.OmittedFields = append(out.OmittedFields, "format", "pageCount", "publicationYear", "publisher", "subtitle")
	}
	return out
}

// LookupItemID finds the public item id of a third-party id. Wikidata matches items.wikidata_qid; other sources
// match items.external_ids by GIN containment (string and numeric forms), and isbn13 also matches the work of an
// edition. source comes from an allowlist and is bound, not concatenated. Each source gets its own indexable
// branch: one WHERE with ORs across them would scan every item.
func LookupItemID(ctx context.Context, db Querier, source, id string) (string, bool, error) {
	match := `SELECT id FROM items WHERE external_ids @> jsonb_build_object($1::text, $2::text)
UNION ALL SELECT id FROM items
WHERE external_ids @> jsonb_build_object($1::text, CASE WHEN $2 ~ '^[0-9]{1,15}$' THEN $2::bigint END)`
	args := []any{source, id}
	switch source {
	case "wikidata":
		match, args = `SELECT id FROM items WHERE wikidata_qid = $1`, []any{id}
	case "isbn13":
		// char(13) cast keeps the editions index usable; the length guard stops it truncating longer input.
		match += `
UNION ALL SELECT work_id FROM editions WHERE length($2) = 13 AND isbn13 = $2::char(13)`
	}

	var out string
	err := db.QueryRowContext(ctx, `SELECT id::text FROM items WHERE visibility = 'public' AND id IN (`+match+`)
ORDER BY id LIMIT 1`, args...).Scan(&out)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("lookup item: %w", err)
	}

	return out, true, nil
}
