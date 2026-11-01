package social

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/uuid"

	"github.com/AlexGalhardo/hexascore-app/app/notifications"
)

// Notification kinds of this file (SDD 14 §7.5, §12.11).
const (
	KindNewComment  = "new_comment"
	KindNewReaction = "new_reaction"
)

// targetColumn maps a route's target type to its foreign-key column; it is the whitelist for the SQL below.
func targetColumn(typ string) (string, bool) {
	switch typ {
	case "review", "list", "comment":
		return typ + "_id", true
	default:
		return "", false
	}
}

// commentCap bounds one thread load; shortcut: no pagination, add a cursor when a target passes it.
const commentCap = 500

// ErrReplyDepth answers a reply to a reply: threads are one level deep.
func ErrReplyDepth() error {
	return errs.NewKey(http.StatusUnprocessableEntity, "social.comment.reply_depth", "Reply to the first comment.", nil).
		WithExtension("code", "reply-depth")
}

// ErrReactSelf answers a reaction on your own review, comment or list.
func ErrReactSelf() error {
	return errs.NewKey(http.StatusUnprocessableEntity, "social.reaction.self", "You cannot react to your own post.", nil).
		WithExtension("code", "cannot-react-to-self")
}

// Comment is one comment with its replies (top level only), as plain text.
type Comment struct {
	ID         string    `json:"id"`
	Author     Actor     `json:"author"`
	Body       string    `json:"body"`
	Likes      int       `json:"likes"`
	Helpful    int       `json:"helpful"`
	MyReaction string    `json:"myReaction,omitempty"`
	Mine       bool      `json:"mine,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
	Replies    []Comment `json:"replies,omitempty"`
}

// Reactions are the counters of a target and the viewer's own reaction ("" for none).
type Reactions struct {
	Likes   int    `json:"likes"`
	Helpful int    `json:"helpful"`
	Mine    string `json:"mine"`
}

// owner returns the author of target typ/id when viewerID may see it: their own, or a public, active author
// (§12.5) on a public item or a public, unhidden list. A comment is visible when its author and its target are.
func owner(ctx context.Context, ex database.Executor, viewerID int64, typ, id string) (int64, error) {
	if _, ok := targetColumn(typ); !ok {
		return 0, ErrNotFound()
	}
	if _, err := uuid.Parse(id); err != nil {
		return 0, ErrNotFound()
	}

	var (
		author       int64
		review, list sql.NullString
		q            string
	)

	switch typ {
	case "review":
		q = `SELECT v.user_id FROM reviews v
JOIN ratings r ON r.user_id = v.user_id AND r.item_id = v.item_id
JOIN items i ON i.id = v.item_id AND i.visibility = 'public'
JOIN users u ON u.id = v.user_id
WHERE v.id = $1 AND (v.user_id = $2 OR (` + visible + `))`
	case "list":
		q = `SELECT l.user_id FROM lists l JOIN users u ON u.id = l.user_id
WHERE l.id = $1 AND (l.user_id = $2 OR (l.is_public AND NOT l.hidden AND ` + visible + `))`
	default:
		q = `SELECT c.user_id, c.review_id, c.list_id FROM comments c JOIN users u ON u.id = c.user_id
WHERE c.id = $1 AND (c.user_id = $2 OR (` + visible + `))`
	}

	dest := []any{&author}
	if typ == "comment" {
		dest = append(dest, &review, &list)
	}

	err := ex.QueryRowContext(ctx, q, id, viewerID).Scan(dest...)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound()
	}
	if err != nil {
		return 0, fmt.Errorf("social: %s owner: %w", typ, err)
	}

	if typ == "comment" {
		parentType, parentID := "review", review.String
		if list.Valid {
			parentType, parentID = "list", list.String
		}
		if _, err := owner(ctx, ex, viewerID, parentType, parentID); err != nil {
			return 0, err
		}
	}

	return author, nil
}

func nickname(ctx context.Context, ex database.Executor, userID int64) (string, error) {
	var n string
	if err := ex.QueryRowContext(ctx, `SELECT coalesce(display_username, username, '') FROM users WHERE id = $1`,
		userID).Scan(&n); err != nil {
		return "", fmt.Errorf("social: nickname: %w", err)
	}

	return n, nil
}

// AddComment posts userID's comment on a review or list; parentID ("" for none) must be a top-level comment on
// the same target. The target's author, and the parent comment's author for a reply, are notified.
func (s *Follows) AddComment(ctx context.Context, userID int64, typ, targetID, body, parentID string) error {
	if typ == "comment" {
		return ErrNotFound()
	}

	return s.DB.Tx(ctx, func(tx *database.Tx) error {
		author, err := owner(ctx, tx, userID, typ, targetID)
		if err != nil {
			return err
		}

		col, _ := targetColumn(typ)
		notify := map[int64]bool{author: true}

		var parent any
		if parentID != "" {
			var (
				parentAuthor int64
				grand        sql.NullString
			)

			err := tx.QueryRowContext(ctx, `SELECT user_id, parent_id FROM comments WHERE id = $1 AND `+col+` = $2`,
				parentID, targetID).Scan(&parentAuthor, &grand)
			if errors.Is(err, sql.ErrNoRows) {
				return ErrNotFound()
			}
			if err != nil {
				return fmt.Errorf("social: comment parent: %w", err)
			}
			if grand.Valid {
				return ErrReplyDepth()
			}

			parent, notify[parentAuthor] = parentID, true
		}

		id := uuid.NewV7().String()
		if _, err := tx.ExecContext(ctx, `INSERT INTO comments (id, user_id, `+col+`, parent_id, body)
VALUES ($1, $2, $3, $4, $5)`, id, userID, targetID, parent, strings.TrimSpace(body)); err != nil {
			return fmt.Errorf("social: add comment: %w", err)
		}

		nick, err := nickname(ctx, tx, userID)
		if err != nil {
			return err
		}

		for to := range notify {
			if to == userID {
				continue
			}
			if _, err := notifications.Send(ctx, tx, to, KindNewComment, "comment:"+id,
				map[string]any{"nickname": nick, "target": typ}); err != nil {
				return err
			}
		}

		return nil
	})
}

// DeleteComment removes userID's own comment and its replies; another user's comment is 404.
func (s *Follows) DeleteComment(ctx context.Context, userID int64, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return ErrNotFound()
	}

	res, err := s.DB.ExecContext(ctx, `DELETE FROM comments WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("social: delete comment: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound()
	}

	return nil
}

