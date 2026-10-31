//go:build integration

package tests

import (
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func TestPublicItemRatings(t *testing.T) {
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{
		"DATABASE_URL": pgtest.Database(t, ""),
	}})
	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}
	ctx := t.Context()
	exec := func(q string, args ...any) {
		t.Helper()
		if _, err := db.ExecContext(ctx, q, args...); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
	insertUser := func(username string, private, banned, deleting bool) int64 {
		t.Helper()
		var id int64
		email := username + "@example.test"
		name := username
		if username == "" {
			email, name = "hidden-name@example.test", "noname"
		}
		var bannedAt, deletingAt any
		if banned {
			bannedAt = time.Now().UTC()
		}
		if deleting {
			deletingAt = time.Now().UTC()
		}
		err := db.QueryRowContext(ctx, `INSERT INTO users
(name, email, password, username, profile_private, banned_at, deletion_scheduled_at, created_at, updated_at)
VALUES ($1, $2, 'x', NULLIF($3, ''), $4, $5, $6, now(), now()) RETURNING id`,
			name, email, username, private, bannedAt, deletingAt).Scan(&id)
		if err != nil {
			t.Fatal(err)
		}
		return id
	}

	exec(`INSERT INTO data_sources (code, name, licence, url) VALUES ('hexascore', 'x', 'proprietary', 'https://x')`)

	itemID := uuid.NewV7().String()
	exec(`INSERT INTO items (id, type, slug, wikidata_qid, trailers)
VALUES ($1, 'game', 'rated-public', 'Q9100', $2::jsonb)`,
		itemID, `[{"youtubeId":"aaaaaaaaaaa","title":"Demo trailer"},{"youtubeId":"bad","title":"no"}]`)
	exec(`INSERT INTO item_translations (item_id, locale, title, title_source)
VALUES ($1, 'en', 'Rated', 'hexascore')`, itemID)

	base := time.Now().UTC()
	names := []string{
		"alice",
		"bobby",
		"carol",
		"david",
		"erika",
		"frank",
		"grace",
		"henry",
		"irene",
		"james",
		"karen",
		"linda",
	}
	for i, name := range names {
		id := insertUser(name, false, false, false)
		exec(`INSERT INTO ratings (user_id, item_id, c1, c2, c3, c4, c5, c6, created_at)
VALUES ($1, $2, 80, 80, 80, 80, 80, 80, $3)`, id, itemID, base.Add(time.Duration(i)*time.Second))
	}
	exec(`INSERT INTO ratings (user_id, item_id, c1, c2, c3, c4, c5, c6, created_at)
VALUES ($1, $2, 50, 50, 50, 50, 50, 50, $3)`, insertUser("quinn", true, false, false), itemID, base.Add(time.Hour))
	exec(`INSERT INTO ratings (user_id, item_id, c1, c2, c3, c4, c5, c6) VALUES ($1, $2, 50, 50, 50, 50, 50, 50)`,
		insertUser("robin", false, true, false), itemID)
	exec(`INSERT INTO ratings (user_id, item_id, c1, c2, c3, c4, c5, c6) VALUES ($1, $2, 50, 50, 50, 50, 50, 50)`,
		insertUser("susan", false, false, true), itemID)
	exec(`INSERT INTO ratings (user_id, item_id, c1, c2, c3, c4, c5, c6) VALUES ($1, $2, 50, 50, 50, 50, 50, 50)`,
		insertUser("", false, false, false), itemID)
	var expiredID int64
	if err := db.QueryRowContext(ctx, `INSERT INTO users
(name, email, password, username, profile_private, banned_at, ban_expires_at, created_at, updated_at)
VALUES ('tanya', 'tanya@example.test', 'x', 'tanya', false, now(), now() - interval '1 minute', now(), now())
RETURNING id`).Scan(&expiredID); err != nil {
		t.Fatal(err)
	}
	exec(`INSERT INTO ratings (user_id, item_id, c1, c2, c3, c4, c5, c6, created_at)
VALUES ($1, $2, 80, 80, 80, 80, 80, 80, $3)`, expiredID, itemID, base.Add(-time.Second))

	hidden := uuid.NewV7().String()
	exec(`INSERT INTO items (id, type, slug, wikidata_qid, visibility)
VALUES ($1, 'game', 'rated-hidden', 'Q9101', 'hidden')`, hidden)

	client := app.Client(t)
	var page1 struct {
		Data []struct {
			Nickname   string    `json:"nickname"`
			Overall    float64   `json:"overall"`
			Categories []float64 `json:"categories"`
			RatedAt    string    `json:"ratedAt"`
		} `json:"data"`
		Page      int `json:"page"`
		PageCount int `json:"page_count"`
	}
	client.GetJSON("/api/items/" + itemID + "/ratings").AssertOK().JSON(&page1)
	if page1.Page != 1 || page1.PageCount != 2 || len(page1.Data) != 10 {
		t.Fatalf("page1 page=%d count=%d rows=%d", page1.Page, page1.PageCount, len(page1.Data))
	}
	if page1.Data[0].Nickname != "linda" || page1.Data[0].Overall != 8 || len(page1.Data[0].Categories) != 6 ||
		page1.Data[0].Categories[0] != 8 {
		t.Fatalf("newest row %+v", page1.Data[0])
	}
	if _, err := time.Parse(time.RFC3339, page1.Data[0].RatedAt); err != nil {
		t.Fatal(err)
	}
	for _, row := range page1.Data {
		switch row.Nickname {
		case "quinn", "robin", "susan", "":
			t.Fatalf("restricted profile listed: %s", row.Nickname)
		}
	}

	var page2 struct {
		Data []struct {
			Nickname string `json:"nickname"`
		} `json:"data"`
		Page int `json:"page"`
	}
	client.GetJSON("/api/items/" + itemID + "/ratings?page=2").AssertOK().JSON(&page2)
	if page2.Page != 2 || len(page2.Data) != 3 || page2.Data[0].Nickname != "bobby" ||
		page2.Data[1].Nickname != "alice" ||
		page2.Data[2].Nickname != "tanya" {
		t.Fatalf("page2 %+v", page2)
	}

	var past struct {
		Data []any `json:"data"`
		Page int   `json:"page"`
	}
	client.GetJSON("/api/items/" + itemID + "/ratings?page=5").AssertOK().JSON(&past)
	if past.Page != 5 || past.Data == nil || len(past.Data) != 0 {
		t.Fatalf("past page=%d data=%#v", past.Page, past.Data)
	}

	client.GetJSON("/api/items/not-a-uuid/ratings").AssertStatus(404)
	client.GetJSON("/api/items/" + uuid.NewV7().String() + "/ratings").AssertStatus(404)
	client.GetJSON("/api/items/" + hidden + "/ratings").AssertStatus(404)
	for _, raw := range []string{"0", "101", "foo", "1.5"} {
		client.GetJSON("/api/items/" + itemID + "/ratings?page=" + raw).AssertStatus(400)
	}

	client.Inertia("").Get("/games/rated-public").AssertOK().AssertInertia("Catalog/Show").
		PropEquals("trailers.0.youtubeId", "aaaaaaaaaaa").
		PropEquals("trailers.0.title", "Demo trailer").
		MissingProp("trailers.1")
}
