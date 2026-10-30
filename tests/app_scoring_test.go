//go:build integration

package tests

import (
	"strconv"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// TestAppScoring proves the shipped /app/v1 scoring writes (SDD 02, SDD 10).
// A repeated rating PUT with the same Idempotency-Key does not store a second
// row, and a token without ratings:write is forbidden.
func TestAppScoring(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{
		"DATABASE_URL": pgtest.Database(t, ""),
	}})
	writer := makeUser(t, app, "score-writer@example.com", "user", true)
	reader := makeUser(t, app, "score-reader@example.com", "user", true)
	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	writeToken, _, err := acc.Tokens.Create(t.Context(), writer.AuthID(), "phone", []string{"ratings:write"}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	readToken, _, err := acc.Tokens.Create(t.Context(), reader.AuthID(), "phone", []string{"status:write"}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}
	item := uuid.NewV7().String()
	if _, err := db.SQL().ExecContext(t.Context(), `INSERT INTO items (id, type, slug, wikidata_qid, release_date, release_precision, release_status)
VALUES ($1, 'game', $2, $3, '2020-01-01', 'day', 'released')`, item, "s-"+item, "Q"+strconv.FormatInt(time.Now().UnixNano(), 10)); err != nil {
		t.Fatal(err)
	}

	path := "/app/v1/items/" + item + "/rating"
	body := map[string]any{"scores": map[string]float64{
		"gameplay": 8, "fun": 8, "sound": 8, "technical": 8, "visual": 8, "storytelling": 8,
	}}

	app.Client(t).WithToken(readToken).PutJSON(path, body).AssertForbidden()
	if n := ratingRows(t, app, item); n != 0 {
		t.Fatalf("forbidden put stored %d ratings", n)
	}

	c := app.Client(t).WithToken(writeToken).WithHeader("Idempotency-Key", "rating-once")
	c.PutJSON(path, body).AssertOK()
	replay := c.PutJSON(path, body).AssertOK()
	if replay.Header("Idempotent-Replayed") != "true" {
		t.Fatal("second put was not replayed")
	}
	if n := ratingRows(t, app, item); n != 1 {
		t.Fatalf("ratings = %d, want 1", n)
	}

	var created, updated int
	if err := db.SQL().QueryRowContext(t.Context(), `SELECT
  count(*) FILTER (WHERE kind = 'rating_created'),
  count(*) FILTER (WHERE kind = 'rating_updated')
FROM rating_events WHERE item_id = $1`, item).Scan(&created, &updated); err != nil {
		t.Fatal(err)
	}
	if created != 1 || updated != 0 {
		t.Fatalf("rating events created=%d updated=%d, want 1 and 0", created, updated)
	}

	app.Client(t).WithToken(writeToken).GetJSON("/app/v1/items/"+item+"/me").
		AssertOK().
		AssertJSONPath("rating.gameplay", 8.0)
}

func ratingRows(t *testing.T, app *vukemtest.App, item string) int {
	t.Helper()

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.SQL().QueryRowContext(t.Context(), `SELECT count(*) FROM ratings WHERE item_id = $1`, item).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}
