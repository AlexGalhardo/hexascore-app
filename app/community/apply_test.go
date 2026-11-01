//go:build integration

package community

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

const applyItem = "00000000-0000-0000-0000-0000000000aa"

// pgErr stands in for a driver error carrying a SQLSTATE.
type pgErr string

func (e pgErr) Error() string    { return "pg error " + string(e) }
func (e pgErr) SQLState() string { return string(e) }

func newApplier(t *testing.T, db *database.DB) (*Applier, *queue.Fake) {
	t.Helper()
	fake := queue.NewFake()
	q := queue.New(fake, slog.Default())
	noop := func(context.Context, CatalogChanged) error { return nil }
	notify := func(context.Context, Notify) error { return nil }
	merged := func(context.Context, catalog.Merged) error { return nil }
	if err := q.Register(queue.Handle(noop), queue.Handle(notify), queue.Handle(merged)); err != nil {
		t.Fatal(err)
	}

	return &Applier{
		DB:    db,
		Queue: q,
		Now:   func() time.Time { return time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC) },
	}, fake
}

func seedApplyItem(t *testing.T, db *database.DB) {
	t.Helper()
	for _, q := range []string{
		`INSERT INTO data_sources (code, name, licence, url) VALUES ('wikidata','Wikidata','CC0-1.0','https://www.wikidata.org')
ON CONFLICT DO NOTHING`,
		`INSERT INTO items (id, type, slug, wikidata_qid, original_title, release_date, release_precision, provenance)
VALUES ('` + applyItem + `', 'movie', 'old', 'Q1', 'Old', '2001-12-19', 'day',
 '{"original_title":"wikidata","release_date":"wikidata"}')`,
		`INSERT INTO item_translations (item_id, locale, title, title_source) VALUES ('` + applyItem + `','en','Old','wikidata')`,
		`INSERT INTO entities (id, kind, slug, default_name) VALUES
 ('00000000-0000-0000-0000-0000000000e1','genre','drama','Drama'),
 ('00000000-0000-0000-0000-0000000000e2','genre','fantasy','Fantasy')`,
		`INSERT INTO item_entities (item_id, entity_id, role, source) VALUES
 ('` + applyItem + `','00000000-0000-0000-0000-0000000000e1','genre','wikidata')`,
	} {
		if _, err := db.ExecContext(context.Background(), q); err != nil {
			t.Fatal(err)
		}
	}
}

func seedApplySuggestion(t *testing.T, db *database.DB, id string, author int64, status, changes string) {
	t.Helper()
	var a any
	if author != 0 {
		a = author
	}
	_, err := db.ExecContext(context.Background(), `INSERT INTO suggestions
(id, author_id, kind, target_type, target_id, target_domain, locale, changes, status, version, submitted_at)
VALUES ($1, $2, 'edit', 'item', $3, 'movie', 'en', $4::jsonb, $5, 1, now())`, id, a, applyItem, changes, status)
	if err != nil {
		t.Fatal(err)
	}
}

var adminActor = Actor{ID: sql.NullInt64{}, Role: "admin"}

func originalTitle(t *testing.T, db *database.DB) (string, string) {
	t.Helper()
	var title, prov string
	if err := db.QueryRowContext(context.Background(),
		`SELECT original_title, provenance->>'original_title' FROM items WHERE id = $1`, applyItem).
		Scan(&title, &prov); err != nil {
		t.Fatal(err)
	}

	return title, prov
}

func auditCount(t *testing.T, db *database.DB) int {
	t.Helper()
	var n int
	if err := db.QueryRowContext(context.Background(), `SELECT count(*) FROM catalog_audit`).Scan(&n); err != nil {
		t.Fatal(err)
	}

	return n
}

