package admin

import (
	"bytes"
	"context"
	"database/sql"
	"embed"
	"errors"
	"io/fs"
	"net/http"
	"net/url"
	"strings"
	"time"

	vadmin "github.com/AlexGalhardo/vukem/admin"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/orm"
	"github.com/AlexGalhardo/vukem/queue"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/view"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/app/community"
	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/moderation"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/entities"
)

//go:embed views
var viewsRoot embed.FS

// ViewsFS holds the template overrides of the panel (layout, show page) and the pages of CatalogHandlers.
func ViewsFS() (fs.FS, error) { return fs.Sub(viewsRoot, "views") }

var affiliateStores = map[string]bool{"amazon": true, "mercadolivre": true, "steam": true, "gog": true}

const (
	maxAffiliateURL = 2048
	panelCSP        = "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; " +
		"form-action 'self'; frame-ancestors 'none'; base-uri 'none'; object-src 'none'"
)

// CatalogHandlers serve the item management pages and the ingestion status page behind the panel gate and CSRF.
type CatalogHandlers struct {
	Accounts *accounts.Accounts
	Queue    *queue.Queue
	Views    *view.Engine
	// Authorize is the panel gate check (404 for non-admins).
	Authorize func(c *web.Ctx) error
	// FlagsURL is the vukem feature-flags page; empty when no flag is registered.
	FlagsURL string
	// AuthorizeApply is the admin-only applySuggestion check every catalogue field edit needs (owner Q-A15).
	AuthorizeApply func(c *web.Ctx) error
	// Editor writes panel field edits with their catalog_audit rows.
	Editor *community.Applier
}

type navLink struct {
	Label, URL string
	Active     bool
}

type panelData struct {
	Title, Home, CSS, JS, AuditURL, FlagsURL string
	Nav                                      []navLink
}

type affiliateRow struct {
	Store, URL string
	Active     bool
}

type sourceRow struct {
	Source                   string
	Records, Pending         int64
	LastFetched, LastChecked string
}

type runRow struct {
	ID                                            int64
	Job, Source, Status, Error, Started, Finished string
	Seen, Changed, Failed                         int
	Rollbackable                                  bool
}

// pageData feeds the panel layout (same field names as vukem's admin page) and the pages below.
type pageData struct {
	Title         string
	Panel         panelData
	Item          models.Item
	Affiliates    []affiliateRow
	Ready         bool
	Sources       []sourceRow
	Runs          []runRow
	Queue         queueView
	Detail        moderation.Detail
	BulkResults   []moderation.BulkResult
	RejectReasons []string
	// CanApply shows the admin-only approve and reject controls (owner Q-A15).
	CanApply bool
	Audit    auditView
	// CanRevert shows the revert form on /admin/catalog-audit.
	CanRevert     bool
	RevertReasons []string
	Locks         []moderation.Lock
	Metrics       moderation.Metrics
	LockTargets   []string
	Reviews       []entities.OpenReviewRow
	Ingest        IngestStatus
}

// panelNav mirrors the sidebar vukem builds from admin.Resources (same order, same plural labels);
// TestAdminIngestionSharesPanelNavigation fails when the two drift apart.
var panelNav = []navLink{
	{Label: "Users", URL: "/admin/users"},
	{Label: "Score bursts", URL: "/admin/score-bursts"},
	{Label: "Rating flags", URL: "/admin/rating-flags"},
	{Label: "Score actions", URL: "/admin/score-actions"},
	{Label: "Score formula versions", URL: "/admin/score-formula-versions"},
	{Label: "Reports", URL: "/admin/reports"},
	{Label: "Contact messages", URL: "/admin/contact-messages"},
	{Label: "Items", URL: "/admin/items"},
	{Label: "Job runs", URL: "/admin/job-runs"},
	{Label: "Ingest failures", URL: "/admin/ingest-failures"},
	{Label: "Dq results", URL: "/admin/dq-results"},
}

func (h CatalogHandlers) newPage(title string) pageData {
	return pageData{Title: title, Panel: panelData{
		Title: "HexaScore admin", Home: "/admin", CSS: "/admin/_assets/admin.css", AuditURL: "/admin/_audit",
		FlagsURL: h.FlagsURL, Nav: panelNav,
	}}
}

func (h CatalogHandlers) render(c *web.Ctx, name string, d pageData) error {
	var buf bytes.Buffer
	if err := h.Views.Render(c.Context(), &buf, name, d); err != nil {
		return errs.Wrap(http.StatusInternalServerError, err)
	}

	c.SetHeader("Cache-Control", "no-store")
	c.SetHeader("X-Frame-Options", "DENY")
	c.SetHeader("Content-Security-Policy", panelCSP)

	return c.Blob(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
}

func (h CatalogHandlers) actor(c *web.Ctx) (models.User, error) {
	if err := h.Authorize(c); err != nil {
		return models.User{}, err
	}

	return auth.UserFrom[models.User](c.Context(), h.Accounts.Web.Name())
}

// itemID reads the {id} path parameter; anything but a uuid is a 404.
func itemID(c *web.Ctx) (string, error) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		return "", errs.ErrNotFound
	}

	return id, nil
}

