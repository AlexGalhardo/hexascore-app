//go:build integration

package catalog_test

import (
	"context"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/orm"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func TestItemConstraints(t *testing.T) {
	url := pgtest.Database(t, "")
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{"DATABASE_URL": url}})
	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	if _, err := db.ExecContext(ctx, `INSERT INTO data_sources (code, name, licence, url)
VALUES ('wikidata', 'Wikidata', 'CC0-1.0', 'https://www.wikidata.org')`); err != nil {
		t.Fatal(err)
	}

	item := models.Item{
		ID:               uuid.NewV7().String(),
		Type:             "game",
		Slug:             "half-life",
		WikidataQID:      "Q279744",
		ReleasePrecision: "day",
		ReleaseStatus:    "released",
		Visibility:       "public",
		Trailers: orm.JSON[[]models.Trailer]{
			V: []models.Trailer{},
		},
		ExternalIDs: orm.JSON[map[string]string]{V: map[string]string{"igdb": "231"}},
		Provenance:  orm.JSON[map[string]string]{V: map[string]string{}},
	}
	if err := orm.Create(ctx, db, &item); err != nil {
		t.Fatalf("valid item: %v", err)
	}
	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO item_translations (item_id, locale, title, title_source) VALUES ($1, 'pt-BR', 'Half-Life', 'wikidata')`,
		item.ID,
	); err != nil {
		t.Fatalf("translation: %v", err)
	}

	insert := `INSERT INTO items (id, type, slug, wikidata_qid, visibility) VALUES ($1, $2, $3, $4, $5)`
	cases := []struct {
		name, typ, slug, qid, visibility, want string
	}{
		{"type check", "podcast", "x", "Q1", "public", "check"},
		{"unique type+slug", "game", "half-life", "Q2", "public", "unique"},
		{"same slug other type", "movie", "half-life", "Q3", "public", ""},
		{"unique wikidata_qid", "book", "hl-book", "Q279744", "public", "unique"},
		{"visibility check", "game", "hl2", "Q4", "deleted", "check"},
		{"takedown allowed", "game", "hl3", "Q5", "takedown", ""},
		{"slug format", "game", "Half Life", "Q6", "public", "check"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := db.ExecContext(ctx, insert, uuid.NewV7().String(), c.typ, c.slug, c.qid, c.visibility)
			switch {
			case c.want == "" && err != nil:
				t.Fatalf("want success, got %v", err)
			case c.want != "" && (err == nil || !strings.Contains(strings.ToLower(err.Error()), c.want)):
				t.Fatalf("want %s violation, got %v", c.want, err)
			}
		})
	}
}
