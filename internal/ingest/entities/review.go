package entities

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/AlexGalhardo/vukem/database"
)

// Review actions of the match queue (SDD 06 §6).
const (
	ActionLink    = "link"
	ActionCreate  = "create"
	ActionDismiss = "dismiss"
)

var (
	// ErrReviewNotFound means no review has that id.
	ErrReviewNotFound = errors.New("entities: review not found")
	// ErrReviewResolved means the review was already closed; resolution is final.
	ErrReviewResolved = errors.New("entities: review already resolved")
	// ErrInvalidReview means the action or its entity does not fit the review.
	ErrInvalidReview = errors.New("entities: invalid review resolution")
)

var qidRe = regexp.MustCompile(`^Q[1-9][0-9]*$`)

// ReviewCandidate is one entity a review offers.
type ReviewCandidate struct {
	ID, Name, Kind, QID string
}

// OpenReviewRow is a review waiting for a decision, with the item it belongs to.
type OpenReviewRow struct {
	ID                                   int64
	ItemID, ItemType, ItemTitle          string
	Role, Source, RawName, RawExternalID string
	RawYear                              int
	Candidates                           []ReviewCandidate
}

// ReviewQueue resolves entity_match_reviews. KindOf maps a credit role to the entity kind it holds.
type ReviewQueue struct {
	DB     *database.DB
	KindOf func(role string) (string, bool)
}

// Open lists the oldest open reviews, at most limit.
func (q ReviewQueue) Open(ctx context.Context, limit int) ([]OpenReviewRow, error) {
	rows, err := q.DB.QueryContext(ctx, `SELECT r.id, r.item_id::text, i.type, i.original_title, r.role, r.source,
  r.raw_name, coalesce(r.raw_year, 0), coalesce(r.raw_external_id, ''),
  coalesce((SELECT json_agg(json_build_object('id', e.id, 'name', e.default_name, 'kind', e.kind,
    'qid', coalesce(e.wikidata_qid, '')) ORDER BY e.id) FROM entities e WHERE e.id = ANY (r.candidates)), '[]')::text
FROM entity_match_reviews r JOIN items i ON i.id = r.item_id
WHERE r.status = 'open' ORDER BY r.created_at, r.id LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("entities: list reviews: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []OpenReviewRow

	for rows.Next() {
		var (
			r    OpenReviewRow
			cand string
		)
		if err := rows.Scan(&r.ID, &r.ItemID, &r.ItemType, &r.ItemTitle, &r.Role, &r.Source, &r.RawName,
			&r.RawYear, &r.RawExternalID, &cand); err != nil {
			return nil, fmt.Errorf("entities: list reviews: %w", err)
		}

		if r.Candidates, err = decodeCandidates(cand); err != nil {
			return nil, err
		}

		out = append(out, r)
	}

	return out, rows.Err()
}

// Resolve applies action to the open review id. link needs an existing entity of the role's kind; create makes a
// native entity from the raw credit and links it; dismiss only closes the review. All three record who and when.
func (q ReviewQueue) Resolve(ctx context.Context, id int64, action, entityID string, by int64) error {
	switch action {
	case ActionLink, ActionCreate, ActionDismiss:
	default:
		return fmt.Errorf("%w: unknown action %q", ErrInvalidReview, action)
	}

	return q.DB.Tx(ctx, func(tx *database.Tx) error {
		var (
			c      Credit
			status string
			year   int
			ext    string
		)

		err := tx.QueryRowContext(ctx, `SELECT item_id::text, role, source, raw_name, coalesce(raw_year, 0),
coalesce(raw_external_id, ''), status FROM entity_match_reviews WHERE id = $1 FOR UPDATE`, id).
			Scan(&c.ItemID, &c.Role, &c.Source, &c.Name, &year, &ext, &status)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrReviewNotFound
		}

		if err != nil {
			return fmt.Errorf("entities: load review %d: %w", id, err)
		}

		if status != "open" {
			return ErrReviewResolved
		}

		c.Year = year
		if qidRe.MatchString(ext) {
			c.QID = ext
		}

		final := "dismissed"
		store := PG{Ex: tx}

		switch action {
		case ActionLink:
			final = "linked"

			if err := q.linkTarget(ctx, tx, c, entityID); err != nil {
				return err
			}

			if err := store.Link(ctx, entityID, c); err != nil {
				return err
			}
		case ActionCreate:
			final = "created"

			if err := q.createNative(ctx, store, &c); err != nil {
				return err
			}
		}

		if _, err := tx.ExecContext(
			ctx,
			`UPDATE entity_match_reviews SET status = $2, resolved_by = $3, resolved_at = now()
WHERE id = $1`,
			id,
			final,
			by,
		); err != nil {
			return fmt.Errorf("entities: close review %d: %w", id, err)
		}

		return nil
	})
}

func (q ReviewQueue) linkTarget(ctx context.Context, tx *database.Tx, c Credit, entityID string) error {
	kind, ok := q.KindOf(c.Role)
	if !ok || entityID == "" {
		return fmt.Errorf("%w: link needs an entity of a known role", ErrInvalidReview)
	}

	var got string

	err := tx.QueryRowContext(ctx, `SELECT kind FROM entities WHERE id::text = $1`, strings.ToLower(entityID)).
		Scan(&got)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: entity %s does not exist", ErrInvalidReview, entityID)
	}

	if err != nil {
		return fmt.Errorf("entities: load entity %s: %w", entityID, err)
	}

	if got != kind {
		return fmt.Errorf("%w: role %s holds %s entities, not %s", ErrInvalidReview, c.Role, kind, got)
	}

	return nil
}

func (q ReviewQueue) createNative(ctx context.Context, store PG, c *Credit) error {
	kind, ok := q.KindOf(c.Role)
	if !ok {
		return fmt.Errorf("%w: unknown role %s", ErrInvalidReview, c.Role)
	}

	c.Kind = kind

	// A QID held elsewhere is the reason this review may exist; creating with it would break the unique key.
	if c.QID != "" {
		if _, held, err := store.ByQID(ctx, c.QID); err != nil {
			return err
		} else if held {
			c.QID = ""
		}
	}

	id, err := store.Create(ctx, *c)
	if err != nil {
		return err
	}

	return store.Link(ctx, id, *c)
}

func decodeCandidates(raw string) ([]ReviewCandidate, error) {
	var in []struct{ ID, Name, Kind, QID string }
	if err := json.Unmarshal([]byte(raw), &in); err != nil {
		return nil, fmt.Errorf("entities: decode candidates: %w", err)
	}

	out := make([]ReviewCandidate, len(in))
	for i, c := range in {
		out[i] = ReviewCandidate(c)
	}

	return out, nil
}