func TestApplyAtomic(t *testing.T) {
	db := openCommunityDB(t)
	seedApplyItem(t, db)
	const id = "00000000-0000-7000-8000-00000000a001"
	seedApplySuggestion(t, db, id, 0, "in_review", `{"title":{"before":"Old","after":"New"}}`)

	a, fake := newApplier(t, db)
	a.failAt = func(step string, _ int) error {
		if step == "end" {
			return errors.New("injected")
		}
		return nil
	}
	if err := a.Approve(context.Background(), id, 1, ApproveOptions{}, adminActor); err == nil {
		t.Fatal("injected failure did not surface")
	}

	if title, prov := originalTitle(t, db); title != "Old" || prov != "wikidata" {
		t.Fatalf("item changed: %s/%s", title, prov)
	}
	assertStatus(t, db, id, "in_review", 1)
	if n := auditCount(t, db); n != 0 {
		t.Fatalf("audit rows = %d", n)
	}
	if n := eventCount(t, db, id); n != 0 {
		t.Fatalf("events = %d", n)
	}
	if n := len(fake.Messages()); n != 0 {
		t.Fatalf("jobs = %d", n)
	}
}

func TestApplyProvenance(t *testing.T) {
	db := openCommunityDB(t)
	seedApplyItem(t, db)
	author := insertAuthor(t, db, "apply-author@example.com", time.Now().AddDate(0, 0, -30), true, false)
	const id = "00000000-0000-7000-8000-00000000a002"
	seedApplySuggestion(t, db, id, author, "submitted", `{
"title":{"before":"Old","after":"New"},
"synopsis.pt-BR":{"before":null,"after":"Uma sinopse escrita com palavras próprias, longa o bastante para passar."},
"release_date":{"before":{"value":"2001-12-19","precision":"day"},"after":{"value":"2001-12-18","precision":"day"}},
"genres":{"add":[{"ID":"00000000-0000-0000-0000-0000000000e2"}],"remove":[{"ID":"00000000-0000-0000-0000-0000000000e1"}]}}`)

	a, fake := newApplier(t, db)
	if err := a.Approve(context.Background(), id, 1, ApproveOptions{}, adminActor); err != nil {
		t.Fatal(err)
	}

	if title, prov := originalTitle(t, db); title != "New" || prov != "community" {
		t.Fatalf("title %s/%s", title, prov)
	}
	var synSrc, relSrc string
	if err := db.QueryRowContext(context.Background(), `SELECT synopsis_source FROM item_translations
WHERE item_id = $1 AND locale = 'pt-BR'`, applyItem).Scan(&synSrc); err != nil || synSrc != "community" {
		t.Fatalf("synopsis source %q %v", synSrc, err)
	}
	if err := db.QueryRowContext(context.Background(), `SELECT string_agg(entity_id::text || ':' || source, ',')
FROM item_entities WHERE item_id = $1`, applyItem).Scan(&relSrc); err != nil ||
		relSrc != "00000000-0000-0000-0000-0000000000e2:community" {
		t.Fatalf("relations %q %v", relSrc, err)
	}

	// One audit row per field, holding the real previous value and provenance.
	var before, prev, src string
	if err := db.QueryRowContext(context.Background(), `SELECT before::text, prev_provenance::text, source::text
FROM catalog_audit WHERE field = 'title' AND suggestion_id = $1`, id).Scan(&before, &prev, &src); err != nil {
		t.Fatal(err)
	}
	if before != `"Old"` || prev != `"wikidata"` || src != "community" {
		t.Fatalf("title audit before=%s prev=%s src=%s", before, prev, src)
	}
	if n := auditCount(t, db); n != 4 {
		t.Fatalf("audit rows = %d, want 4", n)
	}

	assertStatus(t, db, id, "applied", 3)
	var approved int
	if err := db.QueryRowContext(context.Background(), `SELECT approved_total FROM contributor_stats WHERE user_id = $1`,
		author).
		Scan(&approved); err != nil ||
		approved != 1 {
		t.Fatalf("approved_total %d %v", approved, err)
	}

	changed := queue.Dispatched[CatalogChanged](fake)
	if len(changed) != 1 || len(changed[0].AuditIDs) != 4 || changed[0].TargetID != applyItem {
		t.Fatalf("catalog.changed %+v", changed)
	}
	notes := queue.Dispatched[Notify](fake)
	if len(notes) != 1 || notes[0].State != "applied" {
		t.Fatalf("community.notify %+v", notes)
	}
}

