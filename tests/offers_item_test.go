//go:build integration

package tests

import (
	"testing"

	"github.com/AlexGalhardo/vukem/uuid"
)

// TestItemPageCarriesOffers: the page gets every country's offers (so the cached HTML is the same for all visitors),
// the default country follows the locale, and the target URL never reaches the props.
func TestItemPageCarriesOffers(t *testing.T) {
	r := newListRig(t)
	id := r.item("game", "public")

	db, err := r.app.DB()
	if err != nil {
		t.Fatal(err)
	}

	offerID := uuid.NewV7().String()
	if _, err := db.ExecContext(
		t.Context(),
		`INSERT INTO offers (id, item_id, country, store, kind, url, source, licence)
VALUES ($1, $2, 'ZZ', 'steam', 'buy_digital', 'https://store.steampowered.com/app/1', 'wikidata', 'CC0-1.0')`,
		offerID,
		id,
	); err != nil {
		t.Fatal(err)
	}

	r.app.Client(t).Inertia("").Get(r.itemPath(id)).AssertOK().AssertInertia("Catalog/Show").
		PropEquals("offers.defaultCountry", "US").
		HasProp("offers.byCountry.US", 1).
		HasProp("offers.byCountry.BR", 1).
		PropEquals("offers.byCountry.BR.0.name", "Steam").
		PropEquals("offers.byCountry.BR.0.id", offerID).
		PropEquals("offers.commission", false).
		MissingProp("offers.byCountry.BR.0.url")

	r.app.Client(t).Inertia("").Get("/pt-br"+r.itemPath(id)).AssertOK().AssertInertia("Catalog/Show").
		PropEquals("offers.defaultCountry", "BR")
}

// An item nobody sells still gets the section, empty, so the page can say so.
func TestItemPageOffersEmpty(t *testing.T) {
	r := newListRig(t)
	id := r.item("game", "public")

	r.app.Client(t).Inertia("").Get(r.itemPath(id)).AssertOK().AssertInertia("Catalog/Show").
		HasProp("offers.byCountry.US", 0).
		HasProp("offers.countries", 6)
}
