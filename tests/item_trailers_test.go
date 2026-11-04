//go:build integration

package tests

import "testing"

// Production 2026-11-04: the normaliser stores items.trailers as a list of YouTube ids and the item page read only
// objects, so every item with a trailer from Wikidata (Grand Theft Auto VI) answered 500.
func TestItemPageReadsTrailersStoredAsIds(t *testing.T) {
	r := newListRig(t)
	id := r.item("game", "public")

	db, err := r.app.DB()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(t.Context(),
		`UPDATE items SET trailers = '["QdBZY2fkU-0","VQRLujxTm3c"]'::jsonb WHERE id = $1`, id); err != nil {
		t.Fatal(err)
	}

	r.app.Client(t).Inertia("").Get(r.itemPath(id)).AssertOK().AssertInertia("Catalog/Show").
		HasProp("trailers", 2).
		PropEquals("trailers.0.youtubeId", "QdBZY2fkU-0")
}