func TestApplyWritesChanges(t *testing.T) {
	db := openCommunityDB(t)
	seedApplyItem(t, db)
	const id, rival = "00000000-0000-7000-8000-00000000a003", "00000000-0000-7000-8000-00000000a004"
	seedApplySuggestion(t, db, id, 0, "submitted",
		`{"title":{"before":"Old","after":"New"}}`)
	seedApplySuggestion(t, db, rival, 0, "submitted", `{"title":{"before":"Old","after":"Other"}}`)

	a, _ := newApplier(t, db)
	if err := a.Approve(context.Background(), id, 1, ApproveOptions{}, adminActor); err != nil {
		t.Fatal(err)
	}

	var action, fields string
	if err := db.QueryRowContext(context.Background(), `SELECT action, array_to_string(fields, ',') FROM api_changes
WHERE item_id = $1`, applyItem).Scan(&action, &fields); err != nil {
		t.Fatal(err)
	}
	if action != "updated" || fields != "title" {
		t.Fatalf("api_changes %s %s", action, fields)
	}

	var stale bool
	if err := db.QueryRowContext(context.Background(), `SELECT COALESCE((meta->>'stale')::boolean, false)
FROM suggestions WHERE id = $1`, rival).Scan(&stale); err != nil || !stale {
		t.Fatalf("rival not flagged stale: %v %v", stale, err)
	}

	// The rival's before no longer matches: 409 without confirm_stale, applied with it.
	err := a.Approve(context.Background(), rival, 1, ApproveOptions{}, adminActor)
	var te *TransitionError
	if !errors.As(err, &te) || te.Code != "stale-before" || te.Field != "title" {
		t.Fatalf("rival err = %#v", err)
	}
	if err := a.Approve(context.Background(), rival, 1, ApproveOptions{ConfirmStale: true}, adminActor); err != nil {
		t.Fatal(err)
	}
	var before string
	if err := db.QueryRowContext(context.Background(), `SELECT before::text FROM catalog_audit
WHERE suggestion_id = $1`, rival).Scan(&before); err != nil || before != `"New"` {
		t.Fatalf("confirm_stale audit before = %s %v", before, err)
	}
}

func TestApplyLockedAndDeferred(t *testing.T) {
	db := openCommunityDB(t)
	seedApplyItem(t, db)
	const locked, cover = "00000000-0000-7000-8000-00000000a005", "00000000-0000-7000-8000-00000000a006"
	seedApplySuggestion(t, db, locked, 0, "submitted", `{"title":{"before":"Old","after":"New"}}`)
	seedApplySuggestion(t, db, cover, 0, "submitted", `{"cover":{"before":null,"after":"community/covers/x.webp"}}`)
	if _, err := db.ExecContext(
		context.Background(),
		`INSERT INTO catalog_locks (id, target_type, target_id, field, reason)
VALUES ('00000000-0000-7000-8000-00000000c001', 'item', $1, 'title', 'edit war')`,
		applyItem,
	); err != nil {
		t.Fatal(err)
	}

	a, fake := newApplier(t, db)
	err := a.Approve(context.Background(), locked, 1, ApproveOptions{}, adminActor)
	var p *Problem
	if !errors.As(err, &p) || p.Status != 423 {
		t.Fatalf("locked err = %#v", err)
	}

	if err := a.Approve(context.Background(), cover, 1, ApproveOptions{}, adminActor); err != nil {
		t.Fatal(err)
	}
	assertStatus(t, db, cover, "approved", 2)
	if n := auditCount(t, db); n != 0 {
		t.Fatalf("deferred apply wrote %d audit rows", n)
	}
	if n := len(queue.Dispatched[CatalogChanged](fake)); n != 0 {
		t.Fatalf("deferred apply dispatched catalog.changed")
	}
}

