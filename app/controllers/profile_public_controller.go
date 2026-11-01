package controllers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/i18n"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/library"
	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/social"
)

// PublicProfile serves /u/{nickname} and GET /app/v1/users/{nickname} (SDD 14 §6.2, §6.3, §7.2).
type PublicProfile struct {
	db     *database.DB
	guards []string
	api    bool
	// coversBase makes the JSON cover URLs absolute (COVERS_BASE_URL); the Inertia page keeps paths.
	coversBase string
	// Follows adds the `social` prop to the web page when the feature flag `social` is on (SDD 14 §12.2).
	Follows *social.Follows
}

// NewPublicProfile builds the handlers; guards are tried in order to find an optional viewer.
func NewPublicProfile(db *database.DB, api bool, coversBase string, guards ...string) *PublicProfile {
	return &PublicProfile{db: db, guards: guards, api: api, coversBase: coversBase}
}

func (h *PublicProfile) viewer(ctx context.Context) models.User {
	for _, g := range h.guards {
		if u, err := auth.UserFrom[models.User](ctx, g); err == nil {
			return u
		}
	}

	return models.User{}
}

// contributions is the SDD 04 FR-04-15 profile block: counts and approval rate once the user has at least one
// decision, plus tier and accuracy for the user and admins only. nil when there is nothing to show.
func (h *PublicProfile) contributions(ctx context.Context, userID int64, private bool) (map[string]any, error) {
	var (
		submitted, approved, rejected int
		accuracy                      float64
		tier                          string
	)

	err := h.db.QueryRowContext(ctx, `SELECT submitted_total, approved_total, rejected_total, accuracy::float8,
  tier::text FROM contributor_stats WHERE user_id = $1`, userID).Scan(&submitted, &approved, &rejected, &accuracy, &tier)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && approved+rejected == 0) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("profile: contributions: %w", err)
	}

	out := map[string]any{
		"submitted": submitted, "approved": approved, "rejected": rejected,
		"approvalRate": float64(approved) / float64(approved+rejected),
	}
	if private {
		out["tier"], out["accuracy"] = tier, accuracy
	}

	return out, nil
}

// Show answers the profile. A private profile that is not the viewer's own is 200 with the nickname only and
// noindex: the header makes crawlers skip it whatever the page component renders.
func (h *PublicProfile) Show(c *web.Ctx) error {
	ctx := c.Context()

	offset := 0
	if tab := c.Query("tab"); tab == "" || tab == "ratings" {
		offset = library.ParseOffset(c.Query("cursor"))
	}

	viewer := h.viewer(ctx)
	p, err := library.LoadProfile(ctx, h.db, c.Param("nickname"), viewer.ID, i18n.Locale(ctx), offset)
	if err != nil {
		if h.api {
			return err
		}

		return pageError(c, err)
	}

	hidden := p.IsPrivate && !p.IsOwner
	indexable := !hidden && !p.Banned
	if !indexable {
		c.Writer().Header().Set("X-Robots-Tag", "noindex")
	}

	head := map[string]any{
		"nickname": p.Nickname, "initial": p.Initial, "isPrivate": p.IsPrivate, "isOwner": p.IsOwner,
	}

	if h.api {
		if hidden {
			return c.JSON(200, map[string]any{"profile": head})
		}

		for i := range p.Ratings {
			p.Ratings[i].Cover = p.Ratings[i].Cover.WithBase(h.coversBase)
		}

		return c.JSON(200, map[string]any{
			"profile": head, "stats": p.Stats, "ratings": p.Ratings, "lists": p.Lists,
			"meta": map[string]any{"next_cursor": nextCursor(p.NextRatings)},
		})
	}

	props := inertia.Props{"profile": head, "indexable": indexable}
	if !p.Banned {
		follow, ok, err := h.Follows.ProfileProps(ctx, viewer.ID, p.UserID, hidden)
		if err != nil {
			return err
		}
		if ok {
			props["social"] = follow
		}

		if ok && !hidden && c.Query("tab") == "reviews" {
			locale := i18n.Locale(ctx)
			props["reviews"] = inertia.Defer(func(ctx context.Context) (any, error) {
				return h.Follows.UserReviews(ctx, viewer.ID, p.UserID, locale)
			}).Group("personal")
		}

		if ok && c.Query("tab") == "activity" {
			allowed, err := h.Follows.ActivityAllowed(ctx, viewer.ID, p.UserID)
			if err != nil {
				return err
			}
			// null tells the page "not shared" apart from "still loading" (undefined).
			props["activity"] = nil
			if allowed {
				locale := i18n.Locale(ctx)
				props["activity"] = inertia.Defer(func(ctx context.Context) (any, error) {
					return h.Follows.Activity(ctx, p.UserID, locale)
				}).Group("personal")
			}
		}
	}

	if !hidden {
		meta := inertia.ScrollMeta{PageName: "cursor"}
		if p.NextRatings >= 0 {
			meta.NextPage = strconv.Itoa(p.NextRatings)
		}

		props["stats"] = p.Stats
		props["ratings"] = inertia.Scroll(map[string]any{"data": p.Ratings}, meta)
		props["lists"] = p.Lists

		contrib, err := h.contributions(ctx, p.UserID, p.IsOwner || viewer.Role == models.RoleAdmin)
		if err != nil {
			return err
		}
		if contrib != nil {
			props["contributions"] = contrib
		}
	}

	return inertia.Render(c, "User/Profile", props)
}

func nextCursor(n int) any {
	if n < 0 {
		return nil
	}

	return strconv.Itoa(n)
}
