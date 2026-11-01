package jobs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/community"
	"github.com/AlexGalhardo/hexascore-app/app/notifications"
)

// noticeKinds maps the suggestion states that notify the author to the inbox kind (SDD 04 FR-04-14).
var noticeKinds = map[string]string{
	"approved": notifications.SuggestionDecided,
	"applied":  notifications.SuggestionDecided,
	"rejected": notifications.SuggestionDecided,
	"reverted": notifications.SuggestionReverted,
	"expired":  notifications.SuggestionExpired,
}

// handleCommunityNotify writes the author's inbox row and queues the mail (SDD 04 section 6 jobs).
func handleCommunityNotify(app *vukem.App) queue.Registration {
	get := sync.OnceValues(func() (*accounts.Accounts, error) { return accounts.New(app) })

	return queue.Handle(func(ctx context.Context, j community.Notify) error {
		acc, err := get()
		if err != nil {
			return err
		}

		return RunCommunityNotify(ctx, acc.DB, acc, j)
	}, queue.Retries(3))
}

// RunCommunityNotify sends the notice of j in one transaction: the inbox row keyed suggestion:<id>:<state>, and the
// mail only when that row is new, so a retried job writes one row and sends one mail. A suggestion without an
// author (deleted account), an anonymized author or a state without a notice is a no-op; an author who opted out
// of community mail gets the inbox row only.
func RunCommunityNotify(ctx context.Context, db *database.DB, acc *accounts.Accounts, j community.Notify) error {
	kind, ok := noticeKinds[j.State]
	if !ok {
		return nil
	}

	return db.Tx(ctx, func(tx *database.Tx) error {
		var (
			author      sql.NullInt64
			optOut      bool
			reason      sql.NullString
			targetType  string
			targetID    sql.NullString
			suggestKind string
		)

		err := tx.QueryRowContext(ctx, `SELECT u.id, COALESCE(u.community_mail_opt_out, true), s.reject_reason,
  s.kind::text, s.target_type::text, COALESCE(s.target_id::text, s.meta->>'created_id')
FROM suggestions s LEFT JOIN users u ON u.id = s.author_id AND u.anonymized_at IS NULL
WHERE s.id = $1`, j.SuggestionID).Scan(&author, &optOut, &reason, &suggestKind, &targetType, &targetID)
		if errors.Is(err, sql.ErrNoRows) || (err == nil && !author.Valid) {
			return nil
		}

		if err != nil {
			return fmt.Errorf("community.notify: load %s: %w", j.SuggestionID, err)
		}

		payload := map[string]any{
			"suggestionId": j.SuggestionID, "state": j.State, "kind": suggestKind,
			"targetType": targetType, "targetId": targetID.String,
		}
		if reason.Valid {
			payload["reason"] = reason.String
		}

		created, err := notifications.Send(ctx, tx, author.Int64, kind,
			"suggestion:"+j.SuggestionID+":"+j.State, payload)
		if err != nil || !created || optOut {
			return err
		}

		return acc.SendSuggestionNoticeTx(ctx, tx, author.Int64, j.SuggestionID, j.State)
	})
}
