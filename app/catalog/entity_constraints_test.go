//go:build integration

package catalog_test

import (
	"context"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/orm"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func entityDB(t *testing.T) *database.DB {
	t.Helper()
	url := pgtest.Database(t, "")
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{"DATABASE_URL": url}})
	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(context.Background(), `INSERT INTO data_sources (code, name, licence, url)
VALUES ('wikidata', 'Wikidata', 'CC0-1.0', 'https://www.wikidata.org')`); err != nil {
		t.Fatal(err)
	}
	return db
}

func newEntity(t *testing.T, db *database.DB, kind, slug string) string {
	t.Helper()
	e := models.Entity{
		ID:          uuid.NewV7().String(),
		Kind:        kind,
		Slug:        slug,
		DefaultName: slug,
		ExternalIDs: orm.JSON[map[string]string]{V: map[string]string{}},
		Provenance:  orm.JSON[map[string]string]{V: map[string]string{}},
	}
	if err := orm.Create(context.Background(), db, &e); err != nil {
		t.Fatalf("entity %s: %v", slug, err)
	}
	return e.ID
}

func newItem(t *testing.T, db *database.DB, typ, slug, qid, visibility string) string {
	t.Helper()
	id := uuid.NewV7().String()
	if _, err := db.ExecContext(context.Background(),
		`INSERT INTO items (id, type, slug, wikidata_qid, visibility) VALUES ($1, $2, $3, $4, $5)`,
		id, typ, slug, qid, visibility); err != nil {
		t.Fatalf("item %s: %v", slug, err)
	}
	return id
}

func TestEntityConstraints(t *testing.T) {
	db := entityDB(t)
	ctx := context.Background()
	person := newEntity(t, db, "person", "ridley-scott")
	item := newItem(t, db, "movie", "alien", "Q103569", "public")

	cases := []struct {
		name, query, want string
		args              []any
	}{
		{
			"kind check",
			`INSERT INTO entities (id, kind, slug, default_name) VALUES ($1, 'engine', 'unreal', 'Unreal')`,
			"check",
			[]any{uuid.NewV7().String()},
		},
		{
			"slug format",
			`INSERT INTO entities (id, kind, slug, default_name) VALUES ($1, 'person', 'Bad Slug', 'x')`,
			"check",
			[]any{uuid.NewV7().String()},
		},
		{
			"unique kind+slug",
			`INSERT INTO entities (id, kind, slug, default_name) VALUES ($1, 'person', 'ridley-scott', 'x')`,
			"unique",
			[]any{uuid.NewV7().String()},
		},
		{
			"same slug other kind",
			`INSERT INTO entities (id, kind, slug, default_name) VALUES ($1, 'organization', 'ridley-scott', 'x')`,
			"",
			[]any{uuid.NewV7().String()},
		},
		{
			"native entity without qid",
			`INSERT INTO entities (id, kind, slug, default_name) VALUES ($1, 'genre', 'noir', 'Noir')`,
			"",
			[]any{uuid.NewV7().String()},
		},
		{
			"qid format",
			`INSERT INTO entities (id, kind, slug, default_name, wikidata_qid) VALUES ($1, 'genre', 'sf', 'SF', 'X1')`,
			"check",
			[]any{uuid.NewV7().String()},
		},
		{
			"translation locale",
			`INSERT INTO entity_translations (entity_id, locale, name, name_source) VALUES ($1, 'fr', 'x', 'wikidata')`,
			"check",
			[]any{person},
		},
		{
			"translation ok",
			`INSERT INTO entity_translations (entity_id, locale, name, name_source) VALUES ($1, 'pt-BR', 'Ridley Scott', 'wikidata')`,
			"",
			[]any{person},
		},
		{
			"role check",
			`INSERT INTO item_entities (item_id, entity_id, role, source) VALUES ($1, $2, 'translator', 'wikidata')`,
			"check",
			[]any{item, person},
		},
		{
			"role ok",
			`INSERT INTO item_entities (item_id, entity_id, role, source) VALUES ($1, $2, 'director', 'wikidata')`,
			"",
			[]any{item, person},
		},
		{
			"duplicate role",
			`INSERT INTO item_entities (item_id, entity_id, role, source) VALUES ($1, $2, 'director', 'wikidata')`,
			"duplicate",
			[]any{item, person},
		},
		{
			"entity link relation",
			`INSERT INTO entity_links (from_id, to_id, relation) VALUES ($1, $1, 'owns')`,
			"check",
			[]any{person},
		},
		{
			"item link self",
			`INSERT INTO item_links (from_item_id, to_item_id, relation, source) VALUES ($1, $1, 'based_on', 'wikidata')`,
			"check",
			[]any{item},
		},
		{"restrict referenced entity", `DELETE FROM entities WHERE id = $1`, "foreign key", []any{person}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := db.ExecContext(ctx, c.query, c.args...)
			switch {
			case c.want == "" && err != nil:
				t.Fatalf("want success, got %v", err)
			case c.want != "" && (err == nil || !strings.Contains(strings.ToLower(err.Error()), c.want)):
				t.Fatalf("want %s violation, got %v", c.want, err)
			}
		})
	}
}

func TestEntityStats(t *testing.T) {
	db := entityDB(t)
	ctx := context.Background()
	person := newEntity(t, db, "person", "multi-role")
	genre := newEntity(t, db, "genre", "drama")
	movie1 := newItem(t, db, "movie", "m1", "Q1", "public")
	movie2 := newItem(t, db, "movie", "m2", "Q2", "public")
	hidden := newItem(t, db, "movie", "m3", "Q3", "hidden")
	series := newItem(t, db, "series", "s1", "Q4", "public")

	link := func(item, entity, role string) {
		t.Helper()
		if _, err := db.ExecContext(ctx,
			`INSERT INTO item_entities (item_id, entity_id, role, source) VALUES ($1, $2, $3, 'wikidata')`,
			item, entity, role); err != nil {
			t.Fatal(err)
		}
	}
	link(movie1, person, "director")
	link(movie1, person, "writer") // two roles on one item count once
	link(movie2, person, "director")
	link(hidden, person, "director") // not public: excluded
	link(series, person, "creator")
	link(movie1, genre, "genre")

	counts := func() map[string]int {
		t.Helper()
		rows, err := db.QueryContext(ctx, `SELECT entity_id || ':' || type, item_count FROM entity_stats`)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		got := map[string]int{}
		for rows.Next() {
			var k string
			var n int
			if err := rows.Scan(&k, &n); err != nil {
				t.Fatal(err)
			}
			got[k] = n
		}
		return got
	}
	want := map[string]int{person + ":movie": 2, person + ":series": 1, genre + ":movie": 1}
	check := func(label string) {
		t.Helper()
		got := counts()
		if len(got) != len(want) {
			t.Fatalf("%s: got %v want %v", label, got, want)
		}
		for k, n := range want {
			if got[k] != n {
				t.Fatalf("%s: got %v want %v", label, got, want)
			}
		}
	}

	if err := catalog.RefreshEntityStats(ctx, db); err != nil {
		t.Fatal(err)
	}
	check("first run")
	if err := catalog.RefreshEntityStats(ctx, db); err != nil {
		t.Fatal(err)
	}
	check("idempotent rerun")

	if _, err := db.ExecContext(ctx, `DELETE FROM item_entities WHERE item_id = $1`, series); err != nil {
		t.Fatal(err)
	}
	delete(want, person+":series")
	if err := catalog.RefreshEntityStats(ctx, db); err != nil {
		t.Fatal(err)
	}
	check("stale row removed")
}
