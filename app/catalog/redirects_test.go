//go:build integration

package catalog_test

import (
	"context"
	"errors"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

func TestRedirectRenameFlattensChains(t *testing.T) {
	db := entityDB(t)
	ctx := context.Background()
	id := newItem(t, db, "game", "half-life", "Q1", "public")

	for _, slug := range []string{"half-life-1", "half-life-classic"} {
		if err := catalog.RenameSlug(ctx, db, catalog.ScopeItem, "game", id, slug); err != nil {
			t.Fatalf("rename to %s: %v", slug, err)
		}
	}

	for _, old := range []string{"half-life", "half-life-1"} {
		got, ok, err := catalog.ResolveSlug(ctx, db, catalog.ScopeItem, "game", old)
		if err != nil || !ok || got.ID != id || got.Slug != "half-life-classic" {
			t.Fatalf("ResolveSlug(%s) = %+v, %v, %v; want one hop to the live slug", old, got, ok, err)
		}
	}

	if _, ok, _ := catalog.ResolveSlug(ctx, db, catalog.ScopeItem, "movie", "half-life"); ok {
		t.Fatal("redirects are per namespace")
	}

	// Renaming back drops the redirect: a live slug is never also an old slug (I-1).
	if err := catalog.RenameSlug(ctx, db, catalog.ScopeItem, "game", id, "half-life"); err != nil {
		t.Fatal(err)
	}

	if _, ok, _ := catalog.ResolveSlug(ctx, db, catalog.ScopeItem, "game", "half-life"); ok {
		t.Fatal("live slug still redirects")
	}
}

func TestRedirectRenameRefusesSlugOfAnotherRedirect(t *testing.T) {
	db := entityDB(t)
	ctx := context.Background()
	a := newItem(t, db, "game", "a", "Q1", "public")
	b := newItem(t, db, "game", "b", "Q2", "public")

	if err := catalog.RenameSlug(ctx, db, catalog.ScopeItem, "game", a, "a2"); err != nil {
		t.Fatal(err)
	}

	if err := catalog.RenameSlug(ctx, db, catalog.ScopeItem, "game", b, "a"); !errors.Is(err, catalog.ErrSlugReserved) {
		t.Fatalf("rename onto another row's old slug: %v, want ErrSlugReserved", err)
	}

	if err := catalog.RenameSlug(
		ctx,
		db,
		catalog.ScopeItem,
		"game",
		b,
		"trending",
	); !errors.Is(
		err,
		catalog.ErrSlugReserved,
	) {
		t.Fatalf("rename to a reserved word: %v, want ErrSlugReserved", err)
	}
}

func TestRedirectMergeFlattensChains(t *testing.T) {
	db := entityDB(t)
	ctx := context.Background()
	src := newItem(t, db, "book", "dune-old", "Q1", "public")
	dst := newItem(t, db, "book", "dune", "Q2", "public")

	if err := catalog.RenameSlug(ctx, db, catalog.ScopeItem, "book", src, "dune-1965"); err != nil {
		t.Fatal(err)
	}

	if err := catalog.RecordMerge(
		ctx,
		db,
		catalog.ScopeItem,
		"book",
		catalog.Target{ID: src, Slug: "dune-1965"},
		dst,
	); err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(ctx, `DELETE FROM items WHERE id = $1`, src); err != nil {
		t.Fatal(err)
	}

	for _, slug := range []string{"dune-old", "dune-1965"} {
		got, ok, err := catalog.ResolveSlug(ctx, db, catalog.ScopeItem, "book", slug)
		if err != nil || !ok || got.ID != dst || got.Slug != "dune" {
			t.Fatalf("ResolveSlug(%s) = %+v, %v, %v; want the merge target", slug, got, ok, err)
		}
	}

	got, ok, err := catalog.ResolveID(ctx, db, catalog.ScopeItem, "book", src)
	if err != nil || !ok || got.ID != dst {
		t.Fatalf("ResolveID(source) = %+v, %v, %v; want the merge target", got, ok, err)
	}
}

func TestRedirectUniqueSlug(t *testing.T) {
	db := entityDB(t)
	ctx := context.Background()
	newItem(t, db, "game", "doom", "Q1", "public")
	newItem(t, db, "game", "doom-1993", "Q2", "public")
	old := newItem(t, db, "game", "quake", "Q3", "public")

	if err := catalog.RenameSlug(ctx, db, catalog.ScopeItem, "game", old, "quake-1"); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		title string
		year  int
		want  string
	}{
		{"Fresh Game", 2020, "fresh-game"},
		{"Doom", 1993, "doom-2"},
		{"Doom", 2016, "doom-2016"},
		{"Quake", 1996, "quake-1996"},
		{"Trending", 0, "trending-2"},
	} {
		got, err := catalog.UniqueSlug(ctx, db, catalog.ScopeItem, "game", tc.title, tc.year)
		if err != nil || got != tc.want {
			t.Errorf("UniqueSlug(%q, %d) = %q, %v; want %q", tc.title, tc.year, got, err, tc.want)
		}
	}

	if _, err := catalog.UniqueSlug(
		ctx,
		db,
		catalog.ScopeItem,
		"game",
		"日本語",
		0,
	); !errors.Is(
		err,
		catalog.ErrEmptySlug,
	) {
		t.Fatalf("empty title: %v, want ErrEmptySlug", err)
	}
}
