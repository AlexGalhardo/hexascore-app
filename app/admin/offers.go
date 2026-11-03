package admin

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"strings"

	vadmin "github.com/AlexGalhardo/vukem/admin"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/orm"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/offers"
)

const offersResource = "offers"

// offerKinds are the offers.kind values an admin may enter.
var offerKinds = []string{"stream", "rent", "buy", "free", "buy_digital", "mobile", "physical", "read"}

type offerRow struct {
	ID, Store, StoreName, Country, Kind, Source, URL, Status string
	Active                                                   bool
}

type storeOption struct{ Code, Name string }

// offerRows lists every offer of the item, active or not, the active ones first.
func (h CatalogHandlers) offerRows(ctx context.Context, itemID string) ([]offerRow, error) {
	rows, err := h.Accounts.DB.SQL().
		QueryContext(ctx, `SELECT o.id::text, o.store, s.name, o.country::text, o.kind, o.source, o.url,
  o.verify_status, o.active
FROM offers o JOIN stores s ON s.code = o.store WHERE o.item_id = $1
ORDER BY o.active DESC, o.country, s.name, o.kind`, itemID)
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	var out []offerRow

	for rows.Next() {
		var r offerRow
		if err := rows.Scan(
			&r.ID,
			&r.Store,
			&r.StoreName,
			&r.Country,
			&r.Kind,
			&r.Source,
			&r.URL,
			&r.Status,
			&r.Active,
		); err != nil {
			return nil, err
		}

		out = append(out, r)
	}

	return out, rows.Err()
}

func (h CatalogHandlers) storeOptions(ctx context.Context) ([]storeOption, error) {
	rows, err := h.Accounts.DB.SQL().QueryContext(ctx, `SELECT code, name FROM stores WHERE active ORDER BY name, code`)
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	var out []storeOption

	for rows.Next() {
		var s storeOption
		if err := rows.Scan(&s.Code, &s.Name); err != nil {
			return nil, err
		}

		out = append(out, s)
	}

	return out, rows.Err()
}

func auditOffer(ctx context.Context, ex database.Executor, actor models.User, record, action string,
	changes map[string]vadmin.Change,
) error {
	return orm.Create(ctx, ex, &vadmin.AuditLog{
		ActorID: actor.AuthID(), Resource: offersResource, RecordID: record, Action: action,
		Changes: orm.JSON[map[string]vadmin.Change]{V: changes},
	})
}

// SaveOffer handles POST /admin/items/{id}/offers: add an offer by hand. The row has source admin, so the stages never
// overwrite it; an offer with the same country, store and kind is replaced and becomes the admin's. Audited.
func (h CatalogHandlers) SaveOffer(c *web.Ctx) error {
	actor, err := h.actor(c)
	if err != nil {
		return err
	}

	id, err := itemID(c)
	if err != nil {
		return err
	}

	store, country, kind := c.FormValue(
		"store",
	), strings.ToUpper(
		strings.TrimSpace(c.FormValue("country")),
	), c.FormValue(
		"kind",
	)
	link := strings.TrimSpace(c.FormValue("url"))

	v := errs.NewValidation()
	if country != offers.Everywhere && !slices.Contains(offers.Countries, country) {
		v.Append("country", "Choose one of the listed countries or ZZ for everywhere.")
	}

	if !slices.Contains(offerKinds, kind) {
		v.Append("kind", "Unknown kind.")
	}

	if !validAffiliateURL(link) {
		v.Append("url", "The URL must be an https link of at most 2048 characters.")
	}

	if !v.Empty() {
		return v
	}

	err = h.Accounts.DB.Tx(c.Context(), func(tx *database.Tx) error {
		var ok bool
		if err := tx.QueryRowContext(c.Context(), `SELECT true FROM stores WHERE code = $1 AND active`, store).
			Scan(&ok); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errs.NewValidation().Add("store", "Unknown store.")
			}

			return err
		}

		var item bool
		if err := tx.QueryRowContext(c.Context(), `SELECT true FROM items WHERE id = $1`, id).Scan(&item); err != nil {
			return err
		}

		var oldURL sql.NullString

		err := tx.QueryRowContext(c.Context(), `INSERT INTO offers (item_id, country, store, kind, url, source, licence, confidence)
VALUES ($1, $2, $3, $4, $5, 'admin', 'proprietary', 100)
ON CONFLICT ON CONSTRAINT offers_item_id_edition_id_country_store_kind_key DO UPDATE SET url = EXCLUDED.url,
  source = 'admin', licence = 'proprietary', confidence = 100, external_id = NULL, expires_at = NULL, active = true,
  verify_status = 'unverified', broken_runs = 0
RETURNING (SELECT url FROM offers WHERE item_id = $1 AND edition_id IS NULL AND country = $2 AND store = $3 AND kind = $4)`,
			id, country, store, kind, link).
			Scan(&oldURL)
		if err != nil {
			return err
		}

		action, before := vadmin.ActionCreate, any(nil)
		if oldURL.Valid {
			action, before = vadmin.ActionUpdate, oldURL.String
		}

		return auditOffer(c.Context(), tx, actor, id+"/"+country+"/"+store+"/"+kind, action, map[string]vadmin.Change{
			"url": {Before: before, After: link},
		})
	})
	if err != nil {
		if verr, ok := errors.AsType[*errs.ValidationError](err); ok {
			return verr
		}

		return mapTrendingErr(err)
	}

	return c.Redirect("/admin/items/" + id + "/manage")
}

// ToggleOffer handles POST /admin/items/{id}/offers/{offer}/toggle: deactivate an active offer or restore a deactivated
// one. An offer of another item is a 404. Audited.
func (h CatalogHandlers) ToggleOffer(c *web.Ctx) error {
	actor, err := h.actor(c)
	if err != nil {
		return err
	}

	id, err := itemID(c)
	if err != nil {
		return err
	}

	oid := c.Param("offer")
	if _, err := uuid.Parse(oid); err != nil {
		return errs.ErrNotFound
	}

	err = h.Accounts.DB.Tx(c.Context(), func(tx *database.Tx) error {
		var active bool
		if err := tx.QueryRowContext(c.Context(), `UPDATE offers SET active = NOT active,
  broken_runs = CASE WHEN NOT active THEN 0 ELSE broken_runs END,
  verify_status = CASE WHEN NOT active AND verify_status = 'gone' THEN 'unverified' ELSE verify_status END
WHERE id = $1 AND item_id = $2 RETURNING active`, oid, id).Scan(&active); err != nil {
			return err
		}

		return auditOffer(c.Context(), tx, actor, oid, vadmin.ActionUpdate, map[string]vadmin.Change{
			"active": {Before: !active, After: active},
		})
	})
	if err != nil {
		return mapTrendingErr(err)
	}

	return c.Redirect("/admin/items/" + id + "/manage")
}
