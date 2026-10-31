//go:build integration

package moderation

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/schema"

	"github.com/AlexGalhardo/hexascore-app/app/community"
	"github.com/AlexGalhardo/hexascore-app/database/migrations"
)

var base = time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC)

type seed struct {
	kind, target, domain, locale, status, changes, meta string
	author, assignee                                    int64
	ageHours                                            float64
}

func openDB(t *testing.T) *database.DB {
	t.Helper()
	ctx := context.Background()
	db, err := database.Open(ctx, pgtest.Database(t, ""), database.WithDriver(postgres.Driver()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	m, err := schema.NewMigrator(db, migrations.All())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Up(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO users (id, name, email, password, created_at, updated_at)
VALUES (1,'A','a@x.io','x',now(),now()),(2,'B','b@x.io','x',now(),now())`); err != nil {
		t.Fatal(err)
	}

	return db
}

var seq int

func insert(t *testing.T, db *database.DB, s seed) string {
	t.Helper()
	seq++
	id := fmt.Sprintf("00000000-0000-7000-8000-%012d", seq)
	def := func(v, d string) string {
		if v == "" {
			return d
		}

		return v
	}
	kind := def(s.kind, "edit")
	target := sql.NullString{String: def(s.target, "00000000-0000-0000-0000-0000000000aa"), Valid: kind != "create"}
	var author, assignee any
	if s.author != 0 {
		author = s.author
	}
	if s.assignee != 0 {
		assignee = s.assignee
	}
	_, err := db.ExecContext(
		context.Background(),
		`INSERT INTO suggestions
(id, author_id, kind, target_type, target_id, target_domain, locale, changes, status, assigned_to, submitted_at, meta)
VALUES ($1,$2,$3,'item',$4,$5,$6,$7::jsonb,$8,$9,$10,$11::jsonb)`,
		id,
		author,
		kind,
		target,
		def(s.domain, "movie"),
		def(s.locale, "en"),
		def(s.changes, `{"title":{"before":"a","after":"b"}}`),
		def(
			s.status,
			"submitted",
		),
		assignee,
		base.Add(-time.Duration(s.ageHours*float64(time.Hour))),
		def(s.meta, "{}"),
	)
	if err != nil {
		t.Fatal(err)
	}

	return id
}

func ids(p Page) map[string]bool {
	out := map[string]bool{}
	for _, r := range p.Rows {
		out[r.ID] = true
	}

	return out
}

func TestQueueFilters(t *testing.T) {
	db := openDB(t)
	q := Queue{DB: db, Now: func() time.Time { return base }}
	ctx := context.Background()

	green := insert(t, db, seed{ageHours: 2, author: 1})
	amber := insert(t, db, seed{ageHours: 18, author: 2, locale: "pt-BR", kind: "create", domain: "game"})
	red := insert(
		t,
		db,
		seed{ageHours: 30, status: "in_review", assignee: 1, changes: `{"year":{"before":1,"after":2}}`},
	)
	comp := insert(t, db, seed{ageHours: 3, meta: `{"competing":true}`})
	stale := insert(t, db, seed{ageHours: 4, meta: `{"stale":true}`})
	unreach := insert(t, db, seed{ageHours: 5, meta: `{"source_status":"unreachable"}`})
	insert(t, db, seed{status: "applied"})
	insert(t, db, seed{status: "draft"})

	all, err := q.List(ctx, Filter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(all.Rows) != 6 {
		t.Fatalf("default queue has %d rows, want the 6 open ones", len(all.Rows))
	}
	if all.Rows[0].ID != red || all.Rows[0].Age != AgeRed || ids(all)[green] == false {
		t.Fatalf("oldest first with red badge expected, got %+v", all.Rows[0])
	}

	for name, tc := range map[string]struct {
		f    Filter
		want []string
	}{
		"status":    {Filter{Status: "in_review"}, []string{red}},
		"kind":      {Filter{Kind: "create"}, []string{amber}},
		"domain":    {Filter{Domain: "game"}, []string{amber}},
		"locale":    {Filter{Locale: "pt-BR"}, []string{amber}},
		"field":     {Filter{Field: "year"}, []string{red}},
		"author":    {Filter{Author: 2}, []string{amber}},
		"assignee":  {Filter{Assignee: 1}, []string{red}},
		"age green": {Filter{Age: AgeGreen}, []string{green, comp, stale, unreach}},
		"age amber": {Filter{Age: AgeAmber}, []string{amber}},
		"age red":   {Filter{Age: AgeRed}, []string{red}},
		"competing": {Filter{Competing: true}, []string{comp}},
		"stale":     {Filter{Stale: true}, []string{stale}},
	} {
		got, err := q.List(ctx, tc.f)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if len(got.Rows) != len(tc.want) {
			t.Fatalf("%s: %d rows, want %d", name, len(got.Rows), len(tc.want))
		}
		for _, w := range tc.want {
			if !ids(got)[w] {
				t.Fatalf("%s: missing %s", name, w)
			}
		}
	}

	row := map[string]Row{}
	for _, r := range all.Rows {
		row[r.ID] = r
	}
	if !row[comp].Competing || !row[stale].Stale || !row[unreach].Unreachable || row[green].Competing {
		t.Fatalf("badges wrong: %+v", row)
	}
}

func TestQueueKeysetPagination(t *testing.T) {
	db := openDB(t)
	q := Queue{DB: db, Now: func() time.Time { return base }, PageSize: 2}
	for i := range 5 {
		insert(t, db, seed{ageHours: float64(10 - i)})
	}
	seen := map[string]bool{}
	var f Filter
	pages := 0
	for {
		p, err := q.List(context.Background(), f)
		if err != nil {
			t.Fatal(err)
		}
		pages++
		for _, r := range p.Rows {
			if seen[r.ID] {
				t.Fatalf("row %s repeated", r.ID)
			}
			seen[r.ID] = true
		}
		if p.Next == "" {
			break
		}
		f.After = p.Next
	}
	if len(seen) != 5 || pages != 3 {
		t.Fatalf("saw %d rows in %d pages, want 5 in 3", len(seen), pages)
	}
	if _, err := q.List(context.Background(), Filter{After: "garbage"}); !errors.Is(err, ErrBadCursor) {
		t.Fatalf("bad cursor: %v", err)
	}
	// 36 characters but not a UUID: a 404, never a SQL cast error.
	if _, err := q.List(
		context.Background(),
		Filter{After: "1_" + strings.Repeat("z", 36)},
	); !errors.Is(
		err,
		ErrBadCursor,
	) {
		t.Fatalf("non-uuid cursor: %v", err)
	}
}

func TestClaimRelease(t *testing.T) {
	db := openDB(t)
	q := Queue{DB: db, Now: func() time.Time { return base }}
	ctx := context.Background()
	id := insert(t, db, seed{})
	admin1 := community.Actor{ID: sql.NullInt64{Int64: 1, Valid: true}, Role: "admin"}
	admin2 := community.Actor{ID: sql.NullInt64{Int64: 2, Valid: true}, Role: "admin"}

	if err := q.Claim(ctx, id, 1, admin1); err != nil {
		t.Fatal(err)
	}
	var status string
	var assignee sql.NullInt64
	var claimed sql.NullTime
	read := func() {
		if err := db.QueryRowContext(ctx, `SELECT status, assigned_to, claimed_at FROM suggestions WHERE id=$1`, id).
			Scan(&status, &assignee, &claimed); err != nil {
			t.Fatal(err)
		}
	}
	read()
	if status != "in_review" || assignee.Int64 != 1 || !claimed.Valid {
		t.Fatalf("after claim: %s %v %v", status, assignee, claimed)
	}

	// The second admin still holds the version they loaded: 409 stale-before.
	var te *community.TransitionError
	if err := q.Claim(ctx, id, 1, admin2); !errors.As(err, &te) || te.Status != 409 || te.Code != "stale-before" {
		t.Fatalf("second claim: %v, want 409 stale-before", err)
	}

	if err := q.Release(
		ctx,
		id,
		2,
		community.Actor{ID: admin2.ID, Role: "moderator"},
	); !errors.Is(
		err,
		ErrNotAssignee,
	) {
		t.Fatalf("moderator releasing another claim: %v, want ErrNotAssignee", err)
	}
	if err := q.Release(ctx, id, 2, admin2); err != nil {
		t.Fatalf("admin release: %v", err)
	}
	read()
	if status != "submitted" || assignee.Valid || claimed.Valid {
		t.Fatalf("after release: %s %v %v", status, assignee, claimed)
	}
}
