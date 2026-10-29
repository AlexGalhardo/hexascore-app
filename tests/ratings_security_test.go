package tests

import (
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// TestSecurity_A01_RatingNeedsAbility: a token without ratings:write cannot
// write or delete ratings (SDD 02 §6).
func TestSecurity_A01_RatingNeedsAbility(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)
	u := makeUser(t, app, "rater@example.com", "user", true)
	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	plain, _, err := acc.Tokens.Create(t.Context(), u.AuthID(), "ro", []string{"status:write"}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	const path = "/app/v1/items/0197f3a0-0000-7000-8000-000000000000/rating"
	body := map[string]any{"scores": map[string]float64{"gameplay": 7.5}}
	app.Client(t).WithToken(plain).PutJSON(path, body).AssertForbidden()
	app.Client(t).WithToken(plain).Delete(path).AssertForbidden()
	app.Client(t).PutJSON(path, body).AssertUnauthorized()
}

// TestRatingItemLimitedProblemCode: the item-limited 429 carries the kebab
// `code` and Retry-After (SDD 05 05-9).
func TestRatingItemLimitedProblemCode(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}})
	u := makeUser(t, app, "limited@example.com", "user", true)

	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	plain, _, err := acc.Tokens.Create(t.Context(), u.AuthID(), "rw", []string{"ratings:write"}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	const item = "0197f3a0-0000-7000-8000-000000000002"

	q := `INSERT INTO items (id, type, slug, wikidata_qid, ratings_limited_until)
		VALUES ('` + item + `', 'game', 'limited', 'Q7', now() + interval '1 day')`
	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := db.SQL().ExecContext(t.Context(), q); err != nil {
		t.Fatal(err)
	}

	app.Client(t).WithToken(plain).
		PutJSON("/app/v1/items/"+item+"/rating", map[string]any{"scores": map[string]float64{"gameplay": 7.5, "fun": 7, "sound": 7, "storytelling": 7, "technical": 7, "visual": 7}}).
		AssertStatus(429).
		AssertSee(`"code":"item-ratings-limited"`).
		AssertHeaderContains("Retry-After", "")
}
