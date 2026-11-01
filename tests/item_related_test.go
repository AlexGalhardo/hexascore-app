//go:build integration

package tests

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

type relatedRig struct {
	t       *testing.T
	app     *vukemtest.App
	acc     *accounts.Accounts
	subject string
	hidden  string
	n       int
}

// newRelatedRig seeds a game "Zelda" with 9 stored neighbours (one hidden, one sharing the franchise "Saga", the rest
// sharing only a genre), a book adapted from it (item_links inbound) and a movie it is based on (outbound).
func newRelatedRig(t *testing.T) (*relatedRig, []string) {
	t.Helper()

	app := vukemtest.New(
		t,
		bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}},
	)
	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	r := &relatedRig{t: t, app: app, acc: acc}
	r.exec(`INSERT INTO data_sources (code, name, licence, url) VALUES ('hexascore', 'x', 'proprietary', 'https://x')`)

	franchise, genre := uuid.NewV7().String(), uuid.NewV7().String()
	r.exec(
		`INSERT INTO entities (id, kind, slug, default_name) VALUES ($1, 'franchise', 'saga', 'Saga'), ($2, 'genre', 'rpg', 'RPG')`,
		franchise,
		genre,
	)

	r.subject = r.item("game", "public")
	r.link(r.subject, franchise, "franchise")
	r.link(r.subject, genre, "genre")

	var ids, scores []string

	for i := range 9 {
		vis := "public"
		if i == 4 {
			vis = "hidden"
		}

		id := r.item("game", vis)
		if i == 0 {
			r.link(id, franchise, "franchise")
		} else {
			r.link(id, genre, "genre")
		}

		if vis == "hidden" {
			r.hidden = id
		}

		ids, scores = append(ids, id), append(scores, strconv.FormatFloat(0.9-float64(i)*0.05, 'f', 2, 64))
	}

	r.exec(`INSERT INTO item_similar (item_id, similar_ids, scores, cooccur, feature_hash, computed_at)
VALUES ($1, $2::uuid[], $3::real[], $4::int[], '\x00', now())`,
		r.subject, "{"+strings.Join(ids, ",")+"}", "{"+strings.Join(scores, ",")+"}", "{0,0,0,0,0,0,0,0,0}")

	book, movie := r.item("book", "public"), r.item("movie", "public")
	r.exec(
		`INSERT INTO item_links (from_item_id, to_item_id, relation, source) VALUES ($1, $2, 'based_on', 'hexascore'), ($3, $4, 'based_on', 'hexascore')`,
		book,
		r.subject,
		r.subject,
		movie,
	)

	return r, ids
}

func (r *relatedRig) exec(q string, args ...any) {
	r.t.Helper()

	if _, err := r.acc.DB.ExecContext(r.t.Context(), q, args...); err != nil {
		r.t.Fatalf("%v\n%s", err, q)
	}
}

func (r *relatedRig) item(typ, visibility string) string {
	r.n++
	id := uuid.NewV7().String()
	r.exec(`INSERT INTO items (id, type, slug, wikidata_qid, visibility) VALUES ($1, $2, $3, $4, $5)`,
		id, typ, fmt.Sprintf("rel-%d", r.n), fmt.Sprintf("Q%d", 7000+r.n), visibility)
	r.exec(`INSERT INTO item_translations (item_id, locale, title, title_source) VALUES ($1, 'en', $2, 'hexascore')`,
		id, fmt.Sprintf("Title %d", r.n))

	return id
}

func (r *relatedRig) link(item, entity, role string) {
	r.exec(`INSERT INTO item_entities (item_id, entity_id, role, source) VALUES ($1, $2, $3, 'hexascore')`,
		item, entity, role)
}

func TestItemRelated_ShowProps(t *testing.T) {
	t.Parallel()

	r, ids := newRelatedRig(t)
	c := r.app.Client(t)

	// 9 stored, one hidden: the page shows the 8 public ones, in stored order, each with its top shared feature.
	c.Inertia("").Get("/games/rel-1").AssertOK().AssertInertia("Catalog/Show").
		HasProp("similar", 8).
		PropEquals("similar.0.id", ids[0]).
		PropEquals("similar.0.reason.kind", "franchise").
		PropEquals("similar.0.reason.entity", "Saga").
		PropEquals("similar.1.reason.kind", "genre").
		PropEquals("similar.1.reason.entity", "RPG").
		HasProp("crossMedia", 2).
		PropEquals("crossMedia.0.relation", "adapted_as").
		PropEquals("crossMedia.1.relation", "based_on")

	// An item with no item_similar row and no links gets empty lists, never null.
	c.Inertia("").Get("/games/rel-2").AssertOK().AssertInertia("Catalog/Show").
		HasProp("similar", 0).
		HasProp("crossMedia", 0)
}

func TestAppRelated(t *testing.T) {
	t.Parallel()

	r, _ := newRelatedRig(t)

	u, err := r.acc.Users.ByID(
		t.Context(),
		strconv.FormatInt(makeUser(t, r.app, "rel@example.com", "user", true).ID, 10),
	)
	if err != nil {
		t.Fatal(err)
	}

	tok, _, err := r.acc.Tokens.Create(t.Context(), u.AuthID(), "phone", []string{"catalog:read"}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	url := "/app/v1/items/" + r.subject + "/related"

	r.app.Client(t).GetJSON(url).AssertUnauthorized()
	r.app.Client(t).WithToken(tok).GetJSON("/app/v1/items/" + uuid.NewV7().String() + "/related").AssertNotFound()
	r.app.Client(t).WithToken(tok).GetJSON("/app/v1/items/not-a-uuid/related").AssertNotFound()

	r.app.Client(t).WithToken(tok).GetJSON(url).AssertOK().
		AssertJSONCount("similar", 8).
		AssertJSONCount("crossMedia", 2).
		AssertJSONPath("similar.0.reason.kind", "franchise")
}