// React sets userID's one reaction on typ/id (a new kind replaces the old one); the trigger of the reactions table
// moves the counters in the same transaction. The author is notified once per reactor and target.
func (s *Follows) React(ctx context.Context, userID int64, typ, id, kind string) error {
	return s.DB.Tx(ctx, func(tx *database.Tx) error {
		author, err := owner(ctx, tx, userID, typ, id)
		if err != nil {
			return err
		}
		if author == userID {
			return ErrReactSelf()
		}

		col, _ := targetColumn(typ)
		if _, err := tx.ExecContext(ctx, `INSERT INTO reactions (user_id, `+col+`, kind) VALUES ($1, $2, $3)
ON CONFLICT (user_id, `+col+`) WHERE `+col+` IS NOT NULL DO UPDATE SET kind = EXCLUDED.kind
WHERE reactions.kind <> EXCLUDED.kind`, userID, id, kind); err != nil {
			return fmt.Errorf("social: react: %w", err)
		}

		nick, err := nickname(ctx, tx, userID)
		if err != nil {
			return err
		}

		_, err = notifications.Send(ctx, tx, author, KindNewReaction,
			fmt.Sprintf("reaction:%s:%s:%d", typ, id, userID), map[string]any{"nickname": nick, "target": typ})

		return err
	})
}

// Unreact removes userID's reaction on typ/id; none is not an error.
func (s *Follows) Unreact(ctx context.Context, userID int64, typ, id string) error {
	col, ok := targetColumn(typ)
	if _, err := uuid.Parse(id); err != nil || !ok {
		return ErrNotFound()
	}

	if _, err := s.DB.ExecContext(ctx, `DELETE FROM reactions WHERE user_id = $1 AND `+col+` = $2`,
		userID, id); err != nil {
		return fmt.Errorf("social: unreact: %w", err)
	}

	return nil
}

