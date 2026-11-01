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
func (g IngestGate) CommunityFields(ctx context.Context, itemID string) (map[string]normalize.CommunityField, error) {
	out := map[string]normalize.CommunityField{}

	err := g.DB.Tx(ctx, func(tx *database.Tx) error {
		rows, err := tx.QueryContext(ctx, `
SELECT a.field, a.occurred_at FROM catalog_audit a
WHERE a.target_type = 'item' AND a.target_id = $1 AND a.source = 'community'
  AND a.id = (SELECT max(b.id) FROM catalog_audit b
              WHERE b.target_type = a.target_type AND b.target_id = a.target_id AND b.field = a.field)`, itemID)
		if err != nil {
			return fmt.Errorf("community: read community fields: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var field string

			var cf normalize.CommunityField
			if err := rows.Scan(&field, &cf.Since); err != nil {
				return fmt.Errorf("community: read community fields: %w", err)
			}

			out[field] = cf
		}

		if err := rows.Err(); err != nil {
			return fmt.Errorf("community: read community fields: %w", err)
		}

		for field, cf := range out {
			cur, err := scalarCurrent(ctx, tx, "item", itemID, field)
			if err != nil {
				return err
			}

			cf.Current = cur
			out[field] = cf
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

// OpenSystemSuggestion files a submitted `system` suggestion (no author) for
// one field, through the normal draft and submit transitions. A suggestion
// with the same target, field and value that is still open is not repeated,
// so a re-run does not flood the queue.
func (g IngestGate) OpenSystemSuggestion(ctx context.Context, s normalize.SystemSuggestion) error {
	changes, err := json.Marshal(map[string]map[string]json.RawMessage{
		s.Field: {"before": s.Before, "after": s.After},
	})
	if err != nil {
		return fmt.Errorf("community: encode system suggestion: %w", err)
	}

	locale := "en"
	if rest, ok := strings.CutPrefix(s.Field, "translations."); ok {
		locale, _, _ = strings.Cut(rest, ".")
	} else if rest, ok := strings.CutPrefix(s.Field, "synopsis."); ok {
		locale = rest
	}

	system := Actor{Role: "system"}

	return g.DB.Tx(ctx, func(tx *database.Tx) error {
		if _, err := tx.ExecContext(
			ctx,
			`SELECT pg_advisory_xact_lock(hashtext('item'), hashtext($1))`,
			s.ItemID,
		); err != nil {
			return fmt.Errorf("community: lock suggestion target: %w", err)
		}

		var open int
		if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM suggestions
WHERE author_id IS NULL AND target_type = 'item' AND target_id = $1 AND status IN ('submitted','in_review')
  AND changes -> $2 -> 'after' = $3::jsonb`, s.ItemID, s.Field, string(s.After)).Scan(&open); err != nil {
			return fmt.Errorf("community: scan open system suggestions: %w", err)
		}

		if open > 0 {
			return nil
		}

		id := uuid.NewV7().String()
		meta, _ := json.Marshal(map[string]any{"system": true, "source": s.Source})

		if _, err := tx.ExecContext(ctx, `INSERT INTO suggestions
(id, author_id, kind, target_type, target_id, target_domain, locale, changes, note, meta)
VALUES ($1, NULL, 'edit', 'item', $2, $3, $4, $5::jsonb, $6, $7::jsonb)`,
			id, s.ItemID, s.Domain, locale, changes,
			"Source "+s.Source+" differs from the approved community value.", meta); err != nil {
			return fmt.Errorf("community: insert system suggestion: %w", err)
		}

		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO suggestion_events (suggestion_id, from_status, to_status, actor_role)
VALUES ($1, NULL, 'draft', 'system')`,
			id,
		); err != nil {
			return fmt.Errorf("community: insert system suggestion event: %w", err)
		}

		if err := TransitionTx(ctx, tx, id, 1, "submitted", system); err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, `UPDATE suggestions SET submitted_at = now() WHERE id = $1`, id); err != nil {
			return fmt.Errorf("community: set submitted_at: %w", err)
		}

		return nil
	})
}
