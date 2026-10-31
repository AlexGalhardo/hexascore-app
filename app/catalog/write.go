package catalog

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/AlexGalhardo/vukem/database"
)

// SourceCommunity is the data_sources code of community-edited fields (SDD 01 licensing, SDD 04 section 7.3).
const SourceCommunity = "community"

// ErrUnsupportedField is a field the community write path cannot store on that target.
var ErrUnsupportedField = errors.New("catalog: field not writable on this target")

// ErrTargetGone is a write on a row that no longer exists.
var ErrTargetGone = errors.New("catalog: target row not found")

// EnsureSource inserts the data_sources row a write path labels its fields with. Every provenance value and
// *_source column must exist there (DQ-11, foreign keys).
func EnsureSource(ctx context.Context, ex database.Executor, code string) error {
	_, err := ex.ExecContext(ctx, `INSERT INTO data_sources (code, name, licence, url)
VALUES ($1, 'HexaScore community', 'proprietary', 'https://hexascore.app') ON CONFLICT (code) DO NOTHING`, code)
	if err != nil {
		return fmt.Errorf("catalog: ensure source %s: %w", code, err)
	}

	return nil
}

// externalColumns maps edition external id providers to their columns. Values are constants, never input.
var editionExternal = map[string]string{"isbn13": "isbn13", "wikidata": "wikidata_qid", "openlibrary": "openlibrary_id"}

// WriteField stores after as the new value of one scalar suggestion field (SDD 04 section 7.2 names) and labels it
// with source. It returns the field's previous provenance as JSON (a string, or null when none was recorded).
// The caller holds the target row lock.
func WriteField(
	ctx context.Context, ex database.Executor, targetType, targetID, field string, after json.RawMessage, source string,
) (json.RawMessage, error) {
	switch targetType {
	case "item":
		return writeItem(ctx, ex, targetID, field, after, source)
	case "entity":
		return writeEntity(ctx, ex, targetID, field, after, source)
	case "edition":
		return writeEdition(ctx, ex, targetID, field, after)
	}

	return nil, ErrUnsupportedField
}

func writeItem(
	ctx context.Context, ex database.Executor, id, field string, after json.RawMessage, source string,
) (json.RawMessage, error) {
	if locale, ok := strings.CutPrefix(field, "translations."); ok {
		locale, rest, _ := strings.Cut(locale, ".")
		if rest != "title" {
			return nil, ErrUnsupportedField
		}

		return writeItemTitle(ctx, ex, id, locale, after, source)
	}

	if locale, ok := strings.CutPrefix(field, "synopsis."); ok {
		return writeItemSynopsis(ctx, ex, id, locale, after, source)
	}

	switch {
	case field == "title":
		text, null, err := textValue(after)
		if err != nil {
			return nil, err
		}

		return provWrite(ctx, ex, "items", id, "original_title", source, null,
			`UPDATE items SET original_title = $2, updated_at = now() WHERE id = $1`, id, nullable(text, null))
	case field == "release_date":
		var d struct{ Value, Precision string }
		null := string(after) == "null" || len(after) == 0
		if !null {
			if err := json.Unmarshal(after, &d); err != nil {
				return nil, fmt.Errorf("catalog: release_date value: %w", err)
			}
		} else {
			d.Precision = "unknown"
		}

		return provWrite(ctx, ex, "items", id, "release_date", source, null,
			`UPDATE items SET release_date = $2::date, release_precision = $3, updated_at = now() WHERE id = $1`,
			id, nullable(d.Value, null), d.Precision)
	case strings.HasPrefix(field, "external_ids."):
		return writeExternalJSON(ctx, ex, "items", id, field, after, source)
	}

	return nil, ErrUnsupportedField
}

func writeEntity(
	ctx context.Context, ex database.Executor, id, field string, after json.RawMessage, source string,
) (json.RawMessage, error) {
	switch {
	case field == "title":
		text, null, err := textValue(after)
		if err != nil {
			return nil, err
		}

		if null {
			return nil, ErrUnsupportedField
		}

		return provWrite(ctx, ex, "entities", id, "default_name", source, false,
			`UPDATE entities SET default_name = $2, updated_at = now() WHERE id = $1`, id, text)
	case strings.HasPrefix(field, "external_ids."):
		return writeExternalJSON(ctx, ex, "entities", id, field, after, source)
	}

	return nil, ErrUnsupportedField
}

