package admin

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/flags"
	"github.com/AlexGalhardo/vukem/web"
)

// The React resource screens (ADR 0026) read the same tables as the framework panel's resources, but only the
// columns named below: a column that is not listed is never selected, so a secret (a password hash, a 2FA secret,
// a recovery code) cannot reach the page by accident. Every SQL fragment here is a constant of this file; request
// input only ever travels as a bound parameter or as a lookup key into these maps.

const (
	resPerPage    = 25
	resMaxPage    = 10000
	resQueryLimit = 8 * time.Second
)

// resCol is one column of a screen. Kind tells the page how to draw the value: text, id, number, bool, time, date,
// badge or json.
type resCol struct {
	Key, SQL, Kind string
	Sortable       bool
}

// resFilter narrows a list by one column. Kind "enum" compares the column to one of Options; "bool" to true or false.
type resFilter struct {
	Key, SQL, Kind string
	Options        []string
}

type resSpec struct {
	Slug, NavKey, Table string
	// IDExpr identifies a record in its URL; empty means the resource has no detail page.
	IDExpr string
	List   []resCol
	Detail []resCol
	// Search are SQL expressions matched with ILIKE. Prefix matches only the start, for the large tables.
	Search      []string
	Prefix      bool
	DefaultSort string
	Filters     []resFilter
}

func text(key, sqlExpr string) resCol  { return resCol{Key: key, SQL: sqlExpr, Kind: "text"} }
func sortable(c resCol) resCol         { c.Sortable = true; return c }
func idCol(key, sqlExpr string) resCol { return resCol{Key: key, SQL: sqlExpr, Kind: "id"} }
func num(key, sqlExpr string) resCol   { return resCol{Key: key, SQL: sqlExpr, Kind: "number"} }
func flag(key, sqlExpr string) resCol  { return resCol{Key: key, SQL: sqlExpr, Kind: "bool"} }
func when(key, sqlExpr string) resCol  { return resCol{Key: key, SQL: sqlExpr, Kind: "time"} }
func day(key, sqlExpr string) resCol   { return resCol{Key: key, SQL: sqlExpr + "::text", Kind: "date"} }
func badge(key, sqlExpr string) resCol { return resCol{Key: key, SQL: sqlExpr, Kind: "badge"} }
func blob(key, sqlExpr string) resCol  { return resCol{Key: key, SQL: sqlExpr + "::text", Kind: "json"} }

var userRoles = []string{"user", "moderator", "admin"}

