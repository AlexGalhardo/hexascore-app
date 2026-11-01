package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/flags"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/rollback"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

const maxTriggerQIDs = 500

var (
	triggerModes = []string{
		run.ModeBootstrap, run.ModeIncremental, run.ModeFull, run.ModeManual, run.ModeBackfill,
	}
	qidPattern = regexp.MustCompile(`^Q[1-9][0-9]{0,15}$`)
)

// IngestHandlers serve /admin/ingest (viewIngestion) and its two POST routes (runIngestion), CSRF on both. Custom
// panel actions are not released, so triggering and cancelling are app routes (VF-008).
type IngestHandlers struct {
	Pages CatalogHandlers
	DB    *database.DB
	Queue run.Dispatcher
	Flags *flags.Flags
	// AuthorizeView and AuthorizeRun check viewIngestion and runIngestion.
	AuthorizeView, AuthorizeRun, AuthorizeRollback func(c *web.Ctx) error
	Now                                            func() time.Time
}

// IngestStage is a (job, source) option of the trigger form.
type IngestStage struct{ Job, Source string }

// IngestFailureSummary counts the dead-letter targets of one source.
type IngestFailureSummary struct {
	Source        string
	Total, Parked int
}

// IngestAssert is a failing assert of the latest dq.asserts run.
type IngestAssert struct {
	Code, Severity, Observed string
}

// IngestCoverage is how many sampled items missed a field in one run, per type.
type IngestCoverage struct {
	Type, Day        string
	Sampled, Missing int
}

// IngestStatus feeds the status part of the page.
type IngestStatus struct {
	Failures    []IngestFailureSummary
	Asserts     []IngestAssert
	OldestTMDB  string
	Coverage    []IngestCoverage
	Stages      []IngestStage
	Modes       []string
	CanRun      bool
	CanRollback bool
}

// Index renders GET /admin/ingest.
func (h IngestHandlers) Index(c *web.Ctx) error {
	if err := authorizeOr403(h.AuthorizeView, c); err != nil {
		return err
	}

	ctx, db := c.Context(), h.DB.SQL()
	d := h.Pages.newPage("Ingestion")
	d.Ready = true

	var err error
	if d.Runs, err = loadRuns(ctx, db); err != nil {
		return err
	}

	if d.Ingest, err = h.status(ctx, db); err != nil {
		return err
	}

	d.Ingest.CanRun = authorizeOr403(h.AuthorizeRun, c) == nil
	d.Ingest.CanRollback = authorizeOr403(h.AuthorizeRollback, c) == nil

	for i, r := range d.Runs {
		d.Runs[i].Rollbackable = d.Ingest.CanRollback && r.Status != run.StatusQueued &&
			r.Status != run.StatusRunning &&
			strings.HasPrefix(r.Job, "fetch.")
	}

	return h.Pages.render(c, "ingest", d)
}

func (h IngestHandlers) status(ctx context.Context, db *sql.DB) (IngestStatus, error) {
	st := IngestStatus{Modes: triggerModes}

	for _, s := range ingest.Stages() {
		st.Stages = append(st.Stages, IngestStage{Job: s.Job, Source: s.Source})
	}

	rows, err := db.QueryContext(ctx, `SELECT source, count(*), count(*) FILTER (WHERE parked)
FROM ingest_failures GROUP BY source ORDER BY source`)
	if err != nil {
		return st, err
	}

	if err := scanAll(rows, func() error {
		var f IngestFailureSummary
		if err := rows.Scan(&f.Source, &f.Total, &f.Parked); err != nil {
			return err
		}

		st.Failures = append(st.Failures, f)

		return nil
	}); err != nil {
		return st, err
	}

	rows, err = db.QueryContext(ctx, `SELECT code, severity, coalesce(observed::text, '') FROM dq_results
WHERE run_id = (SELECT max(r.run_id) FROM dq_results r JOIN job_runs j ON j.id = r.run_id WHERE j.job = 'dq.asserts')
  AND NOT passed ORDER BY severity, code`)
	if err != nil {
		return st, err
	}

	if err := scanAll(rows, func() error {
		var a IngestAssert
		if err := rows.Scan(&a.Code, &a.Severity, &a.Observed); err != nil {
			return err
		}

		st.Asserts = append(st.Asserts, a)

		return nil
	}); err != nil {
		return st, err
	}

	var oldest sql.NullTime
	if err := db.QueryRowContext(ctx, `SELECT min(fetched_at) FROM source_records WHERE source = 'tmdb'`).
		Scan(&oldest); err != nil {
		return st, err
	}

	if oldest.Valid {
		st.OldestTMDB = strconv.Itoa(int(h.now().Sub(oldest.Time).Hours()/24)) + " days"
	}

	rows, err = db.QueryContext(ctx, `SELECT f.type, to_char(j.created_at, 'YYYY-MM-DD'), max(f.sampled), sum(f.missing)
FROM field_coverage f JOIN job_runs j ON j.id = f.run_id
WHERE j.created_at > now() - interval '14 days'
GROUP BY f.type, f.run_id, j.created_at ORDER BY j.created_at DESC, f.type LIMIT 60`)
	if err != nil {
		return st, err
	}

	err = scanAll(rows, func() error {
		var cv IngestCoverage
		if err := rows.Scan(&cv.Type, &cv.Day, &cv.Sampled, &cv.Missing); err != nil {
			return err
		}

		st.Coverage = append(st.Coverage, cv)

		return nil
	})

	return st, err
}

func (h IngestHandlers) now() time.Time {
	if h.Now != nil {
		return h.Now()
	}

	return time.Now()
}

func scanAll(rows *sql.Rows, each func() error) error {
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		if err := each(); err != nil {
			return err
		}
	}

	return rows.Err()
}