// writeEdition has no per-field provenance: editions carry one row-level source (SDD 01 section 5), which a single
// community field must not relabel, so the audit row is the only record of who wrote it.
func writeEdition(
	ctx context.Context,
	ex database.Executor,
	id, field string,
	after json.RawMessage,
) (json.RawMessage, error) {
	text, null, err := textValue(after)
	if err != nil {
		return nil, err
	}

	column := ""
	if field == "title" && !null {
		column = "title"
	} else if p, ok := strings.CutPrefix(field, "external_ids."); ok {
		column = editionExternal[p]
	}

	if column == "" {
		return nil, ErrUnsupportedField
	}

	if err := execOne(ctx, ex, `UPDATE editions SET `+column+` = $2, updated_at = now() WHERE id = $1`,
		id, nullable(text, null)); err != nil {
		return nil, err
	}

	return json.RawMessage("null"), nil
}

// writeExternalJSON sets or removes one key of the external_ids jsonb column of table (items or entities).
func writeExternalJSON(
	ctx context.Context, ex database.Executor, table, id, field string, after json.RawMessage, source string,
) (json.RawMessage, error) {
	provider := strings.TrimPrefix(field, "external_ids.")

	text, null, err := textValue(after)
	if err != nil {
		return nil, err
	}

	q := `UPDATE ` + table + ` SET external_ids = external_ids || jsonb_build_object($2::text, $3::text),
updated_at = now() WHERE id = $1`
	args := []any{id, provider, text}

	if null {
		q = `UPDATE ` + table + ` SET external_ids = external_ids - $2::text, updated_at = now() WHERE id = $1`
		args = args[:2]
	}

	return provWrite(ctx, ex, table, id, field, source, null, q, args...)
}

// provWrite runs the value update and then records source under key in table.provenance (removing the key when the
// value was cleared). table is a constant.
func provWrite(
	ctx context.Context, ex database.Executor, table, id, key, source string, cleared bool, update string, args ...any,
) (json.RawMessage, error) {
	var prev []byte
	if err := ex.QueryRowContext(ctx, `SELECT COALESCE(provenance -> $2, 'null') FROM `+table+` WHERE id = $1`,
		id, key).Scan(&prev); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTargetGone
		}

		return nil, fmt.Errorf("catalog: read provenance: %w", err)
	}

	if err := execOne(ctx, ex, update, args...); err != nil {
		return nil, err
	}

	q := `UPDATE ` + table + ` SET provenance = provenance || jsonb_build_object($2::text, $3::text) WHERE id = $1`
	pArgs := []any{id, key, source}

	if cleared {
		q = `UPDATE ` + table + ` SET provenance = provenance - $2::text WHERE id = $1`
		pArgs = pArgs[:2]
	}

	if err := execOne(ctx, ex, q, pArgs...); err != nil {
		return nil, err
	}

	return prev, nil
}

func writeItemTitle(
	ctx context.Context, ex database.Executor, id, locale string, after json.RawMessage, source string,
) (json.RawMessage, error) {
	text, null, err := textValue(after)
	if err != nil {
		return nil, err
	}

	if null {
		// title is NOT NULL and the row also holds the synopsis: clearing a translated title is not a field write.
		return nil, ErrUnsupportedField
	}

	var prev []byte
	if err := ex.QueryRowContext(ctx, `SELECT COALESCE((SELECT to_jsonb(title_source) FROM item_translations
WHERE item_id = $1 AND locale = $2), 'null')`, id, locale).Scan(&prev); err != nil {
		return nil, fmt.Errorf("catalog: read title source: %w", err)
	}

	if err := execOne(ctx, ex, `INSERT INTO item_translations (item_id, locale, title, title_source)
SELECT id, $2, $3, $4 FROM items WHERE id = $1
ON CONFLICT (item_id, locale) DO UPDATE SET title = EXCLUDED.title, title_source = EXCLUDED.title_source`,
		id, locale, text, source); err != nil {
		return nil, err
	}

	return prev, nil
}

