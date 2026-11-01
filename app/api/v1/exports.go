package v1

import (
	"context"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/api"
	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// ExportTTL is how long a getExport download link stays valid (SDD 05 §6.2).
const ExportTTL = 5 * time.Minute

// Bucket is the slice of the S3 disk the exports routes use. A nil Bucket (no S3_BUCKET, development) lists
// nothing and 404s every day.
type Bucket interface {
	List(ctx context.Context, dir string) ([]string, error)
	Exists(ctx context.Context, p string) (bool, error)
	PresignGet(p string, expires time.Duration) (string, error)
}

// Exports is listExports and getExport.
type Exports struct {
	Bucket Bucket
	Plans  api.PlanResolver
}

// ExportJSON is one daily ID export.
type ExportJSON struct {
	Type string `json:"type" validate:"required,oneof=games movies series books"`
	Day  string `json:"day"  validate:"required"`
}

// ExportList is listExports, newest day first.
type ExportList struct {
	Data []ExportJSON `json:"data"`
}

// ExportInput is the getExport path.
type ExportInput struct {
	Type string `json:"type" enum:"games,movies,series,books" doc:"games, movies, series or books"`
	Day  string `json:"day"                                   doc:"UTC day, YYYY-MM-DD"`
}

var exportKey = regexp.MustCompile(`^exports/([a-z]+)_ids_(\d{4})_(\d{2})_(\d{2})\.json\.gz$`)

// List is GET /exports, from the bucket prefix (no table).
func (h *Exports) List(c *web.Ctx, _ struct{}) (ExportList, error) {
	if err := h.requirePlan(c); err != nil {
		return ExportList{}, err
	}

	out := ExportList{Data: []ExportJSON{}}
	if h.Bucket == nil {
		return out, nil
	}

	keys, err := h.Bucket.List(c.Context(), "exports")
	if err != nil {
		return ExportList{}, err
	}

	for _, k := range keys {
		if m := exportKey.FindStringSubmatch(k); m != nil {
			out.Data = append(out.Data, ExportJSON{Type: m[1], Day: m[2] + "-" + m[3] + "-" + m[4]})
		}
	}

	slices.SortStableFunc(out.Data, func(a, b ExportJSON) int {
		if a.Day != b.Day {
			return strings.Compare(b.Day, a.Day)
		}

		return strings.Compare(a.Type, b.Type)
	})

	return out, nil
}

// Get is GET /exports/{type}/{day}: 302 to a presigned URL that is never cached.
func (h *Exports) Get(c *web.Ctx, in ExportInput) (struct{}, error) {
	if err := h.requirePlan(c); err != nil {
		return struct{}{}, err
	}

	// The enum tag only documents the values; an unknown type is a 400 invalid-query, never a 422 (FR-05-4).
	if !slices.Contains([]string{"games", "movies", "series", "books"}, in.Type) {
		return struct{}{}, errs.New(
			http.StatusBadRequest,
			"invalid-query: unknown type: use games, movies, series or books",
		)
	}

	day, err := time.Parse(time.DateOnly, in.Day)
	if err != nil {
		return struct{}{}, errs.New(http.StatusBadRequest, "invalid-query: day must be YYYY-MM-DD")
	}

	if h.Bucket == nil {
		return struct{}{}, errs.ErrNotFound
	}

	key := "exports/" + in.Type + "_ids_" + day.Format("2006_01_02") + ".json.gz"

	ok, err := h.Bucket.Exists(c.Context(), key)
	if err != nil {
		return struct{}{}, err
	}

	if !ok {
		return struct{}{}, errs.ErrNotFound
	}

	link, err := h.Bucket.PresignGet(key, ExportTTL)
	if err != nil {
		return struct{}{}, err
	}

	return struct{}{}, errs.New(http.StatusFound, "export").
		WithHeader("Location", link).WithHeader("Cache-Control", "no-store")
}

func (h *Exports) requirePlan(c *web.Ctx) error {
	u, err := auth.User[models.User](c, api.GuardName)
	if err != nil {
		return err
	}

	plan, err := h.Plans(c.Context(), u.ID)
	if err != nil {
		return api.Unavailable(err)
	}

	if plan.ID != api.PlanPro && plan.ID != api.PlanBusiness {
		return errs.New(http.StatusForbidden, "plan required").WithExtension("code", "plan-required")
	}

	return nil
}
