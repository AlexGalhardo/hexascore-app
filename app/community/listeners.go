package community

import (
	"context"
	"fmt"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/queue"
)

// noSuggestion stands in for "no suggestion to spare" in expireOn.
const noSuggestion = "00000000-0000-0000-0000-000000000000"

// WithdrawAuthor moves userID's open suggestions (draft, submitted, in_review) to withdrawn with one system event
// each carrying reason (SDD 04 section 8: author banned or deleted). Approved suggestions stay: an admin already
// decided them. A replay finds nothing left to withdraw.
func WithdrawAuthor(ctx context.Context, ex database.Executor, userID int64, reason string) error {
	if _, err := ex.ExecContext(ctx, `WITH o AS (
  SELECT id, status FROM suggestions
  WHERE author_id = $1 AND status IN ('draft','submitted','in_review') FOR UPDATE
), u AS (
  UPDATE suggestions s SET status = 'withdrawn', version = s.version + 1, updated_at = now(),
    assigned_to = NULL, claimed_at = NULL
  FROM o WHERE s.id = o.id RETURNING s.id
)
INSERT INTO suggestion_events (suggestion_id, from_status, to_status, actor_role, reason)
SELECT o.id, o.status, 'withdrawn', 'system', $2 FROM o JOIN u ON u.id = o.id`, userID, reason); err != nil {
		return fmt.Errorf("community: withdraw suggestions of %d: %w", userID, err)
	}

	return nil
}

// ForgetUser runs inside the account anonymisation transaction (SDD 03): it withdraws the user's open suggestions
// and nulls every community reference to them (NFR-04-5), so public pages show "deleted user". catalog_audit is
// append-only and keeps its actor_id, which now points at the anonymized row.
func ForgetUser(ctx context.Context, ex database.Executor, userID int64) error {
	if err := WithdrawAuthor(ctx, ex, userID, "author_deleted"); err != nil {
		return err
	}

	for _, q := range []string{
		`UPDATE suggestions SET author_id = NULL WHERE author_id = $1`,
		`UPDATE suggestions SET assigned_to = NULL, claimed_at = NULL WHERE assigned_to = $1`,
		`UPDATE suggestions SET decided_by = NULL WHERE decided_by = $1`,
		`UPDATE suggestion_events SET actor_id = NULL WHERE actor_id = $1`,
		`DELETE FROM suggestion_reviews WHERE reviewer_id = $1`,
		`UPDATE catalog_locks SET locked_by = NULL WHERE locked_by = $1`,
	} {
		if _, err := ex.ExecContext(ctx, q, userID); err != nil {
			return fmt.Errorf("community: forget user %d: %w", userID, err)
		}
	}

	return nil
}

// ExpireTargetGone expires the open suggestions on a deleted or merged-away target (reason target_gone, no stats
// impact) and queues each author's notice, in one transaction. The community merge apply does this in its own
// transaction; this covers merges from the admin catalogue and the console.
func ExpireTargetGone(ctx context.Context, db *database.DB, q *queue.Queue, targetID string) error {
	return db.Tx(ctx, func(tx *database.Tx) error {
		return Applier{Queue: q}.expireOn(ctx, tx, targetID, noSuggestion)
	})
}
