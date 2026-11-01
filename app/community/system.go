package community

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/uuid"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/normalize"
)

// IngestGate is what ingestion calls (SDD 06 FR-06-4, SDD 04 §7.3): it reports
// the community-owned fields of an item and opens `system` suggestions when a
// source value diverges. It never applies anything; only an admin applies a
// suggestion, after moderators recommend.
type IngestGate struct {
	DB *database.DB
}

var _ normalize.Community = IngestGate{}

// CommunityFields returns the item fields whose latest audit row is an
// applied community edit, with that row's time and the live value.
func (g IngestGate) CommunityFields(
	ctx context.Context,
	tx *database.Tx,
	itemID string,
) (map[string]normalize.CommunityField, error) {
	out := map[string]normalize.CommunityField{}

	rows, err := tx.QueryContext(ctx, `
SELECT a.field, a.occurred_at FROM catalog_audit a
WHERE a.target_type = 'item' AND a.target_id = $1 AND a.source = 'community'
  AND a.id = (SELECT max(b.id) FROM catalog_audit b
              WHERE b.target_type = a.target_type AND b.target_id = a.target_id AND b.field = a.field)`, itemID)
	if err != nil {
		return nil, fmt.Errorf("community: read community fields: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var field string

		var cf normalize.CommunityField
		if err := rows.Scan(&field, &cf.Since); err != nil {
			return nil, fmt.Errorf("community: read community fields: %w", err)
		}

		out[field] = cf
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("community: read community fields: %w", err)
	}

	for field, cf := range out {
		cur, err := scalarCurrent(ctx, tx, "item", itemID, field)
		if err != nil {
			return nil, err
		}

		cf.Current = cur
		out[field] = cf
	}

	return out, nil
}

// OpenSystemSuggestion files a submitted `system` suggestion (no author) for
// one field, through the normal draft and submit transitions. A suggestion
// with the same target, field and value that is still open, or that a
// moderator already rejected or let expire, is not repeated, so a re-run does
// not flood the queue or reopen a decided one. It runs on the caller's tx.
func (g IngestGate) OpenSystemSuggestion(ctx context.Context, tx *database.Tx, s normalize.SystemSuggestion) error {
	_, err := g.open(ctx, tx, s, "Source "+s.Source+" differs from the approved community value.",
		map[string]any{"system": true, "source": s.Source})

	return err
}

// MissingFieldSource is the meta source of the suggestions the missing-field detector files (SDD 06 §7.7).
const MissingFieldSource = "dq.frontend_fields"

// OpenMissingField files a system suggestion asking the community to fill a required field of an item page. The
// proposed value is null: a moderator rejects it or turns it into a real edit. The dedup of OpenSystemSuggestion
// applies (an open, rejected or expired entry for the field is not repeated); the result says whether a new row was
// filed.
func (g IngestGate) OpenMissingField(ctx context.Context, tx *database.Tx, itemID, domain, field string) (bool, error) {
	return g.open(ctx, tx, normalize.SystemSuggestion{
		ItemID: itemID, Domain: domain, Field: field, Before: json.RawMessage("null"), After: json.RawMessage("null"),
		Source: MissingFieldSource,
	}, "The "+field+" field is missing on the item page.",
		map[string]any{"system": true, "source": MissingFieldSource, "missing_field": field})
}

func (g IngestGate) open(
	ctx context.Context, tx *database.Tx, s normalize.SystemSuggestion, note string, metaFields map[string]any,
) (bool, error) {
	changes, err := json.Marshal(map[string]map[string]json.RawMessage{
		s.Field: {"before": s.Before, "after": s.After},
	})
	if err != nil {
		return false, fmt.Errorf("community: encode system suggestion: %w", err)
	}

	locale := "en"
	if rest, ok := strings.CutPrefix(s.Field, "translations."); ok {
		locale, _, _ = strings.Cut(rest, ".")
	} else if rest, ok := strings.CutPrefix(s.Field, "synopsis."); ok {
		locale = rest
	}

	system := Actor{Role: "system"}

	if _, err := tx.ExecContext(
		ctx,
		`SELECT pg_advisory_xact_lock(hashtext('item'), hashtext($1))`,
		s.ItemID,
	); err != nil {
		return false, fmt.Errorf("community: lock suggestion target: %w", err)
	}

	var open int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM suggestions
WHERE author_id IS NULL AND target_type = 'item' AND target_id = $1 AND status IN ('submitted','in_review','rejected','expired')
  AND changes -> $2 -> 'after' = $3::jsonb`, s.ItemID, s.Field, string(s.After)).Scan(&open); err != nil {
		return false, fmt.Errorf("community: scan open system suggestions: %w", err)
	}

	if open > 0 {
		return false, nil
	}

	id := uuid.NewV7().String()
	meta, _ := json.Marshal(metaFields)

	if _, err := tx.ExecContext(ctx, `INSERT INTO suggestions
(id, author_id, kind, target_type, target_id, target_domain, locale, changes, note, meta)
VALUES ($1, NULL, 'edit', 'item', $2, $3, $4, $5::jsonb, $6, $7::jsonb)`,
		id, s.ItemID, s.Domain, locale, changes,
		note, meta); err != nil {
		return false, fmt.Errorf("community: insert system suggestion: %w", err)
	}

	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO suggestion_events (suggestion_id, from_status, to_status, actor_role)
VALUES ($1, NULL, 'draft', 'system')`,
		id,
	); err != nil {
		return false, fmt.Errorf("community: insert system suggestion event: %w", err)
	}

	if err := TransitionTx(ctx, tx, id, 1, "submitted", system); err != nil {
		return false, err
	}

	if _, err := tx.ExecContext(ctx, `UPDATE suggestions SET submitted_at = now() WHERE id = $1`, id); err != nil {
		return false, fmt.Errorf("community: set submitted_at: %w", err)
	}

	return true, nil
}
