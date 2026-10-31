// Package moderation is the admin side of community edits: the queue (SDD 04 FR-04-8, section 7.4 "Queue") and,
// in later rows, the detail view and the apply service.
package moderation

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/app/community"
)

// Age badges (NFR-04-2): green under 12 h, amber 12 to 24 h, red over 24 h.
const (
	AgeGreen = "green"
	AgeAmber = "amber"
	AgeRed   = "red"

	greenLimit = 12 * time.Hour
	redLimit   = 24 * time.Hour

	// DefaultPageSize is the queue page length when Queue.PageSize is zero.
	DefaultPageSize = 50
)

// ErrBadCursor is a malformed pagination cursor.
var ErrBadCursor = errors.New("moderation: bad cursor")

// ErrNotFound is a suggestion id that does not exist.
var ErrNotFound = errors.New("moderation: suggestion not found")

// ErrNotAssignee is a moderator releasing a claim held by someone else.
var ErrNotAssignee = errors.New("moderation: claim belongs to another moderator")

// Queue reads the moderation queue and moves suggestions in and out of review.
type Queue struct {
	DB       *database.DB
	Now      func() time.Time
	PageSize int
}

// Filter narrows the queue. Zero values mean "any"; the default status is submitted plus in_review.
type Filter struct {
	Status, Kind, TargetType, Domain, Field, Locale, Age string
	Author, Assignee                                     int64
	Competing, Stale                                     bool
	// After is the Next cursor of the previous page.
	After string
}

// Row is one queue line.
type Row struct {
	ID, Kind, TargetType, TargetID, Domain, Locale, Status string
	Fields                                                 []string
	AuthorID, AssignedTo                                   int64
	AuthorTier                                             string
	AuthorAccuracy                                         float64
	Version                                                int
	SubmittedAt                                            time.Time
	Age                                                    string
	Competing, Stale, Unreachable                          bool
}

// Page is one keyset page; Next is empty on the last one.
type Page struct {
	Rows []Row
	Next string
}

func (q Queue) now() time.Time {
	if q.Now != nil {
		return q.Now()
	}

	return time.Now()
}

// AgeBadge classifies a submission time.
func AgeBadge(submitted, now time.Time) string {
	switch age := now.Sub(submitted); {
	case age > redLimit:
		return AgeRed
	case age >= greenLimit:
		return AgeAmber
	default:
		return AgeGreen
	}
}

// List returns the queue oldest first, keyset-paginated on (submitted_at, id) (VF-005 workaround: no framework
// cursor helper).
func (q Queue) List(ctx context.Context, f Filter) (Page, error) {
	size := q.PageSize
	if size <= 0 {
		size = DefaultPageSize
	}

	var (
		where = []string{"s.status IN ('submitted','in_review')"}
		args  []any
	)

	arg := func(v any) string {
		args = append(args, v)

		return "$" + strconv.Itoa(len(args))
	}
	eq := func(col, v string) {
		if v != "" {
			where = append(where, col+"::text = "+arg(v))
		}
	}

	if f.Status != "" {
		where[0] = "s.status::text = " + arg(f.Status)
	}

	eq("s.kind", f.Kind)
	eq("s.target_type", f.TargetType)
	eq("s.target_domain", f.Domain)
	eq("s.locale", f.Locale)

	if f.Field != "" {
		where = append(where, "s.changes ? "+arg(f.Field))
	}

	if f.Author != 0 {
		where = append(where, "s.author_id = "+arg(f.Author))
	}

	if f.Assignee != 0 {
		where = append(where, "s.assigned_to = "+arg(f.Assignee))
	}

	now := q.now()

	switch f.Age {
	case AgeGreen:
		where = append(where, "s.submitted_at > "+arg(now.Add(-greenLimit)))
	case AgeAmber:
		where = append(
			where,
			"s.submitted_at <= "+arg(now.Add(-greenLimit))+" AND s.submitted_at >= "+arg(now.Add(-redLimit)),
		)
	case AgeRed:
		where = append(where, "s.submitted_at < "+arg(now.Add(-redLimit)))
	}

	if f.Competing {
		where = append(where, `s.meta @> '{"competing":true}'`)
	}

	if f.Stale {
		where = append(where, `s.meta @> '{"stale":true}'`)
	}

	if f.After != "" {
		at, id, err := parseCursor(f.After)
		if err != nil {
			return Page{}, err
		}

		where = append(where, "(s.submitted_at, s.id) > ("+arg(at)+", "+arg(id)+"::uuid)")
	}

	rows, err := q.DB.QueryContext(ctx, `
SELECT s.id::text, s.kind::text, s.target_type::text, COALESCE(s.target_id::text, ''), s.target_domain, s.locale,
  s.status::text,
  CASE WHEN jsonb_typeof(s.changes) = 'object'
    THEN (SELECT COALESCE(string_agg(k, ',' ORDER BY k), '') FROM jsonb_object_keys(s.changes) k) ELSE '' END,
  COALESCE(s.author_id, 0), COALESCE(s.assigned_to, 0), COALESCE(cs.tier::text, ''), COALESCE(cs.accuracy, 0),
  s.version, s.submitted_at,
  COALESCE((s.meta ->> 'competing')::boolean, false), COALESCE((s.meta ->> 'stale')::boolean, false),
  COALESCE(s.meta ->> 'source_status' = 'unreachable', false)
FROM suggestions s
LEFT JOIN contributor_stats cs ON cs.user_id = s.author_id
WHERE `+strings.Join(where, " AND ")+`
ORDER BY s.submitted_at, s.id
LIMIT `+strconv.Itoa(size+1), args...)
	if err != nil {
		return Page{}, fmt.Errorf("moderation: list queue: %w", err)
	}

	defer func() { _ = rows.Close() }()

	var page Page

	for rows.Next() {
		var (
			r      Row
			fields string
		)

		if err := rows.Scan(&r.ID, &r.Kind, &r.TargetType, &r.TargetID, &r.Domain, &r.Locale, &r.Status, &fields,
			&r.AuthorID, &r.AssignedTo, &r.AuthorTier, &r.AuthorAccuracy, &r.Version, &r.SubmittedAt,
			&r.Competing, &r.Stale, &r.Unreachable); err != nil {
			return Page{}, fmt.Errorf("moderation: scan queue: %w", err)
		}

		if fields != "" {
			r.Fields = strings.Split(fields, ",")
		}

		r.Age = AgeBadge(r.SubmittedAt, now)
		page.Rows = append(page.Rows, r)
	}

	if err := rows.Err(); err != nil {
		return Page{}, fmt.Errorf("moderation: list queue: %w", err)
	}

	if len(page.Rows) > size {
		page.Rows = page.Rows[:size]
		last := page.Rows[size-1]
		page.Next = strconv.FormatInt(last.SubmittedAt.UnixMicro(), 10) + "_" + last.ID
	}

	return page, nil
}