func writeItemSynopsis(
	ctx context.Context, ex database.Executor, id, locale string, after json.RawMessage, source string,
) (json.RawMessage, error) {
	text, null, err := textValue(after)
	if err != nil {
		return nil, err
	}

	var prev []byte
	if err := ex.QueryRowContext(ctx, `SELECT COALESCE((SELECT to_jsonb(synopsis_source) FROM item_translations
WHERE item_id = $1 AND locale = $2), 'null')`, id, locale).Scan(&prev); err != nil {
		return nil, fmt.Errorf("catalog: read synopsis source: %w", err)
	}

	if null {
		_, err := ex.ExecContext(ctx, `UPDATE item_translations SET synopsis = NULL, synopsis_source = NULL,
synopsis_licence = NULL, synopsis_source_url = NULL WHERE item_id = $1 AND locale = $2`, id, locale)
		if err != nil {
			return nil, fmt.Errorf("catalog: clear synopsis: %w", err)
		}

		return prev, nil
	}

	// A synopsis in a locale without a translation row needs a title there: the original title stands in, labelled
	// with its own provenance (community when it has none).
	if err := execOne(ctx, ex, `INSERT INTO item_translations
(item_id, locale, title, title_source, synopsis, synopsis_source, synopsis_licence)
SELECT id, $2, COALESCE(original_title, slug), COALESCE(provenance ->> 'original_title', $4), $3, $4, 'proprietary'
FROM items WHERE id = $1
ON CONFLICT (item_id, locale) DO UPDATE SET synopsis = EXCLUDED.synopsis, synopsis_source = EXCLUDED.synopsis_source,
  synopsis_licence = EXCLUDED.synopsis_licence, synopsis_source_url = NULL`, id, locale, text, source); err != nil {
		return nil, err
	}

	return prev, nil
}

// RelationRef is one entity on an item relation.
type RelationRef struct {
	ID   string `json:"ID"`
	Role string `json:"Role"`
}

// relationRoles maps a suggestion relation field to the item_entities roles it covers; a single-role field also
// fixes the role of its ops.
var relationRoles = map[string][]string{
	"genres":        {"genre"},
	"platforms":     {"platform"},
	"franchise":     {"franchise"},
	"creators":      {"director", "writer", "cast", "creator", "author"},
	"organizations": {"developer", "publisher", "production_company", "network"},
}

// WriteRelation applies remove then add ops of one relation field to an item and returns the field's entity list
// before and after, as JSON arrays sorted by role and id.
func WriteRelation(
	ctx context.Context, ex database.Executor, itemID, field string, add, remove []RelationRef, source string,
) (json.RawMessage, json.RawMessage, error) {
	roles, ok := relationRoles[field]
	if !ok {
		return nil, nil, ErrUnsupportedField
	}

	before, err := relationList(ctx, ex, itemID, roles)
	if err != nil {
		return nil, nil, err
	}

	role := func(r RelationRef) (string, error) {
		if len(roles) == 1 {
			return roles[0], nil
		}

		if !slices.Contains(roles, r.Role) {
			return "", ErrUnsupportedField
		}

		return r.Role, nil
	}

	for _, r := range remove {
		rl, err := role(r)
		if err != nil {
			return nil, nil, err
		}

		if _, err := ex.ExecContext(
			ctx,
			`DELETE FROM item_entities WHERE item_id = $1 AND entity_id = $2 AND role = $3`,
			itemID,
			r.ID,
			rl,
		); err != nil {
			return nil, nil, fmt.Errorf("catalog: remove relation: %w", err)
		}
	}

	for _, r := range add {
		rl, err := role(r)
		if err != nil {
			return nil, nil, err
		}

		if _, err := ex.ExecContext(ctx, `INSERT INTO item_entities (item_id, entity_id, role, source)
VALUES ($1, $2, $3, $4) ON CONFLICT (item_id, role, entity_id) DO NOTHING`, itemID, r.ID, rl, source); err != nil {
			return nil, nil, fmt.Errorf("catalog: add relation: %w", err)
		}
	}

	after, err := relationList(ctx, ex, itemID, roles)
	if err != nil {
		return nil, nil, err
	}

	return before, after, nil
}

func relationList(ctx context.Context, ex database.Executor, itemID string, roles []string) (json.RawMessage, error) {
	var out []byte

	err := ex.QueryRowContext(ctx, `SELECT COALESCE(jsonb_agg(jsonb_build_object('ID', entity_id::text, 'Role', role)
ORDER BY role, entity_id), '[]') FROM item_entities WHERE item_id = $1 AND role = ANY($2::text[])`,
		itemID, "{"+strings.Join(roles, ",")+"}").Scan(&out)
	if err != nil {
		return nil, fmt.Errorf("catalog: read relations: %w", err)
	}

	return out, nil
}

func execOne(ctx context.Context, ex database.Executor, q string, args ...any) error {
	res, err := ex.ExecContext(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("catalog: write field: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("catalog: write field: %w", err)
	}

	if n == 0 {
		return ErrTargetGone
	}

	return nil
}

func textValue(raw json.RawMessage) (string, bool, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return "", true, nil
	}

	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return "", false, fmt.Errorf("catalog: field value is not a string: %w", err)
	}

	return s, false, nil
}

func nullable(s string, null bool) any {
	if null {
		return nil
	}

	return s
}
