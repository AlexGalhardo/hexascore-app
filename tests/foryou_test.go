//go:build integration

package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/recommend/foryou"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

type foryouRig struct {
	t      *testing.T
	app    *vukemtest.App
	acc    *accounts.Accounts
	svc    *foryou.Service
	user   int64
	liked  []string
	cands  []string
	queued []foryou.RecommendUser
	n      int
}

// newForYouRig seeds a user who rated three games 100, all in one franchise, and two unrated games of that franchise.
func newForYouRig(t *testing.T) *foryouRig {
	t.Helper()

	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}})
	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	r := &foryouRig{t: t, app: app, acc: acc, user: makeUser(t, app, "fy@example.com", "user", true).ID}
	r.svc = &foryou.Service{DB: acc.DB, Now: app.Now, Enqueue: func(_ context.Context, j foryou.RecommendUser) error {
		r.queued = append(r.queued, j)

		return nil
	}}

	r.exec(`INSERT INTO data_sources (code, name, licence, url) VALUES ('hexascore', 'x', 'proprietary', 'https://x'), ('wikidata', 'W', 'CC0-1.0', 'https://www.wikidata.org')`)

	franchise := uuid.NewV7().String()
	r.exec(`INSERT INTO entities (id, kind, slug, default_name) VALUES ($1, 'franchise', 'f', 'F')`, franchise)

	for range 5 {
		r.item("public") // filler keeps N_t above the franchise size
	}

	for range 3 {
		id := r.item("public")
		r.exec(`INSERT INTO item_entities (item_id, entity_id, role, source) VALUES ($1, $2, 'franchise', 'hexascore')`, id, franchise)
		r.exec(`INSERT INTO ratings (user_id, item_id, c1, c2, c3, c4, c5, c6) VALUES ($1, $2, 100, 100, 100, 100, 100, 100)`,
			r.user, id)
		r.liked = append(r.liked, id)
	}

	for range 2 {
		id := r.item("public")
		r.exec(`INSERT INTO item_entities (item_id, entity_id, role, source) VALUES ($1, $2, 'franchise', 'hexascore')`, id, franchise)
		r.cands = append(r.cands, id)
	}

	return r
}

func (r *foryouRig) exec(q string, args ...any) {
	r.t.Helper()

	if _, err := r.acc.DB.ExecContext(r.t.Context(), q, args...); err != nil {
		r.t.Fatalf("%v\n%s", err, q)
	}
}

func (r *foryouRig) item(visibility string) string {
	r.n++
	id := uuid.NewV7().String()
	r.exec(`INSERT INTO items (id, type, slug, wikidata_qid, visibility) VALUES ($1, 'game', $2, $3, $4)`,
		id, fmt.Sprintf("fy-%d", r.n), fmt.Sprintf("Q%d", 9000+r.n), visibility)
	r.exec(`INSERT INTO item_translations (item_id, locale, title, title_source) VALUES ($1, 'en', $2, 'wikidata')`,
		id, fmt.Sprintf("Game %d", r.n))

	return id
}

func (r *foryouRig) rows() int {
	var n int

	if err := r.acc.DB.QueryRowContext(r.t.Context(),
		`SELECT count(*) FROM user_recommendations WHERE user_id = $1`, r.user).Scan(&n); err != nil {
		r.t.Fatal(err)
	}

	return n
}

func TestForYouOptOut(t *testing.T) {
	t.Parallel()

	r := newForYouRig(t)

	l, err := r.svc.Get(t.Context(), r.user, "game", "en")
	if err != nil || l.Basis != foryou.BasisPersonal || len(l.Data) != 2 || r.rows() != 1 {
		t.Fatalf("before opt-out: %+v rows=%d err=%v", l, r.rows(), err)
	}

	if l.Data[0].Reason == nil || l.Data[0].Reason.Kind != "rated" {
		t.Fatalf("reason = %+v", l.Data[0].Reason)
	}

	if err := r.acc.UpdatePrivacy(t.Context(), r.user, false, true); err != nil {
		t.Fatal(err)
	}

	if r.rows() != 0 {
		t.Fatal("opting out must delete user_recommendations")
	}

	l, err = r.svc.Get(t.Context(), r.user, "game", "en")
	if err != nil || len(l.Data) != 0 || r.rows() != 0 {
		t.Fatalf("after opt-out: %+v rows=%d err=%v", l, r.rows(), err)
	}
}

func TestRecommendJobExitsWhenOptedOut(t *testing.T) {
	t.Parallel()

	r := newForYouRig(t)
	r.exec(`UPDATE users SET personalization_opt_out = true WHERE id = $1`, r.user)

	if err := r.svc.Run(t.Context(), foryou.RecommendUser{UserID: r.user, Type: "game"}); err != nil {
		t.Fatal(err)
	}

	if r.rows() != 0 {
		t.Fatal("a queued job for an opted-out user must not write")
	}
}

func TestForYouStaleWhileRefresh(t *testing.T) {
	t.Parallel()

	r := newForYouRig(t)

	if _, err := r.svc.Get(t.Context(), r.user, "game", "en"); err != nil {
		t.Fatal(err)
	}

	if len(r.queued) != 0 {
		t.Fatalf("a fresh compute must not queue a refresh: %v", r.queued)
	}

	if _, err := r.svc.Get(t.Context(), r.user, "game", "en"); err != nil || len(r.queued) != 0 {
		t.Fatalf("a fresh hit must not queue: %v %v", r.queued, err)
	}

	for range foryou.StaleEvents {
		r.exec(`INSERT INTO rating_events (kind, user_id, item_id, item_type) VALUES ('rating_created', $1, $2, 'game')`,
			r.user, r.liked[0])
	}

	l, err := r.svc.Get(t.Context(), r.user, "game", "en")
	if err != nil || len(l.Data) != 2 || len(r.queued) != 1 {
		t.Fatalf("5 newer events: served %d, queued %v, err %v", len(l.Data), r.queued, err)
	}

	r.queued = nil
	r.exec(`UPDATE user_recommendations SET computed_at = $1, events_at_compute = 99 WHERE user_id = $2`,
		time.Now().Add(-25*time.Hour), r.user)

	if _, err := r.svc.Get(t.Context(), r.user, "game", "en"); err != nil || len(r.queued) != 1 {
		t.Fatalf("a row older than 24 h must queue a refresh: %v %v", r.queued, err)
	}
}

func TestForYouEndpoint(t *testing.T) {
	t.Parallel()

	r := newForYouRig(t)

	u, err := r.acc.Users.ByID(t.Context(), fmt.Sprint(r.user))
	if err != nil {
		t.Fatal(err)
	}

	tok, _, err := r.acc.Tokens.Create(t.Context(), u.AuthID(), "phone", []string{"catalog:read"}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	r.app.Client(t).GetJSON("/app/v1/me/for-you?type=game").AssertUnauthorized()
	r.app.Client(t).WithToken(tok).GetJSON("/app/v1/me/for-you?type=nope").AssertUnprocessable()
	r.app.Client(t).WithToken(tok).GetJSON("/app/v1/me/for-you?type=game").AssertOK().AssertSee(`"basis":"personal"`)
}