func TestApplyDeadlockRetry(t *testing.T) {
	db := openCommunityDB(t)
	seedApplyItem(t, db)
	const once, twice = "00000000-0000-7000-8000-00000000a007", "00000000-0000-7000-8000-00000000a008"
	seedApplySuggestion(t, db, once, 0, "submitted", `{"title":{"before":"Old","after":"New"}}`)
	seedApplySuggestion(
		t,
		db,
		twice,
		0,
		"submitted",
		`{"release_date":{"before":{"value":"2001-12-19","precision":"day"},
"after":{"value":"2001-12-20","precision":"day"}}}`,
	)

	a, fake := newApplier(t, db)
	attempts := 0
	a.failAt = func(step string, attempt int) error {
		if step == "end" {
			attempts++
			if attempt == 0 {
				return pgErr("40P01")
			}
		}
		return nil
	}
	if err := a.Approve(context.Background(), once, 1, ApproveOptions{}, adminActor); err != nil {
		t.Fatal(err)
	}
	if attempts != 2 || auditCount(t, db) != 1 || len(queue.Dispatched[CatalogChanged](fake)) != 1 {
		t.Fatalf("attempts=%d audit=%d", attempts, auditCount(t, db))
	}

	a.failAt = func(step string, _ int) error {
		if step == "end" {
			return pgErr("40P01")
		}
		return nil
	}
	if err := a.Approve(
		context.Background(),
		twice,
		1,
		ApproveOptions{},
		adminActor,
	); !errors.Is(
		err,
		ErrApplyDeadlock,
	) {
		t.Fatalf("second deadlock err = %v", err)
	}
	assertStatus(t, db, twice, "submitted", 1)
}

func TestApplyConcurrentCompeting(t *testing.T) {
	db := openCommunityDB(t)
	seedApplyItem(t, db)
	ids := []string{"00000000-0000-7000-8000-00000000a009", "00000000-0000-7000-8000-00000000a010"}
	seedApplySuggestion(t, db, ids[0], 0, "submitted", `{"title":{"before":"Old","after":"A"}}`)
	seedApplySuggestion(t, db, ids[1], 0, "submitted", `{"title":{"before":"Old","after":"B"}}`)

	a, _ := newApplier(t, db)
	errs := make([]error, 2)
	var wg sync.WaitGroup
	for i := range ids {
		wg.Go(func() { errs[i] = a.Approve(context.Background(), ids[i], 1, ApproveOptions{}, adminActor) })
	}
	wg.Wait()

	ok, stale := 0, 0
	for _, err := range errs {
		var te *TransitionError
		switch {
		case err == nil:
			ok++
		case errors.As(err, &te) && te.Code == "stale-before":
			stale++
		default:
			t.Fatalf("unexpected err %v", err)
		}
	}
	if ok != 1 || stale != 1 || auditCount(t, db) != 1 {
		t.Fatalf("ok=%d stale=%d audit=%d", ok, stale, auditCount(t, db))
	}
}

func TestApplyStatusChecks(t *testing.T) {
	db := openCommunityDB(t)
	seedApplyItem(t, db)
	const rejected = "00000000-0000-7000-8000-00000000a011"
	seedApplySuggestion(t, db, rejected, 0, "rejected", `{"title":{"before":"Old","after":"New"}}`)

	a, _ := newApplier(t, db)
	var te *TransitionError
	if err := a.Approve(context.Background(), rejected, 1, ApproveOptions{}, adminActor); !errors.As(err, &te) ||
		te.Code != "invalid-transition" {
		t.Fatalf("rejected err = %#v", err)
	}
	if err := a.Approve(context.Background(), rejected, 7, ApproveOptions{}, adminActor); !errors.As(err, &te) ||
		te.Code != "stale-before" || te.Field != "version" {
		t.Fatalf("version err = %#v", err)
	}
	if err := a.Approve(
		context.Background(),
		"00000000-0000-7000-8000-0000000fffff",
		1,
		ApproveOptions{},
		adminActor,
	); !errors.Is(
		err,
		ErrNotFound,
	) {
		t.Fatalf("missing err = %v", err)
	}
}
