//go:build integration

package tests

import (
	"net/http"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/uuid"
)

func seedTrending(t *testing.T, r *keyRig) (item string) {
	t.Helper()

	item = uuid.NewV7().String()
	day := time.Now().UTC().Format(time.DateOnly)

	if _, err := r.acc.DB.ExecContext(t.Context(), `INSERT INTO data_sources (code, name, licence, url)
VALUES ('wikidata', 'Wikidata', 'CC0-1.0', 'https://www.wikidata.org')`); err != nil {
		t.Fatal(err)
	}

	for _, q := range []string{
		`INSERT INTO items (id, type, slug, wikidata_qid, score, rating_count) VALUES ($1, 'game', 'silksong', 'Q1', 80, 3)`,
		`INSERT INTO item_translations (item_id, locale, title, title_source) VALUES ($1, 'en', 'Silksong', 'wikidata')`,
		`INSERT INTO trending_snapshots (computed_at, item_type, rank, item_id, trend_score, basis)
VALUES (date_trunc('hour', now()), 'game', 1, $1, 118.97, 'trend')`,
		`INSERT INTO item_activity_daily (item_id, day, item_type, views, search_clicks, ratings_new, ratings_new_w,
  list_adds, list_adds_w) VALUES ($1, '` + day + `', 'game', 40, 5, 3, 2.5, 2, 1.5)`,
	} {
		if _, err := r.acc.DB.ExecContext(t.Context(), q, item); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}

	return item
}

func TestAPITrending(t *testing.T) {
	r, client := webhookClient(t, "trend@example.com", "")
	item := seedTrending(t, r)

	res := client.GetJSON("/api/v1/games/trending").AssertOK().AssertMatchesOpenAPI().
		AssertJSONPath("data.0.rank", float64(1)).AssertJSONPath("data.0.basis", "trend").
		AssertJSONPath("data.0.item.id", item).AssertDontSee("window")
	res.AssertHeader("Cache-Control", "public, max-age=300")

	etag := res.Header("ETag")
	if etag == "" {
		t.Fatal("no ETag")
	}

	client.WithHeader("If-None-Match", etag).GetJSON("/api/v1/games/trending").AssertStatus(http.StatusNotModified)
	client.GetJSON("/api/v1/movies/trending").AssertOK().AssertJSONCount("data", 0)
	client.GetJSON("/api/v1/games/trending?window=day").AssertStatus(http.StatusBadRequest)
}

func TestAPIActivityPlan(t *testing.T) {
	r, free := webhookClient(t, "act-free@example.com", "")
	item := seedTrending(t, r)
	path := "/api/v1/games/" + item + "/activity"

	free.GetJSON(path).AssertForbidden().AssertJSONPath("code", "plan-required")

	_, pro := webhookClient(t, "act-pro@example.com", "pro")
	pro.GetJSON(path).AssertForbidden().AssertJSONPath("code", "plan-required")

	r2, biz := webhookClient(t, "act-biz@example.com", "business")
	item2 := seedTrending(t, r2)
	path = "/api/v1/games/" + item2 + "/activity"

	before := quotaUsed(t, biz)
	biz.GetJSON(path).AssertOK().AssertMatchesOpenAPI().AssertJSONPath("data.0.views", float64(40)).
		AssertJSONPath("data.0.listAdds", float64(2)).AssertDontSee("_w").AssertDontSee("Weight")

	if after := quotaUsed(t, biz); after != before+2 {
		t.Fatalf("quota used %d -> %d, want +2 (the activity read and the usage read)", before, after)
	}

	biz.GetJSON(path + "?from=2020-01-01&to=2020-02-01").AssertUnprocessable()
	biz.GetJSON(path + "?from=" + time.Now().AddDate(0, 0, -100).Format(time.DateOnly)).AssertUnprocessable()
	biz.GetJSON("/api/v1/games/" + uuid.NewV7().String() + "/activity").AssertNotFound()
}
