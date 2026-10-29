package controllers

import (
	"context"
	"strconv"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/i18n"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/library"
	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// PublicProfile serves /u/{nickname} and GET /app/v1/users/{nickname} (SDD 14 §6.2, §6.3, §7.2).
type PublicProfile struct {
	db     *database.DB
	guards []string
	api    bool
}

// NewPublicProfile builds the handlers; guards are tried in order to find an optional viewer.
func NewPublicProfile(db *database.DB, api bool, guards ...string) *PublicProfile {
	return &PublicProfile{db: db, guards: guards, api: api}
}

func (h *PublicProfile) viewerID(ctx context.Context) int64 {
	for _, g := range h.guards {
		if u, err := auth.UserFrom[models.User](ctx, g); err == nil {
			return u.ID
		}
	}

	return 0
}

// Show answers the profile. A private profile that is not the viewer's own is 200 with the nickname only and
// noindex: the header makes crawlers skip it whatever the page component renders.
func (h *PublicProfile) Show(c *web.Ctx) error {
	ctx := c.Context()

	offset := 0
	if tab := c.Query("tab"); tab == "" || tab == "ratings" {
		offset = library.ParseOffset(c.Query("cursor"))
	}

	p, err := library.LoadProfile(ctx, h.db, c.Param("nickname"), h.viewerID(ctx), i18n.Locale(ctx), offset)
	if err != nil {
		return err
	}

	hidden := p.IsPrivate && !p.IsOwner
	if hidden {
		c.Writer().Header().Set("X-Robots-Tag", "noindex")
	}

	head := map[string]any{
		"nickname": p.Nickname, "initial": p.Initial, "isPrivate": p.IsPrivate, "isOwner": p.IsOwner,
	}

	if h.api {
		if hidden {
			return c.JSON(200, map[string]any{"profile": head})
		}

		return c.JSON(200, map[string]any{
			"profile": head, "stats": p.Stats, "ratings": p.Ratings, "lists": p.Lists,
			"meta": map[string]any{"next_cursor": nextCursor(p.NextRatings)},
		})
	}

	props := inertia.Props{"profile": head, "seo": map[string]any{"indexable": !hidden}}
	if !hidden {
		meta := inertia.ScrollMeta{PageName: "cursor"}
		if p.NextRatings >= 0 {
			meta.NextPage = strconv.Itoa(p.NextRatings)
		}

		props["stats"] = p.Stats
		props["ratings"] = inertia.Scroll(map[string]any{"data": p.Ratings}, meta)
		props["lists"] = p.Lists
	}

	return inertia.Render(c, "User/Profile", props)
}

func nextCursor(n int) any {
	if n < 0 {
		return nil
	}

	return strconv.Itoa(n)
}