// myReactions maps the ids of typ the viewer reacted to onto the kind.
func (s *Follows) myReactions(
	ctx context.Context,
	viewerID int64,
	typ string,
	ids []string,
) (map[string]string, error) {
	out := map[string]string{}
	if viewerID == 0 || len(ids) == 0 {
		return out, nil
	}

	col, _ := targetColumn(typ)
	rows, err := s.DB.QueryContext(ctx, `SELECT `+col+`::text, kind FROM reactions
WHERE user_id = $1 AND `+col+` = ANY($2::uuid[])`, viewerID, "{"+strings.Join(ids, ",")+"}")
	if err != nil {
		return nil, fmt.Errorf("social: my reactions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, kind string
		if err := rows.Scan(&id, &kind); err != nil {
			return nil, fmt.Errorf("social: my reaction: %w", err)
		}
		out[id] = kind
	}

	return out, rows.Err()
}

// ListReactions returns the counters of list id and the viewer's reaction.
func (s *Follows) ListReactions(ctx context.Context, viewerID int64, id string) (Reactions, error) {
	var r Reactions
	if err := s.DB.QueryRowContext(ctx, `SELECT like_count, helpful_count FROM lists WHERE id = $1`, id).
		Scan(&r.Likes, &r.Helpful); err != nil {
		return r, fmt.Errorf("social: list reactions: %w", err)
	}

	mine, err := s.myReactions(ctx, viewerID, "list", []string{id})
	r.Mine = mine[id]

	return r, err
}

// Comments loads the threads of the given reviews or lists, oldest first, keyed by target id. Comments by authors
// hidden from the viewer (§12.5) are left out, and so are the replies under them.
func (s *Follows) Comments(
	ctx context.Context,
	viewerID int64,
	typ string,
	ids []string,
) (map[string][]Comment, error) {
	out := map[string][]Comment{}
	col, ok := targetColumn(typ)
	if !ok || typ == "comment" || len(ids) == 0 {
		return out, nil
	}

	rows, err := s.DB.QueryContext(ctx, `SELECT c.id::text, c.`+col+`::text, coalesce(c.parent_id::text, ''), c.user_id,
  coalesce(u.display_username, u.username, ''), c.body, c.like_count, c.helpful_count, c.created_at,
  coalesce(rx.kind, '')
FROM comments c JOIN users u ON u.id = c.user_id
LEFT JOIN reactions rx ON rx.comment_id = c.id AND rx.user_id = $2
WHERE c.`+col+` = ANY($1::uuid[]) AND (c.user_id = $2 OR (`+visible+`))
ORDER BY c.created_at, c.id LIMIT $3`, "{"+strings.Join(ids, ",")+"}", viewerID, commentCap)
	if err != nil {
		return nil, fmt.Errorf("social: comments: %w", err)
	}
	defer rows.Close()

	type row struct {
		target, parent string
		c              Comment
	}

	var all []row
	for rows.Next() {
		var (
			r      row
			author int64
		)
		if err := rows.Scan(&r.c.ID, &r.target, &r.parent, &author, &r.c.Author.Nickname, &r.c.Body, &r.c.Likes,
			&r.c.Helpful, &r.c.CreatedAt, &r.c.MyReaction); err != nil {
			return nil, fmt.Errorf("social: comment row: %w", err)
		}
		if r.c.Author.Nickname != "" {
			r.c.Author.Initial = strings.ToUpper(string([]rune(r.c.Author.Nickname)[:1]))
		}
		r.c.Mine = viewerID != 0 && author == viewerID
		all = append(all, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("social: comment rows: %w", err)
	}

	replies := map[string][]Comment{}
	for _, r := range all {
		if r.parent != "" {
			replies[r.parent] = append(replies[r.parent], r.c)
		}
	}
	for _, r := range all {
		if r.parent == "" {
			r.c.Replies = replies[r.c.ID]
			out[r.target] = append(out[r.target], r.c)
		}
	}

	return out, nil
}
