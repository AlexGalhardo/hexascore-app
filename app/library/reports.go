package library

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/events"
	"github.com/AlexGalhardo/vukem/uuid"

	"github.com/AlexGalhardo/hexascore-app/app/notifications"
)

// Report targets, reasons and outcomes (SDD 14 §5).
const (
	TargetList = "list"
	TargetUser = "user"

	OutcomeDismissed  = "dismissed"
	OutcomeListHidden = "list_hidden"
	OutcomeUserBanned = "user_banned"
	OutcomeNoAction   = "no_action"

	MaxReportDetails = 500
	MaxAdminNote     = 1000
)

// ReportReasons and ReportOutcomes mirror the table's CHECKs; suggestion_rejected waits for SDD 04.
var (
	ReportReasons  = []string{"spam", "offensive", "harassment", "impersonation", "other"}
	ReportOutcomes = []string{OutcomeDismissed, OutcomeListHidden, OutcomeUserBanned, OutcomeNoAction}
)

// Problem codes (SDD 14 §6.3, §6.4).
const (
	CodeReportSelf    = "report-self"
	CodeReportNotOpen = "report-not-open"
)

// Reports is the report write path: members file them, admins resolve them.
type Reports struct {
	DB     *database.DB
	Events *events.Bus
	Now    func() time.Time
	// Ban bans a user on behalf of an admin (accounts.Ban); a user who is already banned is not an error.
	Ban func(ctx context.Context, actorID, userID int64, reason string) error
}

// ReportInput is a validated report; Target is a list id or a nickname.
type ReportInput struct {
	TargetType string
	Target     string
	Reason     string
	Details    string
}

func problem(status int, key, code, msg string) error {
	return errs.NewKey(status, key, msg, map[string]any{"code": code})
}

// Create files a report by reporterID. A repeat of an open report succeeds without a new row (created is
// false). A list the reporter cannot see answers 404 like a missing one; reporting yourself answers 403
// report-self.
func (s *Reports) Create(ctx context.Context, reporterID int64, in ReportInput) (created bool, err error) {
	targetID, owner, err := s.resolveTarget(ctx, reporterID, in)
	if err != nil {
		return false, err
	}

	if owner == reporterID {
		return false, problem(http.StatusForbidden, "reports.self", CodeReportSelf, "You cannot report yourself.")
	}

	var details any
	if d := strings.TrimSpace(in.Details); d != "" {
		details = d
	}

	res, err := s.DB.ExecContext(ctx,
		`INSERT INTO reports (id, reporter_id, target_type, target_id, reason, details, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT DO NOTHING`,
		uuid.NewV7().String(), reporterID, in.TargetType, targetID, in.Reason, details, s.Now().UTC())
	if err != nil {
		return false, err
	}

	n, err := res.RowsAffected()

	return n == 1, err
}

// resolveTarget returns the stored target id and the user it belongs to (the list owner or the user itself).
func (s *Reports) resolveTarget(ctx context.Context, reporterID int64, in ReportInput) (string, int64, error) {
	switch in.TargetType {
	case TargetList:
		if _, err := uuid.Parse(in.Target); err != nil {
			return "", 0, ErrNotFound()
		}

		var owner int64
		var public, hidden, private bool

		err := s.DB.QueryRowContext(ctx,
			`SELECT l.user_id, l.is_public, l.hidden, u.profile_private FROM lists l JOIN users u ON u.id = l.user_id
WHERE l.id = $1`, in.Target).Scan(&owner, &public, &hidden, &private)
		if errors.Is(err, sql.ErrNoRows) {
			return "", 0, ErrNotFound()
		}

		if err != nil {
			return "", 0, err
		}

		if owner != reporterID && (!public || hidden || private) {
			return "", 0, ErrNotFound()
		}

		return in.Target, owner, nil
	case TargetUser:
		var id int64

		err := s.DB.QueryRowContext(ctx, `SELECT id FROM users WHERE username = $1 AND anonymized_at IS NULL`,
			strings.ToLower(strings.TrimSpace(in.Target))).Scan(&id)
		if errors.Is(err, sql.ErrNoRows) {
			return "", 0, ErrNotFound()
		}

		return strconv.FormatInt(id, 10), id, err
	}

	return "", 0, ErrNotFound()
}

// ResolveInput is an admin's decision on an open report.
type ResolveInput struct {
	Outcome   string
	AdminNote string
}

type openReport struct {
	status, targetType, targetID string
}

