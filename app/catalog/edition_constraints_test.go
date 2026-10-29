//go:build integration

package catalog_test

import (
	"context"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/uuid"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

func TestISBNSQL(t *testing.T) {
	db := entityDB(t)
	for _, isbn := range []string{
		"9780306406157", "9780306406158", "9791032305690", "9770306406156", "978030640615", "",
	} {
		var got bool
		if err := db.QueryRowContext(context.Background(), `SELECT isbn13_is_valid($1)`, isbn).Scan(&got); err != nil {
			t.Fatal(err)
		}
		if want := catalog.ValidISBN13(isbn); got != want {
			t.Errorf("isbn13_is_valid(%q) = %v, Go says %v", isbn, got, want)
		}
	}
}

func TestEditionConstraints(t *testing.T) {
	db := entityDB(t)
	ctx := context.Background()
	book := newItem(t, db, "book", "dune", "Q190192", "public")
	otherBook := newItem(t, db, "book", "emma", "Q216040", "public")
	movie := newItem(t, db, "movie", "dune-film", "Q1", "public")
	edition := uuid.NewV7().String()

	ins := `INSERT INTO editions (id, work_id, isbn13, title, language, source) VALUES ($1, $2, $3, 'Dune', 'und', 'wikidata')`
	cases := []struct {
		name, query, want string
		args              []any
	}{
		{"valid isbn, language und, no publisher", ins, "", []any{edition, book, "9780306406157"}},
		{"bad check digit", ins, "check", []any{uuid.NewV7().String(), book, "9780306406158"}},
		{"duplicate isbn", ins, "unique", []any{uuid.NewV7().String(), book, "9780306406157"}},
		{"work must be a book", ins, "foreign key", []any{uuid.NewV7().String(), movie, "9791032305690"}},
		{
			"needs an identifier",
			`INSERT INTO editions (id, work_id, title, language, source) VALUES ($1, $2, 'x', 'en', 'wikidata')`,
			"check",
			[]any{uuid.NewV7().String(), book},
		},
		{
			"format check",
			`INSERT INTO editions (id, work_id, openlibrary_id, title, language, format, source) VALUES ($1, $2, 'OL1M', 'x', 'en', 'scroll', 'wikidata')`,
			"check",
			[]any{uuid.NewV7().String(), book},
		},
		{
			"primary edition of the same work",
			`UPDATE items SET primary_edition_id = $1 WHERE id = $2`,
			"",
			[]any{edition, book},
		},
		{
			"primary edition of another work",
			`UPDATE items SET primary_edition_id = $1 WHERE id = $2`,
			"foreign key",
			[]any{edition, otherBook},
		},
		{
			"redirect needs old slug or id",
			`INSERT INTO catalog_redirects (scope, namespace, target_id, reason) VALUES ('item', 'book', $1, 'merge')`,
			"check",
			[]any{book},
		},
		{
			"redirect slug ok",
			`INSERT INTO catalog_redirects (scope, namespace, old_slug, target_id, reason) VALUES ('item', 'book', 'dune-old', $1, 'rename')`,
			"",
			[]any{book},
		},
		{
			"redirect slug unique per namespace",
			`INSERT INTO catalog_redirects (scope, namespace, old_slug, target_id, reason) VALUES ('item', 'book', 'dune-old', $1, 'rename')`,
			"unique",
			[]any{book},
		},
		{
			"redirect same slug other namespace",
			`INSERT INTO catalog_redirects (scope, namespace, old_slug, target_id, reason) VALUES ('item', 'movie', 'dune-old', $1, 'rename')`,
			"",
			[]any{movie},
		},
		{
			"redirect reason check",
			`INSERT INTO catalog_redirects (scope, namespace, old_slug, target_id, reason) VALUES ('item', 'book', 'x', $1, 'typo')`,
			"check",
			[]any{book},
		},
		{
			"affiliate store check",
			`INSERT INTO affiliate_links (item_id, store, url) VALUES ($1, 'ebay', 'https://x')`,
			"check",
			[]any{book},
		},
		{
			"affiliate ok",
			`INSERT INTO affiliate_links (item_id, store, url) VALUES ($1, 'amazon', 'https://x')`,
			"",
			[]any{book},
		},
		{"deleting the primary edition keeps the item", `DELETE FROM editions WHERE id = $1`, "", []any{edition}},
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

	var primary *string
	if err := db.QueryRowContext(ctx, `SELECT primary_edition_id::text FROM items WHERE id = $1`, book).
		Scan(&primary); err != nil {
		t.Fatal(err)
	}
	if primary != nil {
		t.Fatalf("primary_edition_id = %v after edition delete, want null", *primary)
	}
}