// resSpecs is the allow-list of screens, in sidebar order.
var resSpecs = []resSpec{
	{
		Slug: "users", NavKey: "users", Table: "users", IDExpr: "id::text", DefaultSort: "id DESC", Prefix: true,
		Search:  []string{"name", "email", "username"},
		Filters: []resFilter{{Key: "role", SQL: "role", Kind: "enum", Options: userRoles}},
		List: []resCol{
			sortable(num("id", "id")), sortable(text("name", "name")), sortable(text("email", "email")),
			badge("role", "role"), when("email_verified_at", "email_verified_at"), when("banned_at", "banned_at"),
			sortable(when("created_at", "created_at")),
		},
		Detail: []resCol{
			num("id", "id"), text("name", "name"), text("email", "email"), text("username", "username"),
			badge("role", "role"), text("locale", "locale"), when("email_verified_at", "email_verified_at"),
			flag("two_factor_enabled", "two_factor_confirmed_at IS NOT NULL"),
			when("created_at", "created_at"), when("updated_at", "updated_at"),
			when("banned_at", "banned_at"), text("ban_reason", "ban_reason"), when("ban_expires_at", "ban_expires_at"),
			when("deletion_scheduled_at", "deletion_scheduled_at"), when("anonymized_at", "anonymized_at"),
			flag("profile_private", "profile_private"), flag("personalization_opt_out", "personalization_opt_out"),
			flag("community_mail_opt_out", "community_mail_opt_out"), flag("feed_opt_in", "feed_opt_in"),
			flag("is_seed", "is_seed"),
		},
	},
	{
		Slug: "score-bursts", NavKey: "scoreBursts", Table: "score_bursts", IDExpr: "id::text", DefaultSort: "id DESC",
		Search: []string{"item_id::text", "trigger"},
		List: []resCol{
			sortable(num("id", "id")), idCol("item_id", "item_id::text"), text("trigger", "trigger"),
			sortable(when("started_at", "started_at")), badge("resolution", "coalesce(resolution, '')"),
			when("resolved_at", "resolved_at"),
		},
		Detail: []resCol{
			num("id", "id"), idCol("item_id", "item_id::text"), text("trigger", "trigger"),
			when("started_at", "started_at"), when("limited_until", "limited_until"),
			badge("resolution", "coalesce(resolution, '')"), when("resolved_at", "resolved_at"),
			num("resolved_by", "resolved_by"), blob("stats", "stats"),
		},
	},
	{
		Slug: "rating-flags", NavKey: "ratingFlags", Table: "rating_flags", DefaultSort: "created_at DESC",
		Search: []string{"user_id::text", "item_id::text", "reason"},
		List: []resCol{
			num("user_id", "user_id"), idCol("item_id", "item_id::text"), num("shadow_weight", "shadow_weight"),
			text("reason", "reason"), num("burst_id", "burst_id"), sortable(when("created_at", "created_at")),
		},
	},
	{
		Slug:        "score-actions",
		NavKey:      "scoreActions",
		Table:       "score_actions",
		IDExpr:      "id::text",
		DefaultSort: "id DESC",
		Search:      []string{"kind", "item_id::text", "note"},
		List: []resCol{
			sortable(num("id", "id")), sortable(when("at", "at")), num("actor", "actor"), badge("kind", "kind"),
			idCol("item_id", "item_id::text"), num("user_id", "user_id"), num("burst_id", "burst_id"),
		},
		Detail: []resCol{
			num("id", "id"), when("at", "at"), num("actor", "actor"), badge("kind", "kind"),
			idCol("item_id", "item_id::text"), num("user_id", "user_id"), num("burst_id", "burst_id"),
			text("note", "note"), blob("before", `"before"`), blob("after", `"after"`),
		},
	},
	{
		Slug:        "score-formula-versions",
		NavKey:      "formulaVersions",
		Table:       "score_formula_versions",
		IDExpr:      "version::text",
		DefaultSort: "version DESC",
		List: []resCol{
			sortable(num("version", "version")), num("m_game", "m_game"), num("m_movie", "m_movie"),
			num("m_series", "m_series"), num("m_book", "m_book"), sortable(when("activated_at", "activated_at")),
		},
		Detail: []resCol{
			num("version", "version"), num("m_game", "m_game"), num("m_movie", "m_movie"), num("m_series", "m_series"),
			num("m_book", "m_book"), when("activated_at", "activated_at"), text("changelog", "changelog"),
			blob("params", "params"),
		},
	},
	{
		Slug: "reports", NavKey: "reports", Table: "reports", IDExpr: "id::text", DefaultSort: "created_at DESC",
		Search:  []string{"reason", "target_id", "details"},
		Filters: []resFilter{{Key: "status", SQL: "status", Kind: "enum", Options: []string{"open", "resolved"}}},
		List: []resCol{
			idCol("id", "id::text"), badge("status", "status"), text("target_type", "target_type"),
			text("target_id", "target_id"), text("reason", "reason"), badge("outcome", "coalesce(outcome, '')"),
			sortable(when("created_at", "created_at")),
		},
		Detail: []resCol{
			idCol("id", "id::text"), badge("status", "status"), num("reporter_id", "reporter_id"),
			text("target_type", "target_type"), text("target_id", "target_id"), text("reason", "reason"),
			text("details", "details"), badge("outcome", "coalesce(outcome, '')"), text("admin_note", "admin_note"),
			num("resolved_by", "resolved_by"), when("resolved_at", "resolved_at"), when("created_at", "created_at"),
		},
	},
	{
		Slug: "contact-messages", NavKey: "contactMessages", Table: "contact_messages", IDExpr: "id::text",
		DefaultSort: "created_at DESC", Search: []string{"name", "email", "subject", "message"},
		Filters: []resFilter{{Key: "handled", SQL: "handled", Kind: "bool"}},
		List: []resCol{
			idCol("id", "id::text"), text("name", "name"), text("email", "email"), text("subject", "subject"),
			flag("handled", "handled"), sortable(when("created_at", "created_at")),
		},
		Detail: []resCol{
			idCol("id", "id::text"), num("user_id", "user_id"), text("name", "name"), text("email", "email"),
			text("subject", "subject"), text("message", "message"), flag("handled", "handled"),
			when("created_at", "created_at"),
		},
	},
	{
		Slug:        "items",
		NavKey:      "items",
		Table:       "items",
		IDExpr:      "id::text",
		DefaultSort: "updated_at DESC, id DESC",
		Prefix:      true,
		Search:      []string{"slug", "original_title", "wikidata_qid"},
		Filters: []resFilter{
			{Key: "type", SQL: "type", Kind: "enum", Options: []string{"game", "movie", "series", "book"}},
			{Key: "visibility", SQL: "visibility", Kind: "enum", Options: []string{"public", "hidden", "takedown"}},
			{
				Key:     "release_status",
				SQL:     "release_status",
				Kind:    "enum",
				Options: []string{"announced", "released", "cancelled"},
			},
		},
		List: []resCol{
			badge("type", "type"),
			sortable(text("slug", "slug")),
			text("original_title", "original_title"),
			day(
				"release_date",
				"release_date",
			),
			badge("release_status", "release_status"),
			badge("visibility", "visibility"),
			sortable(num("score", "score")),
			sortable(num("rating_count", "rating_count")),
			sortable(when("updated_at", "updated_at")),
		},
		Detail: []resCol{
			idCol("id", "id::text"),
			badge("type", "type"),
			text("slug", "slug"),
			text("wikidata_qid", "wikidata_qid"),
			text("original_title", "original_title"),
			day("release_date", "release_date"),
			text("release_precision", "release_precision"),
			badge("release_status", "release_status"),
			badge(
				"visibility",
				"visibility",
			),
			num("runtime_minutes", "runtime_minutes"),
			num("season_count", "season_count"),
			num("episode_count", "episode_count"),
			num("page_count", "page_count"),
			num("score", "score"),
			num("rating_count", "rating_count"),
			text("score_confidence", "score_confidence"),
			num("popularity_prior", "popularity_prior"),
			num("quality_score", "quality_score"),
			text("cover_key", "cover_key"),
			text("cover_source", "cover_source"),
			blob("external_ids", "external_ids"),
			when("last_verified_at", "last_verified_at"),
			when("created_at", "created_at"),
			when("updated_at", "updated_at"),
		},
	},
	{
		Slug: "job-runs", NavKey: "jobRuns", Table: "job_runs", IDExpr: "id::text", DefaultSort: "id DESC",
		Search: []string{"job", "source", "triggered_by"},
		Filters: []resFilter{{Key: "status", SQL: "status", Kind: "enum", Options: []string{
			"queued", "running", "succeeded", "partial", "failed", "skipped", "cancelled",
		}}},
		List: []resCol{
			sortable(num("id", "id")), text("job", "job"), text("source", "coalesce(source, '')"), text("mode", "mode"),
			badge("status", "status"), num("seen", "seen"), num("changed", "changed"), num("failed", "failed"),
			sortable(when("started_at", "started_at")), when("finished_at", "finished_at"),
		},
		Detail: []resCol{
			num("id", "id"), text("job", "job"), text("source", "coalesce(source, '')"), text("mode", "mode"),
			badge("status", "status"), text("triggered_by", "triggered_by"), num("attempt", "attempt"),
			num("seen", "seen"), num("changed", "changed"), num("failed", "failed"), num("skipped", "skipped"),
			num("rate_limited", "rate_limited"), num("total_estimate", "total_estimate"),
			when("started_at", "started_at"), when("heartbeat_at", "heartbeat_at"), when("finished_at", "finished_at"),
			text("error", "error"), blob("params", "params"), blob("cursor", "cursor"),
		},
	},
	{
		Slug: "ingest-failures", NavKey: "ingestFailures", Table: "ingest_failures", DefaultSort: "last_seen DESC",
		Search:  []string{"source", "external_id", "error_class"},
		Filters: []resFilter{{Key: "parked", SQL: "parked", Kind: "bool"}},
		List: []resCol{
			text("source", "source"), text("external_id", "external_id"), text("error_class", "error_class"),
			num("attempts", "attempts"), flag("parked", "parked"), when("first_seen", "first_seen"),
			sortable(when("last_seen", "last_seen")), text("error", "left(error, 200)"),
		},
	},
	{
		Slug: "dq-results", NavKey: "dqResults", Table: "dq_results", DefaultSort: "checked_at DESC, run_id DESC",
		Search:  []string{"code", "severity"},
		Filters: []resFilter{{Key: "passed", SQL: "passed", Kind: "bool"}},
		List: []resCol{
			num("run_id", "run_id"), text("code", "code"), badge("severity", "severity"), flag("passed", "passed"),
			num("observed", "observed"), num("threshold", "threshold"), sortable(when("checked_at", "checked_at")),
		},
	},
	{
		Slug: "audit-logs", NavKey: "auditLog", Table: "admin_audit_logs", IDExpr: "id::text", DefaultSort: "id DESC",
		Search: []string{"resource", "record_id", "action", "actor_id"},
		List: []resCol{
			sortable(num("id", "id")), sortable(when("created_at", "created_at")), text("actor_id", "actor_id"),
			badge("action", "action"), text("resource", "resource"), text("record_id", "record_id"),
		},
		Detail: []resCol{
			num("id", "id"), when("created_at", "created_at"), text("actor_id", "actor_id"), badge("action", "action"),
			text("resource", "resource"), text("record_id", "record_id"), blob("changes", "changes"),
		},
	},
}

