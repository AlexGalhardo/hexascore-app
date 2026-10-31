//go:build integration

package community

import (
	"context"
	"errors"
	"testing"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/queue"
)

const createBook = "00000000-0000-0000-0000-0000000000cb"

func insertCreate(t *testing.T, db *database.DB, id, targetType, domain, changes string) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), `INSERT INTO suggestions
(id, kind, target_type, target_domain, locale, changes, status, version, submitted_at)
VALUES ($1, 'create', $2, $3, 'en', $4::jsonb, 'in_review', 1, now())`, id, targetType, domain, changes); err != nil {
		t.Fatal(err)
	}
}

func createAuditRows(t *testing.T, db *database.DB, id, targetType, newID string) int {
	t.Helper()
	var rows int
	if err := db.QueryRowContext(context.Background(), `SELECT count(*) FROM catalog_audit WHERE suggestion_id = $1
  AND source = 'create' AND target_type = $2 AND target_id = $3 AND before = 'null'::jsonb`,
		id, targetType, newID).Scan(&rows); err != nil {
		t.Fatal(err)
	}

	return rows
}

func TestApplyCreateEntity(t *testing.T) {
	db := openCommunityDB(t)
	const id = "00000000-0000-7000-8000-00000000c001"
	insertCreate(t, db, id, "entity", "movie", `{"title":{"before":null,"after":"Jane Director"},
"entity_kind":{"before":null,"after":"person"},"external_ids.wikidata":{"before":null,"after":"Q777"}}`)

	a, _ := newApplier(t, db)
	if err := a.Approve(context.Background(), id, 1, ApproveOptions{}, adminActor); err != nil {
		t.Fatal(err)
	}

	var newID, kind, slug, qid, name, prov string
	if err := db.QueryRowContext(context.Background(), `SELECT e.id::text, e.kind, e.slug, e.wikidata_qid,
  e.default_name, e.provenance->>'default_name'
FROM entities e JOIN suggestions s ON s.meta->>'created_id' = e.id::text WHERE s.id = $1`, id).
		Scan(&newID, &kind, &slug, &qid, &name, &prov); err != nil {
		t.Fatal(err)
	}
	if kind != "person" || slug != "jane-director" || qid != "Q777" || name != "Jane Director" || prov != "community" {
		t.Fatalf("created entity %s %s %s %s %s", kind, slug, qid, name, prov)
	}
	// title and external_ids.wikidata; entity_kind is the row's identity, not an editable field.
	if n := createAuditRows(t, db, id, "entity", newID); n != 2 {
		t.Fatalf("create audit rows = %d, want 2", n)
	}
	assertStatus(t, db, id, "applied", 3)

	// Same Q-id again: 409 suggestion-duplicate naming the existing entity, nothing written.
	const dup = "00000000-0000-7000-8000-00000000c002"
	insertCreate(t, db, dup, "entity", "movie", `{"title":{"before":null,"after":"Jane Again"},
"entity_kind":{"before":null,"after":"person"},"external_ids.wikidata":{"before":null,"after":"Q777"}}`)
	var p *Problem
	if err := a.Approve(context.Background(), dup, 1, ApproveOptions{}, adminActor); !errors.As(err, &p) ||
		p.Code != "suggestion-duplicate" || p.ExistingID != newID {
		t.Fatalf("duplicate err = %#v", err)
	}
	assertStatus(t, db, dup, "in_review", 1)
}

func TestApplyCreateEdition(t *testing.T) {
	db := openCommunityDB(t)
	if _, err := db.ExecContext(context.Background(), `INSERT INTO items (id, type, slug, wikidata_qid, original_title)
VALUES ('`+createBook+`', 'book', 'dom-casmurro', 'Q1080', 'Dom Casmurro')`); err != nil {
		t.Fatal(err)
	}
	const id = "00000000-0000-7000-8000-00000000c003"
	insertCreate(t, db, id, "edition", "book", `{"title":{"before":null,"after":"Dom Casmurro (Penguin)"},
"work_id":{"before":null,"after":"`+createBook+`"},"language":{"before":null,"after":"pt-BR"},
"external_ids.isbn13":{"before":null,"after":"9780306406157"}}`)

	a, fake := newApplier(t, db)
	if err := a.Approve(context.Background(), id, 1, ApproveOptions{}, adminActor); err != nil {
		t.Fatal(err)
	}

	var newID, work, title, lang, isbn, source string
	if err := db.QueryRowContext(context.Background(), `SELECT e.id::text, e.work_id::text, e.title, e.language,
  e.isbn13, e.source FROM editions e JOIN suggestions s ON s.meta->>'created_id' = e.id::text WHERE s.id = $1`, id).
		Scan(&newID, &work, &title, &lang, &isbn, &source); err != nil {
		t.Fatal(err)
	}
	if work != createBook || title != "Dom Casmurro (Penguin)" || lang != "pt-BR" || isbn != "9780306406157" ||
		source != "community" {
		t.Fatalf("created edition %s %s %s %s %s", work, title, lang, isbn, source)
	}
	if n := createAuditRows(t, db, id, "edition", newID); n != 2 {
		t.Fatalf("create audit rows = %d, want 2 (title, isbn13)", n)
	}
	assertStatus(t, db, id, "applied", 3)
	if n := len(queue.Dispatched[CatalogChanged](fake)); n != 1 {
		t.Fatalf("catalog.changed jobs = %d", n)
	}

	// Same ISBN: 409 suggestion-duplicate on the existing edition.
	const dup = "00000000-0000-7000-8000-00000000c004"
	insertCreate(t, db, dup, "edition", "book", `{"title":{"before":null,"after":"Again"},
"work_id":{"before":null,"after":"`+createBook+`"},"language":{"before":null,"after":"en"},
"external_ids.isbn13":{"before":null,"after":"9780306406157"}}`)
	var p *Problem
	if err := a.Approve(context.Background(), dup, 1, ApproveOptions{}, adminActor); !errors.As(err, &p) ||
		p.Code != "suggestion-duplicate" || p.ExistingID != newID {
		t.Fatalf("duplicate err = %#v", err)
	}

	// A parent work that is gone is a 422 on work_id, not a 500.
	const orphan = "00000000-0000-7000-8000-00000000c005"
	insertCreate(t, db, orphan, "edition", "book", `{"title":{"before":null,"after":"Orphan"},
"work_id":{"before":null,"after":"00000000-0000-0000-0000-0000000000ff"},"language":{"before":null,"after":"en"},
"external_ids.openlibrary":{"before":null,"after":"OL1M"}}`)
	var ve *ValidationError
	if err := a.Approve(context.Background(), orphan, 1, ApproveOptions{}, adminActor); !errors.As(err, &ve) ||
		ve.Field != "work_id" {
		t.Fatalf("orphan err = %#v", err)
	}
	assertStatus(t, db, orphan, "in_review", 1)
}
