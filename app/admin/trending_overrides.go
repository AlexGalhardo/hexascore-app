package admin

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	vadmin "github.com/AlexGalhardo/vukem/admin"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/orm"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

const (
	maxTrendingReason = 200
	trendingResource  = "trending-overrides"
)

type trendingRow struct {
	ID                  int64
	Kind, Until, Reason string
}

// trendingOverrides lists the item's pins and bans, active or not, newest first.
func (h CatalogHandlers) trendingOverrides(ctx context.Context, itemID string) ([]trendingRow, error) {
	rows, err := h.Accounts.DB.SQL().
		QueryContext(ctx, `SELECT id, kind, coalesce(to_char(until, 'YYYY-MM-DD'), ''), reason
FROM trending_overrides WHERE item_id = $1 ORDER BY created_at DESC, id DESC`, itemID)
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	var out []trendingRow

	for rows.Next() {
		var r trendingRow
		if err := rows.Scan(&r.ID, &r.Kind, &r.Until, &r.Reason); err != nil {
			return nil, err
		}

		out = append(out, r)
	}

	return out, rows.Err()
}

func auditTrending(ctx context.Context, ex database.Executor, actor models.User, record, action string,
	changes map[string]vadmin.Change,
) error {
	return orm.Create(ctx, ex, &vadmin.AuditLog{
		ActorID: actor.AuthID(), Resource: trendingResource, RecordID: record, Action: action,
		Changes: orm.JSON[map[string]vadmin.Change]{V: changes},
	})
}

// SaveTrendingOverride handles POST /admin/items/{id}/trending: pin or ban the item on the trending lists, optionally
// until a date (UTC, inclusive). One override per kind and item stays: a new one replaces the old. Both take effect at
// the next hourly ComputeTrending run. Audited.
func (h CatalogHandlers) SaveTrendingOverride(c *web.Ctx) error {
	actor, err := h.actor(c)
	if err != nil {
		return err
	}

	id, err := itemID(c)
	if err != nil {
		return err
	}

	kind, reason := c.FormValue("kind"), strings.TrimSpace(c.FormValue("reason"))

	v := errs.NewValidation()
	if kind != "pinned" && kind != "banned" {
		v.Append("kind", "Choose pinned or banned.")
	}

	if len([]rune(reason)) > maxTrendingReason {
		v.Append("reason", "The reason must be at most 200 characters.")
	}

	var until any

	if raw := strings.TrimSpace(c.FormValue("until")); raw != "" {
		day, perr := time.Parse(time.DateOnly, raw)

		switch {
		case perr != nil:
			v.Append("until", "Use a date as YYYY-MM-DD.")
		case !day.AddDate(0, 0, 1).After(time.Now().UTC()):
			v.Append("until", "The date must not be in the past.")
		default:
			until = day.AddDate(0, 0, 1) // the end of that day
		}
	}

	if !v.Empty() {
		return v
	}

	err = h.Accounts.DB.Tx(c.Context(), func(tx *database.Tx) error {
		var one int
		if err := tx.QueryRowContext(c.Context(), `SELECT 1 FROM items WHERE id = $1`, id).Scan(&one); err != nil {
			return err
		}

		var old sql.NullString

		err := tx.QueryRowContext(c.Context(), `DELETE FROM trending_overrides WHERE item_id = $1 AND kind = $2
RETURNING reason`, id, kind).Scan(&old)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		if _, err := tx.ExecContext(
			c.Context(),
			`INSERT INTO trending_overrides (item_id, kind, until, reason, admin_id)
VALUES ($1, $2, $3, $4, $5)`,
			id,
			kind,
			until,
			reason,
			actor.ID,
		); err != nil {
			return err
		}

		action, before := vadmin.ActionCreate, any(nil)
		if old.Valid {
			action, before = vadmin.ActionUpdate, old.String
		}

		return auditTrending(c.Context(), tx, actor, id+"/"+kind, action, map[string]vadmin.Change{
			"reason": {Before: before, After: reason},
		})
	})
	if err != nil {
		return mapTrendingErr(err)
	}

	return c.Redirect("/admin/items/" + id + "/manage")
}

// mapTrendingErr turns a missing item or override into a 404.
func mapTrendingErr(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return errs.ErrNotFound
	}

	return mapItemErr(err)
}

// DeleteTrendingOverride handles POST /admin/items/{id}/trending/{override}/delete. An override of another item is a
// 404. Audited.
func (h CatalogHandlers) DeleteTrendingOverride(c *web.Ctx) error {
	actor, err := h.actor(c)
	if err != nil {
		return err
	}

	id, err := itemID(c)
	if err != nil {
		return err
	}

	oid, err := strconv.ParseInt(c.Param("override"), 10, 64)
	if err != nil || oid < 1 {
		return errs.ErrNotFound
	}

	err = h.Accounts.DB.Tx(c.Context(), func(tx *database.Tx) error {
		var kind, reason string
		if err := tx.QueryRowContext(c.Context(), `DELETE FROM trending_overrides WHERE id = $1 AND item_id = $2
RETURNING kind, reason`, oid, id).Scan(&kind, &reason); err != nil {
			return err
		}

		return auditTrending(c.Context(), tx, actor, id+"/"+kind, vadmin.ActionDelete, map[string]vadmin.Change{
			"reason": {Before: reason, After: nil},
		})
	})
	if err != nil {
		return mapTrendingErr(err)
	}

	return c.Redirect("/admin/items/" + id + "/manage")
}