func parseCursor(s string) (time.Time, string, error) {
	micro, id, ok := strings.Cut(s, "_")
	if !ok {
		return time.Time{}, "", ErrBadCursor
	}

	n, err := strconv.ParseInt(micro, 10, 64)
	if err != nil || len(id) != 36 {
		return time.Time{}, "", ErrBadCursor
	}

	return time.UnixMicro(n).UTC(), id, nil
}

// Claim moves a submitted suggestion to in_review for the actor. version is the one the page showed, so a second
// moderator claiming the same row gets a 409 stale-before from the state machine.
func (q Queue) Claim(ctx context.Context, id string, version int, actor community.Actor) error {
	return q.DB.Tx(ctx, func(tx *database.Tx) error {
		if _, err := assignee(ctx, tx, id); err != nil {
			return err
		}

		if err := community.TransitionTx(ctx, tx, id, version, "in_review", actor); err != nil {
			return err
		}

		if _, err := tx.ExecContext(
			ctx,
			`UPDATE suggestions SET assigned_to = $2, claimed_at = $3 WHERE id = $1`,
			id,
			actor.ID,
			q.now(),
		); err != nil {
			return fmt.Errorf("moderation: record claim: %w", err)
		}

		return nil
	})
}

// Release puts an in_review suggestion back in the queue. An admin releases any claim; a moderator only its own.
func (q Queue) Release(ctx context.Context, id string, version int, actor community.Actor) error {
	return q.DB.Tx(ctx, func(tx *database.Tx) error {
		holder, err := assignee(ctx, tx, id)
		if err != nil {
			return err
		}

		if actor.Role != "admin" && (!holder.Valid || !actor.ID.Valid || holder.Int64 != actor.ID.Int64) {
			return ErrNotAssignee
		}

		if err := community.TransitionTx(ctx, tx, id, version, "submitted", actor); err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx,
			`UPDATE suggestions SET assigned_to = NULL, claimed_at = NULL WHERE id = $1`, id); err != nil {
			return fmt.Errorf("moderation: clear claim: %w", err)
		}

		return nil
	})
}

func assignee(ctx context.Context, tx *database.Tx, id string) (sql.NullInt64, error) {
	var holder sql.NullInt64

	err := tx.QueryRowContext(ctx, `SELECT assigned_to FROM suggestions WHERE id = $1 FOR UPDATE`, id).Scan(&holder)
	if errors.Is(err, sql.ErrNoRows) {
		return holder, ErrNotFound
	}

	if err != nil {
		return holder, fmt.Errorf("moderation: load suggestion: %w", err)
	}

	return holder, nil
}