func specBySlug(slug string) (resSpec, bool) {
	i := slices.IndexFunc(resSpecs, func(s resSpec) bool { return s.Slug == slug })
	if i < 0 {
		return resSpec{}, false
	}

	return resSpecs[i], true
}

// ResourceHandlers serve the React list and detail screens of the framework resources and the feature flags.
type ResourceHandlers struct {
	DB *database.DB
	// Flags is the registered feature flags; nil hides the flags screen.
	Flags *flags.Flags
	// Authorize is the panel gate check (404 for non-admins).
	Authorize func(c *web.Ctx) error
}

// resRow is one record: its id and one cell per column, in column order. Cells are positional so the keys of a
// column (occurred_at, user_id) are never renamed by the camelCase conversion of the props.
type resRow struct {
	ID    string `json:"id"`
	Cells []any  `json:"cells"`
}

type resColView struct {
	Key      string `json:"key"`
	Kind     string `json:"kind"`
	Sortable bool   `json:"sortable"`
}

type resFilterView struct {
	Key     string   `json:"key"`
	Kind    string   `json:"kind"`
	Options []string `json:"options"`
	Value   string   `json:"value"`
}

func colViews(cols []resCol) []resColView {
	out := make([]resColView, 0, len(cols))
	for _, c := range cols {
		out = append(out, resColView{Key: c.Key, Kind: c.Kind, Sortable: c.Sortable})
	}

	return out
}

