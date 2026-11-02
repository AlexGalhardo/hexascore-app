//go:build integration

package tests

import (
	"strings"
	"testing"
)

// A field whose provenance v2 confidence is 50 never reaches /api/v1; a confidence of 60 does.
func TestFieldReviews_APIOmitsLowConfidenceField(t *testing.T) {
	r := newAPICatalog(t)

	set := func(conf int) {
		t.Helper()

		if _, err := r.acc.DB.ExecContext(t.Context(), `UPDATE items SET provenance_v2 =
  jsonb_build_object('release_date', jsonb_build_object('source', 'wikidata', 'confidence', $2::int)) WHERE id = $1`,
			r.game, conf); err != nil {
			t.Fatal(err)
		}
	}

	set(50)

	res := r.client.GetJSON("/api/v1/games/" + r.game).AssertOK()
	if body := res.Body(); strings.Contains(body, "2017-02-24") || !strings.Contains(body, "releaseDate") {
		t.Fatalf("low-confidence release date exposed, or not listed as omitted: %s", body)
	}

	set(60)
	r.client.GetJSON("/api/v1/games/"+r.game).AssertOK().AssertJSONPath("releaseDate", "2017-02-24")
}
