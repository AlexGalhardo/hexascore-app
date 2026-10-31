//go:build integration

package scoring_test

import (
	"fmt"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

var qids atomic.Int64

type rig struct {
	t   *testing.T
	app *vukemtest.App
	acc *accounts.Accounts
}

func newRig(t *testing.T) *rig {
	t.Helper()
	url := pgtest.Database(t, "")
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{"DATABASE_URL": url}})
	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}
	return &rig{t: t, app: app, acc: acc}
}

// token creates a user and a token with the given abilities.
func (r *rig) token(n int, abilities ...string) *vukemtest.Client {
	return r.tokenAged(n, 200, abilities...)
}

// tokenAged is token for an account created ageDays days ago.
func (r *rig) tokenAged(n, ageDays int, abilities ...string) *vukemtest.Client {
	r.t.Helper()
	db, err := r.app.DB()
	if err != nil {
		r.t.Fatal(err)
	}
	var id int64
	// email_verified_at is required by the /app/v1 guard (auth.Verified). Account age stays created_at.
	if err := db.SQL().
		QueryRowContext(r.t.Context(), `INSERT INTO users (name, email, password, email_verified_at, created_at, updated_at)
VALUES ('U', $1, 'x', now(), now() - $2 * interval '1 day', now()) RETURNING id`, fmt.Sprintf("u%d@example.com", n), ageDays).
		Scan(&id); err != nil {
		r.t.Fatal(err)
	}
	plain, _, err := r.acc.Tokens.Create(r.t.Context(), strconv.FormatInt(id, 10), "t", abilities, time.Hour)
	if err != nil {
		r.t.Fatal(err)
	}
	return r.app.Client(r.t).WithToken(plain)
}

func (r *rig) item(typ, release, precision, status string) string {
	r.t.Helper()
	db, err := r.app.DB()
	if err != nil {
		r.t.Fatal(err)
	}
	id := uuid.NewV7().String()
	var rel any
	if release != "" {
		rel = release
	}
	if _, err := db.SQL().
		ExecContext(r.t.Context(), `INSERT INTO items (id, type, slug, wikidata_qid, release_date, release_precision, release_status)
VALUES ($1, $2, $3, $4, $5, $6, $7)`, id, typ, "s-"+id, fmt.Sprintf("Q%d", qids.Add(1)), rel, precision, status); err != nil {
		r.t.Fatal(err)
	}
	return id
}

var gameScores = map[string]any{"scores": map[string]float64{
	"gameplay": 8, "fun": 8, "sound": 8, "technical": 8, "visual": 8, "storytelling": 8,
}}

func TestThreshold(t *testing.T) {
	r := newRig(t)
	item := r.item("game", "2020-01-01", "day", "released")
	path := "/app/v1/items/" + item + "/rating"

	for i := 1; i <= 5; i++ {
		resp := r.token(i, "ratings:write").PutJSON(path, gameScores).AssertOK()
		resp.AssertJSONPath("item.ratingCount", float64(i))
		if i < 5 {
			resp.AssertJSONPath("item.score", nil)
		} else if v, _ := resp.JSONPath("item.score"); v == nil {
			t.Fatal("5th rating: score is null")
		}
	}
}

func TestReleaseGate(t *testing.T) {
	r := newRig(t)
	c := r.token(1, "ratings:write")
	tomorrow := time.Now().UTC().AddDate(0, 0, 2).Format("2006-01-02")
	thisYear := fmt.Sprintf("%d-01-01", time.Now().UTC().Year())

	cases := []struct {
		name string
		item string
		want int
	}{
		{"future day", r.item("game", tomorrow, "day", "released"), 409},
		{"year precision before 31 Dec", r.item("game", thisYear, "year", "released"), 409},
		{"no date", r.item("game", "", "unknown", "released"), 200},
		{"announced without date", r.item("game", "", "unknown", "announced"), 409},
		{"past day", r.item("game", "2020-01-01", "day", "released"), 200},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(*testing.T) {
			c.PutJSON("/app/v1/items/"+tc.item+"/rating", gameScores).AssertStatus(tc.want)
		})
	}
}

func TestRatingWriteValidationEditDelete(t *testing.T) {
	r := newRig(t)
	c := r.token(1, "ratings:write")
	item := r.item("game", "2020-01-01", "day", "released")
	path := "/app/v1/items/" + item + "/rating"

	c.PutJSON(path, map[string]any{"scores": map[string]float64{"gameplay": 8.25}}).AssertUnprocessable()
	c.PutJSON(path, map[string]any{"scores": map[string]float64{"gameplay": 8}}).AssertUnprocessable()
	c.PutJSON("/app/v1/items/"+uuid.NewV7().String()+"/rating", gameScores).AssertNotFound()
	c.PutJSON("/app/v1/items/not-a-uuid/rating", gameScores).AssertNotFound()

	c.PutJSON(path, gameScores).AssertOK()
	c.PutJSON(path, gameScores).AssertOK() // edit
	vukemtest.AssertDatabaseHas(t, r.app, "ratings", map[string]any{"item_id": item, "c1": 80})
	vukemtest.AssertDatabaseHas(t, r.app, "rating_events", map[string]any{"item_id": item, "kind": "rating_updated"})

	c.Delete(path).AssertNoContent()
	vukemtest.AssertDatabaseMissing(t, r.app, "ratings", map[string]any{"item_id": item})
	vukemtest.AssertDatabaseHas(t, r.app, "rating_events", map[string]any{"item_id": item, "kind": "rating_deleted"})
}

func TestRatingWebRedirectsBack(t *testing.T) {
	r := newRig(t)
	db, err := r.app.DB()
	if err != nil {
		t.Fatal(err)
	}
	var id int64
	if err := db.SQL().
		QueryRowContext(t.Context(), `INSERT INTO users (name, email, password, email_verified_at, created_at, updated_at)
VALUES ('W', 'w@example.com', 'x', now(), now(), now()) RETURNING id`).
		Scan(&id); err != nil {
		t.Fatal(err)
	}
	item := r.item("game", "2020-01-01", "day", "released")
	c := r.app.Client(t).WithHeader("Sec-Fetch-Site", "same-origin").WithHeader("Referer", "/games/x").
		ActingAs("web", authID(id))
	c.PutJSON("/items/"+item+"/rating", gameScores).AssertRedirect("/games/x")
	vukemtest.AssertDatabaseHas(t, r.app, "ratings", map[string]any{"item_id": item})
}

type authID int64

func (a authID) AuthID() string { return strconv.FormatInt(int64(a), 10) }
