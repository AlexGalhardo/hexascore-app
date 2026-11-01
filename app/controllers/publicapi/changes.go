package publicapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/AlexGalhardo/vukem/api"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/api/changes"
	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	appmiddleware "github.com/AlexGalhardo/hexascore-app/app/http/middleware"
)

// ChangeJSON is one listChanges row. action is created, updated or deleted (SDD 05 §8).
type ChangeJSON struct {
	ID         string   `json:"id"                   validate:"required"`
	Action     string   `json:"action"               validate:"required,oneof=created updated deleted"`
	ItemType   string   `json:"itemType"             validate:"required,oneof=games movies series books"`
	ItemID     string   `json:"itemId"               validate:"required"`
	Fields     []string `json:"fields"`
	MergedInto *string  `json:"mergedInto,omitempty"`
	ChangedAt  string   `json:"changedAt"            validate:"required"`
}

// ChangePageJSON is one keyset page of changes in the vukem envelope.
type ChangePageJSON struct {
	Data  []ChangeJSON `json:"data"`
	Meta  api.Meta     `json:"meta"`
	Links api.Links    `json:"links"`
}

// ListChangesInput is the listChanges query. since is checked in the handler so a bad value is 400, not 422.
type ListChangesInput struct {
	Since  string `json:"since"  doc:"RFC 3339, at most 14 days ago"`
	Type   string `json:"type"   doc:"games, movies, series or books"`
	Cursor string `json:"cursor" doc:"opaque next_cursor of the previous page"`
	Limit  string `json:"limit"  doc:"1 to 1000, default 100"`
}

// ListChanges is listChanges: GET /api/v1/changes (SDD 05 FR-05-9).
func (h *Controller) ListChanges(c *web.Ctx, in ListChangesInput) (ChangePageJSON, error) {
	if err := checkQuery(c, "since", "type", "cursor", "limit"); err != nil {
		return ChangePageJSON{}, err
	}

	since, err := parseSince(in.Since, h.now())
	if err != nil {
		return ChangePageJSON{}, err
	}

	itemType, ok := typeOf[in.Type]
	if in.Type != "" && !ok {
		return ChangePageJSON{}, invalid("type must be games, movies, series or books")
	}

	limit, err := changesLimit(in.Limit)
	if err != nil {
		return ChangePageJSON{}, err
	}

	after, err := h.changesAfter(in.Cursor, in.Type)
	if err != nil {
		return ChangePageJSON{}, err
	}

	rows, more, err := changes.List(c.Context(), h.DB, changes.Filter{
		Since: since, ItemType: itemType, After: after, Limit: limit,
	})
	if err != nil {
		return ChangePageJSON{}, err
	}

	out := ChangePageJSON{
		Data:  make([]ChangeJSON, len(rows)),
		Meta:  api.Meta{PerPage: limit},
		Links: api.Links{Self: c.Request().URL.RequestURI()},
	}
	for i, row := range rows {
		out.Data[i] = changeJSON(row)
	}
	if more {
		last := rows[len(rows)-1]
		out.Meta.NextCursor = h.Codec.Encode(catalog.Cursor{
			S: changesSort(in.Type), V: strconv.FormatInt(last.ChangedAt.UTC().UnixMicro(), 10),
			ID: strconv.FormatInt(last.ID, 10), D: catalog.DirNext,
		})
		q := c.Request().URL.Query()
		q.Set("cursor", out.Meta.NextCursor)
		next := c.Request().URL.Path + "?" + q.Encode()
		out.Links.Next = &next
	}

	return out, NotModified(c, out)
}

func (h *Controller) now() time.Time {
	if h.Now != nil {
		return h.Now()
	}

	return time.Now()
}

// parseSince accepts an instant at most changes.Window before now. The equal bound is allowed;
// the future is invalid-query, not changes-window-exceeded (SDD 05 §8).
func parseSince(raw string, now time.Time) (time.Time, error) {
	if raw == "" {
		return time.Time{}, invalid("since is required")
	}

	since, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, invalid("since must be RFC 3339")
	}

	now, since = now.UTC(), since.UTC()
	if since.After(now) {
		return time.Time{}, invalid("since is in the future")
	}
	if since.Before(now.Add(-changes.Window)) {
		return time.Time{}, errs.New(http.StatusBadRequest, "since is more than 14 days ago").
			WithHeader(appmiddleware.ProblemCodeHeader, "changes-window-exceeded")
	}

	return since, nil
}

// changesLimit is 1..1000. listChanges does not use the plan page size (SDD 05 §6.2: limit ≤ 1000).
func changesLimit(raw string) (int, error) {
	if raw == "" {
		return changes.DefaultLimit, nil
	}

	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		return 0, invalid("limit must be a positive integer")
	}
	if n > changes.MaxLimit {
		return 0, invalid("limit exceeds 1000")
	}

	return n, nil
}

func changesSort(section string) string {
	if section == "" {
		return "changes"
	}

	return "changes:" + section
}

func (h *Controller) changesAfter(token, section string) (*changes.After, error) {
	if token == "" {
		return nil, nil
	}

	cur, err := h.Codec.Decode(token, changesSort(section))
	if err != nil || cur.D != catalog.DirNext {
		return nil, badQuery(catalog.ErrInvalidCursor)
	}

	id, err := strconv.ParseInt(cur.ID, 10, 64)
	us, err2 := strconv.ParseInt(cur.V, 10, 64)
	if err != nil || err2 != nil {
		return nil, badQuery(catalog.ErrInvalidCursor)
	}

	return &changes.After{At: time.UnixMicro(us).UTC(), ID: id}, nil
}

func changeJSON(row changes.Row) ChangeJSON {
	out := ChangeJSON{
		ID: strconv.FormatInt(row.ID, 10), Action: row.Action, ItemType: sectionOf(row.ItemType),
		ItemID: row.ItemID, Fields: row.Fields, ChangedAt: row.ChangedAt.UTC().Format(time.RFC3339Nano),
	}
	if row.MergedInto != "" {
		out.MergedInto = &row.MergedInto
	}

	return out
}