func selectList(cols []resCol, idExpr string) string {
	parts := make([]string, 0, len(cols)+1)
	// The id gets its own output name so ORDER BY id still means the table column, not two outputs called id.
	if idExpr != "" {
		parts = append(parts, idExpr+" AS __id")
	} else {
		parts = append(parts, "'' AS __id")
	}

	for _, c := range cols {
		parts = append(parts, c.SQL)
	}

	return strings.Join(parts, ", ")
}

// likeEscape makes a user term safe inside an ILIKE pattern.
var likeEscape = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

func scanRow(rows *sql.Rows, cols []resCol) (resRow, error) {
	dest := make([]any, len(cols)+1)
	holders := make([]any, len(cols)+1)

	for i := range dest {
		holders[i] = &dest[i]
	}

	if err := rows.Scan(holders...); err != nil {
		return resRow{}, err
	}

	row := resRow{ID: fmt.Sprint(dest[0]), Cells: make([]any, len(cols))}

	for i := range cols {
		v := dest[i+1]
		if b, ok := v.([]byte); ok {
			v = string(b)
		}

		row.Cells[i] = v
	}

	return row, nil
}

// List renders GET /admin/r/{slug}.
func (h ResourceHandlers) List(c *web.Ctx) error {
	if err := h.Authorize(c); err != nil {
		return err
	}

	spec, ok := specBySlug(c.Param("slug"))
	if !ok {
		return errs.ErrNotFound
	}

	var (
		where []string
		args  []any
	)

	arg := func(v any) string {
		args = append(args, v)

		return "$" + strconv.Itoa(len(args))
	}

	q := strings.TrimSpace(c.Query("q"))
	if len(q) > 100 {
		q = q[:100]
	}

	if q != "" && len(spec.Search) > 0 {
		pattern := likeEscape.Replace(q) + "%"
		if !spec.Prefix {
			pattern = "%" + pattern
		}

		ph := arg(pattern)
		clauses := make([]string, 0, len(spec.Search))

		for _, s := range spec.Search {
			clauses = append(clauses, s+"::text ILIKE "+ph)
		}

		where = append(where, "("+strings.Join(clauses, " OR ")+")")
	}

	filters := make([]resFilterView, 0, len(spec.Filters))

	for _, f := range spec.Filters {
		val := c.Query("f_" + f.Key)
		view := resFilterView{Key: f.Key, Kind: f.Kind, Options: f.Options}

		switch {
		case f.Kind == "bool" && (val == "true" || val == "false"):
			where = append(where, f.SQL+" = "+arg(val == "true"))
			view.Value = val
		case f.Kind == "enum" && slices.Contains(f.Options, val):
			where = append(where, f.SQL+" = "+arg(val))
			view.Value = val
		}

		filters = append(filters, view)
	}

	order := spec.DefaultSort
	sortKey := c.Query("sort")
	dir := "DESC"

	for _, col := range spec.List {
		if col.Sortable && col.Key == sortKey {
			if c.Query("dir") == "asc" {
				dir = "ASC"
			}

			order = col.SQL + " " + dir
			if spec.IDExpr != "" {
				order += ", " + spec.IDExpr + " " + dir
			}

			break
		}
	}

	if order == spec.DefaultSort {
		sortKey, dir = "", ""
	}

	page, _ := strconv.Atoi(c.Query("page"))
	page = min(max(page, 1), resMaxPage)

	query := "SELECT " + selectList(spec.List, spec.IDExpr) + " FROM " + spec.Table
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	query += " ORDER BY " + order + " LIMIT " + arg(resPerPage+1) + " OFFSET " + arg((page-1)*resPerPage)

	ctx, cancel := context.WithTimeout(c.Context(), resQueryLimit)
	defer cancel()

	rows, err := h.DB.SQL().QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("admin resource %s: %w", spec.Slug, err)
	}

	defer func() { _ = rows.Close() }()

	out := []resRow{}

	for rows.Next() {
		r, err := scanRow(rows, spec.List)
		if err != nil {
			return fmt.Errorf("admin resource %s: %w", spec.Slug, err)
		}

		out = append(out, r)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("admin resource %s: %w", spec.Slug, err)
	}

	hasNext := len(out) > resPerPage
	if hasNext {
		out = out[:resPerPage]
	}

	return renderReact(c, "Admin/ResourceList", map[string]any{
		"resource": map[string]any{
			"slug":       spec.Slug,
			"navKey":     spec.NavKey,
			"hasDetail":  spec.IDExpr != "",
			"searchable": len(spec.Search) > 0,
		},
		"columns": colViews(spec.List),
		"rows":    out,
		"filters": filters,
		"page":    page,
		"hasNext": hasNext,
		"q":       q,
		"sort":    sortKey,
		"dir":     strings.ToLower(dir),
	})
}