// Trigger handles POST /admin/ingest/runs. 422 bad input, 423 source or master switch off, 409 run already active.
func (h IngestHandlers) Trigger(c *web.Ctx) error {
	if err := authorizeOr403(h.AuthorizeRun, c); err != nil {
		return err
	}

	u, err := auth.UserFrom[models.User](c.Context(), h.Pages.Accounts.Web.Name())
	if err != nil {
		return err
	}

	spec, err := triggerSpec(c, u.ID)
	if err != nil {
		return err
	}

	if !run.Allowed(c.Context(), h.Flags, spec.Source) {
		return errs.New(http.StatusLocked, "This source is disabled by its kill switch.")
	}

	if _, err := run.Create(c.Context(), h.DB, h.Queue, spec); err != nil {
		if errors.Is(err, run.ErrRunActive) {
			return errs.New(http.StatusConflict, "A run of this job is already active.")
		}

		return err
	}

	return c.Redirect("/admin/ingest")
}

func invalid(msg string) error { return errs.New(http.StatusUnprocessableEntity, msg) }

func triggerSpec(c *web.Ctx, adminID int64) (run.Spec, error) {
	spec := run.Spec{
		Job: c.FormValue("job"), Source: strings.TrimSpace(c.FormValue("source")), Mode: c.FormValue("mode"),
		TriggeredBy: "admin:" + strconv.FormatInt(adminID, 10),
	}

	if !slices.Contains(triggerModes, spec.Mode) {
		return spec, invalid("Unknown mode.")
	}

	stages := ingest.Stages()

	i := slices.IndexFunc(stages, func(s ingest.Stage) bool {
		return s.Job == spec.Job && (spec.Source == "" || s.Source == spec.Source)
	})
	if i < 0 {
		return spec, invalid("Unknown job or source.")
	}

	spec.Source = stages[i].Source
	params := map[string]any{}

	if raw := strings.TrimSpace(c.FormValue("limit")); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 {
			return spec, invalid("Limit must be a positive number.")
		}

		params["limit"] = n
	}

	if raw := strings.TrimSpace(c.FormValue("since")); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return spec, invalid("Since must be an RFC 3339 time.")
		}

		params["since"] = t.UTC()
	}

	if qids := strings.Fields(strings.ReplaceAll(c.FormValue("qids"), ",", " ")); len(qids) > 0 {
		if len(qids) > maxTriggerQIDs {
			return spec, invalid("At most 500 QIDs per run.")
		}

		for _, q := range qids {
			if !qidPattern.MatchString(q) {
				return spec, invalid("QIDs look like Q42.")
			}
		}

		params["qids"] = qids
	}

	if c.FormValue("allow_count_drop") != "" {
		reason := strings.TrimSpace(c.FormValue("reason"))
		if reason == "" || len(reason) > 500 {
			return spec, invalid("allow_count_drop needs a reason of up to 500 characters.")
		}

		params["allow_count_drop"] = true
		params["reason"] = reason
	}

	b, err := json.Marshal(params)
	if err != nil {
		return spec, err
	}

	spec.Params = b

	return spec, nil
}

// Cancel handles POST /admin/ingest/runs/{id}/cancel: the run stops at its next chunk boundary.
func (h IngestHandlers) Cancel(c *web.Ctx) error {
	if err := authorizeOr403(h.AuthorizeRun, c); err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		return errs.ErrNotFound
	}

	res, err := h.DB.ExecContext(c.Context(), `UPDATE job_runs SET status = 'cancelled', finished_at = now()
WHERE id = $1 AND status IN ('queued', 'running')`, id)
	if err != nil {
		return err
	}

	if n, _ := res.RowsAffected(); n == 0 {
		var one int
		if err := h.DB.QueryRowContext(c.Context(), `SELECT 1 FROM job_runs WHERE id = $1`, id).Scan(&one); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errs.ErrNotFound
			}

			return err
		}

		return errs.New(http.StatusConflict, "This run already ended.")
	}

	return c.Redirect("/admin/ingest")
}

// Rollback handles POST /admin/ingest/runs/{id}/rollback (rollbackIngestion): it queues the undo of a finished
// fetch run (SDD 06 §7.8). 404 unknown run, 409 history-expired, already rolled back or another rollback active,
// 422 without a reason or for a run that is not a finished fetch run.
func (h IngestHandlers) Rollback(c *web.Ctx) error {
	if err := authorizeOr403(h.AuthorizeRollback, c); err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		return errs.ErrNotFound
	}

	reason := strings.TrimSpace(c.FormValue("reason"))
	if reason == "" || len(reason) > 500 {
		return invalid("A rollback needs a reason of up to 500 characters.")
	}

	u, err := auth.UserFrom[models.User](c.Context(), h.Pages.Accounts.Web.Name())
	if err != nil {
		return err
	}

	_, err = ingest.StartRollback(c.Context(), h.DB, h.Queue, id, u.ID, reason)

	switch {
	case err == nil:
		return c.Redirect("/admin/ingest")
	case errors.Is(err, rollback.ErrRunNotFound):
		return errs.ErrNotFound
	case errors.Is(err, rollback.ErrNotRollbackable):
		return invalid("Only a finished fetch run can be rolled back.")
	case errors.Is(err, rollback.ErrHistoryExpired):
		return errs.New(http.StatusConflict, "history-expired: restore the dump instead (runbook in SDD 11).")
	case errors.Is(err, rollback.ErrAlreadyRolledBack):
		return errs.New(http.StatusConflict, "This run was already rolled back.")
	case errors.Is(err, run.ErrRunActive):
		return errs.New(http.StatusConflict, "run-active: another rollback is under way.")
	}

	return err
}
