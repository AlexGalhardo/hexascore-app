package resolve

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/AlexGalhardo/vukem/database"
)

var (
	// ErrNotFound means no conflict has that id.
	ErrNotFound = errors.New("resolve: conflict not found")
	// ErrResolved means the conflict was already decided; the decision is final.
	ErrResolved = errors.New("resolve: conflict already resolved")
)

// OpenConflict is a conflict waiting for an admin, with the item it collides with.
type OpenConflict struct {
	ID                             int64
	QID, Type, Kind, Key, Value    string
	OtherID, OtherSlug, OtherTitle string
	Similarity                     float64
	CreatedAt                      time.Time
}

// Open lists the oldest open conflicts, at most limit.
func Open(ctx context.Context, ex database.Executor, limit int) ([]OpenConflict, error) {
	rows, err := ex.QueryContext(
		ctx,
		`SELECT c.id, c.wikidata_qid, c.item_type, c.kind, c.key, c.value, c.other_item_id::text,
  i.slug, coalesce(t.title, i.slug), coalesce(c.similarity, 0), c.created_at
FROM item_conflicts c JOIN items i ON i.id = c.other_item_id
LEFT JOIN item_translations t ON t.item_id = i.id AND t.locale = 'en'
WHERE c.status = 'open' ORDER BY c.created_at, c.id LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("resolve: list open conflicts: %w", err)
	}

	defer func() { _ = rows.Close() }()

	var out []OpenConflict

	for rows.Next() {
		var c OpenConflict
		if err := rows.Scan(&c.ID, &c.QID, &c.Type, &c.Kind, &c.Key, &c.Value, &c.OtherID, &c.OtherSlug, &c.OtherTitle,
			&c.Similarity, &c.CreatedAt); err != nil {
			return nil, err
		}

		out = append(out, c)
	}

	return out, rows.Err()
}

// Decide closes an open conflict as allowed or dismissed for the admin. Allowing puts the QID's Wikidata record back
// in the normaliser's pending set, so the item is created on the next run (without the contested id for an
// external-id conflict); dismissing keeps it out for good.
func Decide(ctx context.Context, db *database.DB, id int64, state string, admin int64) error {
	if state != StateAllowed && state != StateDismissed {
		return fmt.Errorf("resolve: state %q cannot close a conflict", state)
	}

	return db.Tx(ctx, func(tx *database.Tx) error {
		var qid string

		err := tx.QueryRowContext(ctx, `UPDATE item_conflicts SET status = $2, resolved_at = now(), resolved_by = $3
WHERE id = $1 AND status = 'open' RETURNING wikidata_qid`, id, state, admin).Scan(&qid)
		if errors.Is(err, sql.ErrNoRows) {
			var exists bool
			if qerr := tx.QueryRowContext(ctx, `SELECT true FROM item_conflicts WHERE id = $1`, id).
				Scan(&exists); qerr != nil {
				return ErrNotFound
			}

			return ErrResolved
		}

		if err != nil {
			return fmt.Errorf("resolve: decide conflict %d: %w", id, err)
		}

		if state == StateAllowed {
			if _, err := tx.ExecContext(ctx, `UPDATE source_records SET normalized_at = NULL
WHERE source = 'wikidata' AND external_id = $1`, qid); err != nil {
				return fmt.Errorf("resolve: reopen %s: %w", qid, err)
			}
		}

		return nil
	})
}