// Manage renders GET /admin/items/{id}/manage: affiliate links and the merge form.
func (h CatalogHandlers) Manage(c *web.Ctx) error {
	if _, err := h.actor(c); err != nil {
		return err
	}

	id, err := itemID(c)
	if err != nil {
		return err
	}

	d := h.newPage("Manage item")
	d.Item.ID = id
	db := h.Accounts.DB.SQL()

	err = db.QueryRowContext(c.Context(), `SELECT type, slug FROM items WHERE id = $1`, id).
		Scan(&d.Item.Type, &d.Item.Slug)
	if errors.Is(err, sql.ErrNoRows) {
		return errs.ErrNotFound
	}

	if err != nil {
		return err
	}

	rows, err := db.QueryContext(c.Context(),
		`SELECT store, url, active FROM affiliate_links WHERE item_id = $1 ORDER BY store`, id)
	if err != nil {
		return err
	}

	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var r affiliateRow
		if err := rows.Scan(&r.Store, &r.URL, &r.Active); err != nil {
			return err
		}

		d.Affiliates = append(d.Affiliates, r)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return h.render(c, "items_manage", d)
}

func audit(ctx context.Context, ex database.Executor, actor models.User, record, action string,
	changes map[string]vadmin.Change,
) error {
	return orm.Create(ctx, ex, &vadmin.AuditLog{
		ActorID: actor.AuthID(), Resource: "affiliate-links", RecordID: record, Action: action,
		Changes: orm.JSON[map[string]vadmin.Change]{V: changes},
	})
}

func validAffiliateURL(raw string) bool {
	u, err := url.Parse(raw)

	return err == nil && len(raw) <= maxAffiliateURL && u.Scheme == "https" && u.Host != ""
}

// SaveAffiliateLink handles POST /admin/items/{id}/affiliate-links (upsert) and audits it.
func (h CatalogHandlers) SaveAffiliateLink(c *web.Ctx) error {
	actor, err := h.actor(c)
	if err != nil {
		return err
	}

	id, err := itemID(c)
	if err != nil {
		return err
	}

	store, link := c.FormValue("store"), strings.TrimSpace(c.FormValue("url"))
	active := c.FormValue("active") == "true"

	v := errs.NewValidation()
	if !affiliateStores[store] {
		v.Append("store", "Unknown store.")
	}

	if !validAffiliateURL(link) {
		v.Append("url", "The URL must be an https link of at most 2048 characters.")
	}

	if !v.Empty() {
		return v
	}

	err = h.Accounts.DB.Tx(c.Context(), func(tx *database.Tx) error {
		var (
			oldURL    sql.NullString
			oldActive sql.NullBool
		)

		err := tx.QueryRowContext(c.Context(),
			`SELECT url, active FROM affiliate_links WHERE item_id = $1 AND store = $2 FOR UPDATE`, id, store).
			Scan(&oldURL, &oldActive)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		_, err = tx.ExecContext(
			c.Context(),
			`INSERT INTO affiliate_links (item_id, store, url, active) VALUES ($1, $2, $3, $4)
ON CONFLICT (item_id, store) DO UPDATE SET url = EXCLUDED.url, active = EXCLUDED.active`,
			id,
			store,
			link,
			active,
		)
		if err != nil {
			return err
		}

		action := vadmin.ActionUpdate

		var before, beforeActive any
		if oldURL.Valid {
			before, beforeActive = oldURL.String, oldActive.Bool
		} else {
			action = vadmin.ActionCreate
		}

		return audit(c.Context(), tx, actor, id+"/"+store, action, map[string]vadmin.Change{
			"url":    {Before: before, After: link},
			"active": {Before: beforeActive, After: active},
		})
	})
	if err != nil {
		return mapItemErr(err)
	}

	return c.Redirect("/admin/items/" + id + "/manage")
}

// DeleteAffiliateLink handles POST /admin/items/{id}/affiliate-links/{store}/delete and audits it.
func (h CatalogHandlers) DeleteAffiliateLink(c *web.Ctx) error {
	actor, err := h.actor(c)
	if err != nil {
		return err
	}

	id, err := itemID(c)
	if err != nil {
		return err
	}

	store := c.Param("store")
	if !affiliateStores[store] {
		return errs.ErrNotFound
	}

	err = h.Accounts.DB.Tx(c.Context(), func(tx *database.Tx) error {
		var old string

		err := tx.QueryRowContext(c.Context(),
			`DELETE FROM affiliate_links WHERE item_id = $1 AND store = $2 RETURNING url`, id, store).Scan(&old)
		if errors.Is(err, sql.ErrNoRows) {
			return errs.ErrNotFound
		}

		if err != nil {
			return err
		}

		return audit(c.Context(), tx, actor, id+"/"+store, vadmin.ActionDelete,
			map[string]vadmin.Change{"url": {Before: old}})
	})
	if err != nil {
		return err
	}

	return c.Redirect("/admin/items/" + id + "/manage")
}

