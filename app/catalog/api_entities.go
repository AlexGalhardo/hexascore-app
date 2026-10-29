package catalog

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"
)

// APIEntity is one entity for the paid API after the audience policy. Excluded fields are zero (SDD 01 §6.1).
type APIEntity struct {
	ID          string
	Kind        string
	Slug        string
	WikidataQID string
	Code        string
	Name        string
	Description *Synopsis
	BornOn      string
	DiedOn      string
	FoundedYear *int
	ExternalIDs map[string]string
	Provenance  map[string]FieldSource
	Meta        *Meta
}

// LoadAPIEntity reads entity id filtered by p. kind restricts the entity kind ("" = any); a missing id or a
// different kind is ok=false. Field sources come from entities.provenance and the translation columns; a field
// without a known source is hidden by the policy (fail closed, ADR 0006).
func LoadAPIEntity(ctx context.Context, db Querier, id, kind, locale string, p Policy) (APIEntity, bool, error) {
	var (
		e                               APIEntity
		qid, code, nameSrc              sql.NullString
		desc, descSrc, descLic, descURL sql.NullString
		born, died                      sql.NullTime
		founded                         sql.NullInt64
		extRaw, provRaw                 []byte
	)
	err := db.QueryRowContext(ctx, `SELECT e.id::text, e.kind, e.slug, e.wikidata_qid, e.code,
  coalesce(t.name, te.name, e.default_name),
  CASE WHEN t.name IS NOT NULL THEN t.name_source WHEN te.name IS NOT NULL THEN te.name_source
    ELSE coalesce(e.provenance->>'default_name', 'hexascore') END,
  CASE WHEN t.description IS NOT NULL THEN t.description ELSE te.description END,
  CASE WHEN t.description IS NOT NULL THEN t.description_source ELSE te.description_source END,
  CASE WHEN t.description IS NOT NULL THEN t.description_licence ELSE te.description_licence END,
  CASE WHEN t.description IS NOT NULL THEN t.description_source_url ELSE te.description_source_url END,
  e.born_on, e.died_on, e.founded_year, e.external_ids, e.provenance
FROM entities e
LEFT JOIN entity_translations t ON t.entity_id = e.id AND t.locale = $2
LEFT JOIN entity_translations te ON te.entity_id = e.id AND te.locale = 'en'
WHERE e.id = $1::uuid AND ($3 = '' OR e.kind = $3)`, id, locale, kind).Scan(
		&e.ID, &e.Kind, &e.Slug, &qid, &code, &e.Name, &nameSrc, &desc, &descSrc, &descLic, &descURL,
		&born, &died, &founded, &extRaw, &provRaw)
	if errors.Is(err, sql.ErrNoRows) {
		return APIEntity{}, false, nil
	}
	if err != nil {
		return APIEntity{}, false, fmt.Errorf("load api entity: %w", err)
	}

	prov := map[string]string{}
	if err := json.Unmarshal(provRaw, &prov); err != nil {
		return APIEntity{}, false, fmt.Errorf("entity %s provenance: %w", e.ID, err)
	}
	ext, err := externalIDs(extRaw)
	if err != nil {
		return APIEntity{}, false, fmt.Errorf("entity %s external ids: %w", e.ID, err)
	}
	licences, err := sourceLicences(ctx, db)
	if err != nil {
		return APIEntity{}, false, err
	}

	var omitted []string
	e.WikidataQID, e.Code = qid.String, code.String
	e.Provenance = map[string]FieldSource{}
	keep := func(field, source string) {
		e.Provenance[field] = FieldSource{Source: source, Licence: licences[source]}
	}

	if p.Allows(nameSrc.String) {
		keep("name", nameSrc.String)
	} else {
		e.Name = ""
		omitted = append(omitted, "name")
	}
	if desc.Valid {
		if p.AllowsContent(descSrc.String) {
			e.Description = &Synopsis{Text: desc.String, Licence: descLic.String, SourceURL: descURL.String}
			keep("description", descSrc.String)
		} else {
			omitted = append(omitted, "description")
		}
	}
	dated := func(field string, v sql.NullTime, set func(string)) {
		if !v.Valid {
			return
		}
		if p.AllowsContent(prov[field]) {
			set(v.Time.Format(time.DateOnly))
			keep(camel(field), prov[field])
		} else {
			omitted = append(omitted, camel(field))
		}
	}
	dated("born_on", born, func(s string) { e.BornOn = s })
	dated("died_on", died, func(s string) { e.DiedOn = s })
	if founded.Valid {
		if p.AllowsContent(prov["founded_year"]) {
			y := int(founded.Int64)
			e.FoundedYear = &y
			keep("foundedYear", prov["founded_year"])
		} else {
			omitted = append(omitted, "foundedYear")
		}
	}
	for key, value := range ext {
		if slices.Contains(identifierKeys, key) || p.Allows(key) {
			if e.ExternalIDs == nil {
				e.ExternalIDs = map[string]string{}
			}
			e.ExternalIDs[key] = value
		} else if !slices.Contains(omitted, "externalIds") {
			omitted = append(omitted, "externalIds")
		}
	}

	if len(omitted) > 0 {
		slices.Sort(omitted)
		e.Meta = &Meta{OmittedFields: omitted}
	}

	return e, true, nil
}

func camel(snake string) string {
	switch snake {
	case "born_on":
		return "bornOn"
	case "died_on":
		return "diedOn"
	}

	return snake
}
