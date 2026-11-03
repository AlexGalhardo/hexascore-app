//go:build integration

package refresh_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/refresh"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

type rig struct {
	t       *testing.T
	db      *database.DB
	ctx     context.Context
	n       int
	missing map[string]bool
	stats   []string
}

func newRig(t *testing.T) *rig {
	t.Helper()

	app := vukemtest.New(
		t,
		bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}},
	)

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	r := &rig{t: t, db: db, ctx: context.Background(), missing: map[string]bool{}}
	r.exec(
		`INSERT INTO data_sources (code, name, licence, url) VALUES ('wikidata', 'Wikidata', 'CC0-1.0', 'https://www.wikidata.org')`,
	)

	return r
}

func (r *rig) exec(q string, args ...any) {
	r.t.Helper()

	if _, err := r.db.ExecContext(r.ctx, q, args...); err != nil {
		r.t.Fatalf("%v\n%s", err, q)
	}
}

func (r *rig) item(cover string, popularity float64) string {
	r.t.Helper()

	r.n++
	id := uuid.NewV7().String()

	var key any
	if cover != "" {
		key = cover
	}

	r.exec(
		`INSERT INTO items (id, type, slug, wikidata_qid, cover_key, popularity_prior) VALUES ($1, 'movie', $2, $3, $4, $5)`,
		id,
		fmt.Sprintf("r-%d", r.n),
		fmt.Sprintf("Q%d", 7000+r.n),
		key,
		popularity,
	)

	return id
}

func (r *rig) stat(_ context.Context, key string) (bool, error) {
	r.stats = append(r.stats, key)

	return !r.missing[key], nil
}

func (r *rig) step(now time.Time) (run.Result, error) {
	r.t.Helper()

	var res run.Result

	r.exec(`INSERT INTO job_runs (id, job, mode, status, params, triggered_by) OVERRIDING SYSTEM VALUE
VALUES (1, 'refresh.existing', 'full', 'running', '{}', 'test') ON CONFLICT DO NOTHING`)

	err := r.db.Tx(r.ctx, func(tx *database.Tx) (e error) {
		res, e = refresh.Step(
			refresh.Config{Stat: r.stat, Batch: 2, Now: func() time.Time { return now }},
		)(
			r.ctx,
			tx,
			run.Run{ID: 1},
		)

		return e
	})

	return res, err
}

func (r *rig) score(id string) (score int, verified bool) {
	r.t.Helper()

	var s *int

	var v *time.Time

	if err := r.db.QueryRowContext(r.ctx, `SELECT quality_score, last_verified_at FROM items WHERE id = $1`, id).
		Scan(&s, &v); err != nil {
		r.t.Fatal(err)
	}

	if s == nil {
		return -1, v != nil
	}

	return *s, v != nil
}

var now = time.Date(2026, 11, 3, 1, 30, 0, 0, time.UTC)

// RFC-001 section 7: cover 25, English synopsis 15, pt-BR and es titles 10, release date 10, two credits 10, a trailer
// 10, provenance confidence mean at least 80 for 20.
func TestQualityScoreAddsUp(t *testing.T) {
	r := newRig(t)

	full := r.item("covers/a.webp", 5)
	r.exec(`UPDATE items SET release_date = '1999-03-31', trailers = '["abc"]'::jsonb,
provenance_v2 = '{"title":{"source":"wikidata","confidence":90},"release_date":{"source":"wikidata","confidence":80}}'::jsonb WHERE id = $1`, full)
	r.exec(
		`INSERT INTO item_translations (item_id, locale, title, title_source, synopsis, synopsis_source, synopsis_licence) VALUES
($1, 'en', 'T', 'wikidata', 'S', 'wikidata', 'CC0-1.0'), ($1, 'pt-BR', 'T', 'wikidata', NULL, NULL, NULL), ($1, 'es', 'T', 'wikidata', NULL, NULL, NULL)`,
		full,
	)

	for i, name := range []string{"Director", "Writer"} {
		eid := uuid.NewV7().String()
		r.exec(
			`INSERT INTO entities (id, kind, slug, default_name) VALUES ($1, 'person', $2, $3)`,
			eid,
			fmt.Sprintf("p%d", i),
			name,
		)
		r.exec(
			`INSERT INTO item_entities (item_id, entity_id, role, source) VALUES ($1, $2, $3, 'wikidata')`,
			full,
			eid,
			[]string{"director", "writer"}[i],
		)
	}

	bare := r.item("", 1)
	coverOnly := r.item("covers/c.webp", 1)

	for range 3 {
		if _, err := r.step(now); err != nil {
			t.Fatal(err)
		}
	}

	if s, v := r.score(full); s != 100 || !v {
		t.Fatalf("complete item: score %d verified %v, want 100", s, v)
	}

	if s, v := r.score(bare); s != 0 || !v {
		t.Fatalf("bare item: score %d verified %v, want 0", s, v)
	}

	if s, _ := r.score(coverOnly); s != 25 {
		t.Fatalf("cover only: score %d, want 25", s)
	}
}