// Merge handles POST /admin/items/{id}/merge: folds the item into form field "into" through catalog.Merge.
func (h CatalogHandlers) Merge(c *web.Ctx) error {
	actor, err := h.actor(c)
	if err != nil {
		return err
	}

	from, err := itemID(c)
	if err != nil {
		return err
	}

	into := strings.TrimSpace(c.FormValue("into"))
	if _, err := uuid.Parse(into); err != nil {
		return errs.NewValidation().Add("into", "The target must be an item id.")
	}

	if c.FormValue("confirm") != "merge" {
		return errs.NewValidation().Add("confirm", "Confirm the merge.")
	}

	if err := catalog.Merge(c.Context(), h.Accounts.DB, h.Queue, from, into, actor.ID); err != nil {
		return mapMergeErr(err)
	}

	return c.Redirect("/admin/items/" + into)
}

func mapMergeErr(err error) error {
	switch {
	case errors.Is(err, catalog.ErrMergeNotFound):
		return errs.ErrNotFound
	case errors.Is(err, catalog.ErrMergeSelf), errors.Is(err, catalog.ErrMergeTypeMismatch),
		errors.Is(err, catalog.ErrMergeRedirected):
		return errs.New(http.StatusConflict, err.Error())
	}

	return err
}

// mapItemErr turns the foreign-key failure of an unknown item into a 404.
func mapItemErr(err error) error {
	if strings.Contains(err.Error(), "affiliate_links_item_id_fkey") {
		return errs.ErrNotFound
	}

	return err
}

// Ingestion renders GET /admin/ingestion. The tables come with the ingestion phase (SDD 06 §5), so they are probed
// first and absent ones render a "not ready" note.
func (h CatalogHandlers) Ingestion(c *web.Ctx) error {
	if _, err := h.actor(c); err != nil {
		return err
	}

	ctx, db := c.Context(), h.Accounts.DB.SQL()
	d := h.newPage("Ingestion")

	var recs, runs sql.NullString
	if err := db.QueryRowContext(ctx, `SELECT to_regclass('source_records')::text, to_regclass('job_runs')::text`).
		Scan(&recs, &runs); err != nil {
		return err
	}

	d.Ready = recs.Valid && runs.Valid
	if !d.Ready {
		return h.render(c, "ingestion", d)
	}

	var err error
	if d.Sources, err = loadSources(ctx, db); err != nil {
		return err
	}

	if d.Runs, err = loadRuns(ctx, db); err != nil {
		return err
	}

	return h.render(c, "ingestion", d)
}

func loadSources(ctx context.Context, db *sql.DB) ([]sourceRow, error) {
	rows, err := db.QueryContext(ctx, `SELECT source, count(*),
  count(*) FILTER (WHERE normalized_at IS NULL OR normalized_at < fetched_at), max(fetched_at), max(checked_at)
FROM source_records GROUP BY source ORDER BY source`)
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	var out []sourceRow

	for rows.Next() {
		var (
			r      sourceRow
			f, chk sql.NullTime
		)
		if err := rows.Scan(&r.Source, &r.Records, &r.Pending, &f, &chk); err != nil {
			return nil, err
		}

		r.LastFetched, r.LastChecked = fmtTime(f), fmtTime(chk)
		out = append(out, r)
	}

	return out, rows.Err()
}

func loadRuns(ctx context.Context, db *sql.DB) ([]runRow, error) {
	rows, err := db.QueryContext(
		ctx,
		`SELECT DISTINCT ON (job, coalesce(source, '')) id, job, coalesce(source, ''), status, coalesce(error, ''),
  seen, changed, failed, started_at, finished_at
FROM job_runs ORDER BY job, coalesce(source, ''), created_at DESC`,
	)
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	var out []runRow

	for rows.Next() {
		var (
			r       runRow
			st, fin sql.NullTime
		)
		if err := rows.Scan(
			&r.ID,
			&r.Job,
			&r.Source,
			&r.Status,
			&r.Error,
			&r.Seen,
			&r.Changed,
			&r.Failed,
			&st,
			&fin,
		); err != nil {
			return nil, err
		}

		r.Started, r.Finished = fmtTime(st), fmtTime(fin)
		out = append(out, r)
	}

	return out, rows.Err()
}

func fmtTime(t sql.NullTime) string {
	if !t.Valid {
		return ""
	}

	return t.Time.UTC().Format(time.DateTime)
}
