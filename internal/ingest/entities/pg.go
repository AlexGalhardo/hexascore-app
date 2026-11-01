package entities

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/uuid"

	"github.com/AlexGalhardo/hexascore-app/internal/search"
)

// PG is the Store over the entities, item_entities and entity_match_reviews tables.
type PG struct {
	Ex database.Executor
}

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

// maxSlug keeps the slug under the entities CHECK (120) with room for a suffix.
const maxSlug = 80

// ByQID finds the entity holding qid, whatever its kind.
func (p PG) ByQID(ctx context.Context, qid string) (Entity, bool, error) {
	var e Entity

	err := p.Ex.QueryRowContext(ctx, `SELECT id::text, kind FROM entities WHERE wikidata_qid = $1`, qid).
		Scan(&e.ID, &e.Kind)
	if err == sql.ErrNoRows {
		return Entity{}, false, nil
	}

	return e, err == nil, err
}

// ByExternalID finds the entities of kind whose external_ids hold key = id.
func (p PG) ByExternalID(ctx context.Context, key, id, kind string) ([]Entity, error) {
	return p.list(
		ctx,
		`SELECT id::text, kind, 0 FROM entities WHERE kind = $1 AND external_ids ->> $2 = $3 ORDER BY id`,
		kind,
		key,
		id,
	)
}

// ByName compares the normalised default_name, the expression the entities_name_norm_idx index covers.
func (p PG) ByName(ctx context.Context, normalized, kind string) ([]Entity, error) {
	return p.list(ctx, `SELECT id::text, kind, coalesce(extract(year FROM born_on)::int, founded_year, 0)
FROM entities WHERE kind = $1 AND `+search.EntityNameMatch("$2")+` ORDER BY id`, kind, normalized)
}

func (p PG) list(ctx context.Context, q string, args ...any) ([]Entity, error) {
	rows, err := p.Ex.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("entities: lookup: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []Entity

	for rows.Next() {
		var e Entity
		if err := rows.Scan(&e.ID, &e.Kind, &e.Year); err != nil {
			return nil, fmt.Errorf("entities: lookup: %w", err)
		}

		out = append(out, e)
	}

	return out, rows.Err()
}

// Create inserts a native entity. The slug is the name, then the name with the QID, then with the id, so a clash
// on (kind, slug) never loses the credit.
func (p PG) Create(ctx context.Context, c Credit) (string, error) {
	name := strings.TrimSpace(c.Name)
	if name == "" {
		name = c.QID + c.ExternalID
	}

	ext := "{}"
	if c.ExternalIDKey != "" && c.ExternalID != "" {
		b, _ := json.Marshal(map[string]string{c.ExternalIDKey: c.ExternalID})
		ext = string(b)
	}

	id := uuid.NewV7().String()
	base := slug(name)

	candidates := []string{base}
	if c.QID != "" {
		candidates = append(candidates, base+"-"+strings.ToLower(c.QID))
	}

	candidates = append(candidates, base+"-"+id[len(id)-12:])

	for _, s := range candidates {
		var got string

		err := p.Ex.QueryRowContext(ctx, `INSERT INTO entities (id, kind, slug, wikidata_qid, default_name, external_ids)
VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6::jsonb) ON CONFLICT (kind, slug) DO NOTHING RETURNING id::text`,
			id, c.Kind, s, c.QID, name, ext).
			Scan(&got)
		if err == sql.ErrNoRows {
			continue
		}

		if err != nil {
			return "", fmt.Errorf("entities: create %s: %w", name, err)
		}

		return got, nil
	}

	return "", fmt.Errorf("entities: no free slug for %q", name)
}

func slug(name string) string {
	s := strings.Trim(nonAlnum.ReplaceAllString(NormalizeName(name), "-"), "-")
	if s == "" {
		s = "entity"
	}

	if len(s) > maxSlug {
		s = strings.TrimRight(s[:maxSlug], "-")
	}

	return s
}

// Link upserts the item_entities row and teaches the entity the QID or external id the credit brought.
func (p PG) Link(ctx context.Context, entityID string, c Credit) error {
	if _, err := p.Ex.ExecContext(
		ctx,
		`INSERT INTO item_entities (item_id, entity_id, role, position, character_name, source)
VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6)
ON CONFLICT (item_id, role, entity_id) DO UPDATE SET position = EXCLUDED.position, character_name = EXCLUDED.character_name`,
		c.ItemID,
		entityID,
		c.Role,
		c.Position,
		c.Character,
		c.Source,
	); err != nil {
		return fmt.Errorf("entities: link %s: %w", entityID, err)
	}

	if c.QID != "" {
		if _, err := p.Ex.ExecContext(ctx, `UPDATE entities SET wikidata_qid = $2, updated_at = now()
WHERE id = $1 AND wikidata_qid IS NULL AND NOT EXISTS (SELECT 1 FROM entities WHERE wikidata_qid = $2)`, entityID, c.QID); err != nil {
			return fmt.Errorf("entities: set qid: %w", err)
		}
	}

	if c.ExternalIDKey != "" && c.ExternalID != "" {
		if _, err := p.Ex.ExecContext(
			ctx,
			`UPDATE entities SET external_ids = jsonb_build_object($2::text, $3::text) || external_ids,
updated_at = now() WHERE id = $1 AND NOT jsonb_exists(external_ids, $2)`,
			entityID,
			c.ExternalIDKey,
			c.ExternalID,
		); err != nil {
			return fmt.Errorf("entities: set external id: %w", err)
		}
	}

	return nil
}

// OpenReview queues the credit once: the unique key (item, role, source, raw_name) makes a rerun a no-op.
func (p PG) OpenReview(ctx context.Context, c Credit, candidates []string) error {
	raw := c.Name
	if strings.TrimSpace(raw) == "" {
		raw = c.QID + c.ExternalID
	}

	extID := c.QID
	if c.ExternalID != "" {
		extID = c.ExternalID
	}

	arr, _ := json.Marshal(candidates)

	_, err := p.Ex.ExecContext(
		ctx,
		`INSERT INTO entity_match_reviews (item_id, role, source, raw_name, raw_year, raw_external_id, candidates)
VALUES ($1, $2, $3, $4, NULLIF($5, 0), NULLIF($6, ''),
  ARRAY(SELECT jsonb_array_elements_text($7::jsonb)::uuid))
ON CONFLICT (item_id, role, source, raw_name) DO NOTHING`,
		c.ItemID,
		c.Role,
		c.Source,
		raw,
		c.Year,
		extID,
		string(arr),
	)
	if err != nil {
		return fmt.Errorf("entities: open review: %w", err)
	}

	return nil
}
