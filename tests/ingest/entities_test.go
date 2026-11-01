//go:build integration

package ingest_test

import (
	"testing"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/entities"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/normalize"
)

const entItem = "0192f1c6-0000-7000-8000-0000000000a1"

func seedEntityItem(t *testing.T) (store entities.PG, countQ func(string, ...any) int) {
	t.Helper()

	db := newDB(t)

	if err := normalize.SeedDataSources(t.Context(), db); err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(t.Context(),
		`INSERT INTO items (id, type, slug, wikidata_qid) VALUES ($1, 'movie', 'film', 'Q1')`, entItem); err != nil {
		t.Fatal(err)
	}

	return entities.PG{Ex: db}, func(q string, a ...any) int { return count(t, db, q, a...) }
}

func resolveOK(t *testing.T, s entities.PG, c entities.Credit) entities.Outcome {
	t.Helper()

	c.ItemID, c.Source = entItem, "wikidata"

	out, err := entities.Resolve(t.Context(), s, c)
	if err != nil {
		t.Fatal(err)
	}

	return out
}

func TestEntitiesCreateThenLinkByQidAndName(t *testing.T) {
	s, n := seedEntityItem(t)

	dir := entities.Credit{Role: "director", Kind: "person", QID: "Q55", Name: "Lana Wachowski"}
	if got := resolveOK(t, s, dir); got != entities.Created {
		t.Fatalf("first = %v, want Created", got)
	}

	if got := resolveOK(t, s, dir); got != entities.Linked {
		t.Fatalf("rerun = %v, want Linked", got)
	}

	// the same person credited without a QID by name links to the entity created above
	writer := entities.Credit{Role: "writer", Kind: "person", Name: "lana  WACHOWSKI"}
	if got := resolveOK(t, s, writer); got != entities.Linked {
		t.Fatalf("by name = %v, want Linked", got)
	}

	if c := n(`SELECT count(*) FROM entities`); c != 1 {
		t.Errorf("entities = %d, want 1", c)
	}

	if c := n(`SELECT count(*) FROM item_entities WHERE source = 'wikidata'`); c != 2 {
		t.Errorf("item_entities = %d, want 2", c)
	}
}

func TestEntitiesSlugClashKeepsTheCredit(t *testing.T) {
	s, n := seedEntityItem(t)

	if _, err := s.Ex.ExecContext(t.Context(), `INSERT INTO entities (id, kind, slug, default_name)
VALUES (gen_random_uuid(), 'genre', 'drama', 'Something else')`); err != nil {
		t.Fatal(err)
	}

	if got := resolveOK(
		t,
		s,
		entities.Credit{Role: "genre", Kind: "genre", QID: "Q2", Name: "Drama"},
	); got != entities.Created {
		t.Fatalf("outcome = %v, want Created", got)
	}

	if c := n(`SELECT count(*) FROM entities WHERE slug = 'drama-q2'`); c != 1 {
		t.Errorf("entity with the QID slug = %d, want 1", c)
	}

	if got := resolveOK(
		t,
		s,
		entities.Credit{Role: "country", Kind: "country", QID: "Q3", Name: "Drama"},
	); got != entities.Created {
		t.Fatalf("other kind = %v, want Created", got)
	}
}

func TestEntitiesAmbiguousReviewIsNotDuplicated(t *testing.T) {
	s, n := seedEntityItem(t)

	for _, q := range []string{"Q10", "Q11"} {
		if _, err := s.Ex.ExecContext(t.Context(), `INSERT INTO entities (id, kind, slug, wikidata_qid, default_name)
VALUES (gen_random_uuid(), 'person', 'john-smith-'||lower($1::text), $1::text, 'John Smith')`, q); err != nil {
			t.Fatal(err)
		}
	}

	c := entities.Credit{Role: "cast", Kind: "person", Name: "John Smith"}
	for range 2 {
		if got := resolveOK(t, s, c); got != entities.Reviewed {
			t.Fatalf("outcome = %v, want Reviewed", got)
		}
	}

	if c := n(
		`SELECT count(*) FROM entity_match_reviews WHERE status = 'open' AND cardinality(candidates) = 2`,
	); c != 1 {
		t.Errorf("open reviews = %d, want 1", c)
	}

	if c := n(`SELECT count(*) FROM item_entities`); c != 0 {
		t.Errorf("item_entities = %d, want 0 while the review is open", c)
	}
}

func TestEntitiesQidUnderAnotherKindOpensReview(t *testing.T) {
	s, n := seedEntityItem(t)

	resolveOK(t, s, entities.Credit{Role: "production_company", Kind: "organization", QID: "Q77", Name: "Studio"})

	if got := resolveOK(
		t,
		s,
		entities.Credit{Role: "director", Kind: "person", QID: "Q77", Name: "Studio"},
	); got != entities.Reviewed {
		t.Fatalf("outcome = %v, want Reviewed", got)
	}

	if c := n(`SELECT count(*) FROM entity_match_reviews`); c != 1 {
		t.Errorf("reviews = %d, want 1", c)
	}
}

func TestEntitiesExternalIdLinksAndIsLearned(t *testing.T) {
	s, n := seedEntityItem(t)

	first := entities.Credit{Role: "director", Kind: "person", QID: "Q5", Name: "Someone"}
	resolveOK(t, s, first)

	// a TMDB credit with the same name teaches the entity its tmdb_person id
	byName := entities.Credit{
		Role:          "writer",
		Kind:          "person",
		Name:          "Someone",
		ExternalIDKey: "tmdb_person",
		ExternalID:    "9",
	}
	if got := resolveOK(t, s, byName); got != entities.Linked {
		t.Fatalf("by name = %v", got)
	}

	other := entities.Credit{
		Role:          "cast",
		Kind:          "person",
		Name:          "Totally Different Label",
		ExternalIDKey: "tmdb_person",
		ExternalID:    "9",
	}
	if got := resolveOK(t, s, other); got != entities.Linked {
		t.Fatalf("by external id = %v, want Linked", got)
	}

	if c := n(`SELECT count(*) FROM entities WHERE external_ids ->> 'tmdb_person' = '9'`); c != 1 {
		t.Errorf("entities with the external id = %d, want 1", c)
	}
}

func TestNormalizeWritesItemEntities(t *testing.T) {
	db := newDB(t)
	loadWikidata(t, db)
	drain(t, db, (&normSpy{}).job(0))

	ie := count(t, db, `SELECT count(*) FROM item_entities`)
	ents := count(t, db, `SELECT count(*) FROM entities`)

	if ie == 0 || ents == 0 {
		t.Fatalf("item_entities = %d, entities = %d, want credits resolved", ie, ents)
	}

	repend(t, db, "1 hour")
	drain(t, db, (&normSpy{}).job(0))

	if got := count(t, db, `SELECT count(*) FROM item_entities`); got != ie {
		t.Errorf("item_entities after a rerun = %d, want %d", got, ie)
	}

	if got := count(t, db, `SELECT count(*) FROM entities`); got != ents {
		t.Errorf("entities after a rerun = %d, want %d", got, ents)
	}
}