func TestMissingCoverIsReportedAndNotCounted(t *testing.T) {
	r := newRig(t)
	good := r.item("covers/ok.webp", 2)
	lost := r.item("covers/lost.webp", 1)
	r.missing["covers/lost.webp"] = true

	var failed int

	for range 2 {
		res, err := r.step(now)
		if err != nil {
			t.Fatal(err)
		}

		failed += res.Failed
	}

	if failed != 1 {
		t.Fatalf("failed = %d, want 1 (the cover that is not in the bucket)", failed)
	}

	if s, _ := r.score(good); s != 25 {
		t.Fatalf("good cover: %d", s)
	}

	if s, v := r.score(lost); s != 0 || !v {
		t.Fatalf("a cover missing from the bucket earns nothing: score %d verified %v", s, v)
	}

	var n int
	if err := r.db.QueryRowContext(r.ctx, `SELECT count(*) FROM ingest_failures WHERE error_class = 'cover_missing'`).
		Scan(&n); err != nil ||
		n != 1 {
		t.Fatalf("cover_missing failures = %d, %v", n, err)
	}

	// the cover columns are the cover job's: the check never clears them
	var key string
	if err := r.db.QueryRowContext(r.ctx, `SELECT cover_key FROM items WHERE id = $1`, lost).
		Scan(&key); err != nil ||
		key != "covers/lost.webp" {
		t.Fatalf("cover_key = %q, %v: the refresh must not touch it", key, err)
	}
}

// Never-verified items first, items with a cover before items without, then the more popular; a verified item comes
// back only after the refresh interval.
func TestOrderAndInterval(t *testing.T) {
	r := newRig(t)
	plain := r.item("", 9)
	covered := r.item("covers/x.webp", 1)
	popular := r.item("covers/y.webp", 8)

	res, err := r.step(now) // batch of 2
	if err != nil || res.Done {
		t.Fatalf("first page: %+v, %v", res, err)
	}

	if _, v := r.score(popular); !v {
		t.Fatal("the popular covered item must go first")
	}

	if _, v := r.score(covered); !v {
		t.Fatal("covered items come before uncovered ones")
	}

	if _, v := r.score(plain); v {
		t.Fatal("the uncovered item waits for the next page")
	}

	if _, err := r.step(now); err != nil {
		t.Fatal(err)
	}

	asked := len(r.stats)

	// nothing is due the next day
	res, err = r.step(now.AddDate(0, 0, 1))
	if err != nil || !res.Done || res.Targets != 0 || len(r.stats) != asked {
		t.Fatalf("day after: %+v, %v, stats %d -> %d", res, err, asked, len(r.stats))
	}

	// after the interval everything is due again
	later := now.Add(refresh.Every + time.Hour)

	for range 3 {
		if _, err := r.step(later); err != nil {
			t.Fatal(err)
		}
	}

	if len(r.stats) != asked+2 {
		t.Fatalf("covers checked again = %d, want 2", len(r.stats)-asked)
	}
}