// Show renders GET /admin/r/{slug}/{id}.
func (h ResourceHandlers) Show(c *web.Ctx) error {
	if err := h.Authorize(c); err != nil {
		return err
	}

	spec, ok := specBySlug(c.Param("slug"))
	if !ok || spec.IDExpr == "" {
		return errs.ErrNotFound
	}

	ctx, cancel := context.WithTimeout(c.Context(), resQueryLimit)
	defer cancel()

	rows, err := h.DB.SQL().QueryContext(ctx,
		"SELECT "+selectList(
			spec.Detail,
			spec.IDExpr,
		)+" FROM "+spec.Table+" WHERE "+spec.IDExpr+" = $1 LIMIT 1", c.Param("id"))
	if err != nil {
		return fmt.Errorf("admin resource %s: %w", spec.Slug, err)
	}

	defer func() { _ = rows.Close() }()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return fmt.Errorf("admin resource %s: %w", spec.Slug, err)
		}

		return errs.ErrNotFound
	}

	rec, err := scanRow(rows, spec.Detail)
	if err != nil {
		return fmt.Errorf("admin resource %s: %w", spec.Slug, err)
	}

	return renderReact(c, "Admin/ResourceShow", map[string]any{
		"resource": map[string]any{"slug": spec.Slug, "navKey": spec.NavKey},
		"id":       rec.ID,
		"columns":  colViews(spec.Detail),
		"cells":    rec.Cells,
	})
}

// FlagRow is one registered feature flag with its global value.
type FlagRow struct {
	Name, Description, Kind, Default, Value string
	Locked, Stored                          bool
}

// FlagsPage renders GET /admin/r/flags: the registered flags and their global values. Changing one stays on the framework
// page, which records who changed what in the audit log.
func (h ResourceHandlers) FlagsPage(c *web.Ctx) error {
	if err := h.Authorize(c); err != nil {
		return err
	}

	if h.Flags == nil {
		return errs.ErrNotFound
	}

	states, err := h.Flags.States(c.Context(), flags.Global)
	if err != nil {
		return errs.Wrap(http.StatusServiceUnavailable, err)
	}

	rows := make([]FlagRow, 0, len(states))
	for _, s := range states {
		rows = append(rows, FlagRow{
			Name: s.Name, Description: s.Description, Kind: string(s.Kind), Default: s.Default,
			Value: s.Value, Locked: s.Locked, Stored: s.Stored,
		})
	}

	return renderReact(c, "Admin/Flags", map[string]any{"flags": rows})
}