// Resolve applies the outcome to the target and closes every open report on it alike, notifying each reporter
// (SDD 14 §7.3). A report that is not open answers 409 report-not-open.
func (s *Reports) Resolve(ctx context.Context, actorID int64, id string, in ResolveInput) error {
	rep, err := s.load(ctx, s.DB, id)
	if err != nil {
		return err
	}

	if rep.status != "open" {
		return notOpen()
	}

	if err := validOutcome(rep.targetType, in.Outcome); err != nil {
		return err
	}

	if in.Outcome == OutcomeUserBanned {
		uid, err := strconv.ParseInt(rep.targetID, 10, 64)
		if err != nil {
			return ErrNotFound()
		}

		if err := s.Ban(ctx, actorID, uid, "Report "+id+": "+in.AdminNote); err != nil {
			return err
		}
	}

	return s.DB.Tx(ctx, func(tx *database.Tx) error {
		// Re-read under lock: two admins can resolve the same report.
		rep, err := s.load(ctx, tx, id, true)
		if err != nil {
			return err
		}

		if rep.status != "open" {
			return notOpen()
		}

		status := "resolved"
		if in.Outcome == OutcomeDismissed {
			status = "dismissed"
		}

		var note any
		if n := strings.TrimSpace(in.AdminNote); n != "" {
			note = n
		}

		all, err := closeOpenReports(ctx, tx, rep, status, in.Outcome, note, actorID, s.Now().UTC())
		if err != nil {
			return err
		}

		for _, c := range all {
			if !c.reporter.Valid {
				continue
			}

			_, err := notifications.Send(ctx, tx, c.reporter.Int64, notifications.ReportResolved, "report:"+c.id,
				map[string]any{"reportId": c.id, "targetType": rep.targetType, "outcome": in.Outcome})
			if err != nil {
				return err
			}
		}

		if in.Outcome == OutcomeListHidden {
			return s.hideList(ctx, tx, rep.targetID, id)
		}

		return nil
	})
}

type closedReport struct {
	id       string
	reporter sql.NullInt64
}

// closeOpenReports closes every open report on the target of rep alike and returns them.
func closeOpenReports(
	ctx context.Context, tx *database.Tx, rep openReport, status, outcome string, note any, actorID int64, at time.Time,
) ([]closedReport, error) {
	rows, err := tx.QueryContext(ctx, `UPDATE reports SET status = $1, outcome = $2, admin_note = $3,
resolved_by = $4, resolved_at = $5
WHERE target_type = $6 AND target_id = $7 AND status = 'open' RETURNING id, reporter_id`,
		status, outcome, note, actorID, at, rep.targetType, rep.targetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []closedReport

	for rows.Next() {
		var c closedReport
		if err := rows.Scan(&c.id, &c.reporter); err != nil {
			return nil, err
		}

		all = append(all, c)
	}

	return all, rows.Err()
}

func (s *Reports) hideList(ctx context.Context, tx *database.Tx, listID, reportID string) error {
	var owner int64
	var title string

	err := tx.QueryRowContext(ctx,
		`UPDATE lists SET hidden = true, updated_at = $2 WHERE id = $1 RETURNING user_id, title`,
		listID, s.Now().UTC()).Scan(&owner, &title)
	if errors.Is(err, sql.ErrNoRows) {
		return nil // the owner deleted the list meanwhile; the reports are closed all the same
	}

	if err != nil {
		return err
	}

	if _, err := notifications.Send(ctx, tx, owner, notifications.ListHidden, "report:"+reportID,
		map[string]any{"listId": listID, "title": title}); err != nil {
		return err
	}

	if s.Events != nil {
		tx.AfterCommit(func(ctx context.Context) {
			_ = s.Events.Dispatch(ctx, ListVisibilityChanged{ListID: listID, Visible: false})
		})
	}

	return nil
}

func (s *Reports) load(ctx context.Context, ex database.Executor, id string, lock ...bool) (openReport, error) {
	if _, err := uuid.Parse(id); err != nil {
		return openReport{}, ErrNotFound()
	}

	q := `SELECT status, target_type, target_id FROM reports WHERE id = $1`
	if len(lock) > 0 && lock[0] {
		q += " FOR UPDATE"
	}

	var r openReport

	err := ex.QueryRowContext(ctx, q, id).Scan(&r.status, &r.targetType, &r.targetID)
	if errors.Is(err, sql.ErrNoRows) {
		return r, ErrNotFound()
	}

	return r, err
}

func validOutcome(targetType, outcome string) error {
	ok := outcome == OutcomeDismissed || outcome == OutcomeNoAction ||
		(outcome == OutcomeListHidden && targetType == TargetList) ||
		(outcome == OutcomeUserBanned && targetType == TargetUser)
	if !ok {
		return problem(http.StatusUnprocessableEntity, "reports.outcome", "invalid-outcome",
			"That outcome does not apply to this report.")
	}

	return nil
}

func notOpen() error {
	return problem(http.StatusConflict, "reports.not_open", CodeReportNotOpen, "The report is already resolved.")
}
