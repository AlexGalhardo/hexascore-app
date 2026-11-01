//go:build integration

package ingest_test

import (
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/database"
)

const (
	itemGapA = "0192f1c6-0000-7000-8000-0000000000c1"
	itemGapB = "0192f1c6-0000-7000-8000-0000000000c2"
	itemGapC = "0192f1c6-0000-7000-8000-0000000000c3"
)

// seedSourceRun builds a finished fetch.<source> run and returns its id.
func seedSourceRun(t *testing.T, db *database.DB, source string) int64 {
	t.Helper()

	var id int64
	if err := db.QueryRowContext(t.Context(), `INSERT INTO job_runs (job, source, mode, status, triggered_by, started_at,
  finished_at, created_at)
VALUES ('fetch.' || $1::text, $1, 'full', 'succeeded', 'schedule', now() - interval '1 hour', now() - interval '10 minutes',
  now() - interval '1 hour') RETURNING id`, source).
		Scan(&id); err != nil {
		t.Fatal(err)
	}

	return id
}

func execAll(t *testing.T, db *database.DB, run int64, stmts ...string) {
	t.Helper()

	rep := strings.NewReplacer("$1", itoa(int(run)))

	for _, s := range stmts {
		if _, err := db.ExecContext(t.Context(), rep.Replace(s)); err != nil {
			t.Fatalf("%v: %s", err, s)
		}
	}
}

// A non-wikidata record the bad run inserted is deleted: the Wikidata record of the same QID must be queued again
// and the cover that source gave the item must go, or the item keeps data nothing backs any more.
func TestRollbackOfNonWikidataInsertRequeuesWikidataAndClearsCover(t *testing.T) {
	db := newDB(t)
	target := seedSourceRun(t, db, "igdb")

	execAll(t, db, target,
		`INSERT INTO data_sources (code, name, licence, url) VALUES
('wikidata','Wikidata','CC0-1.0','https://www.wikidata.org'), ('igdb','IGDB','x','https://www.igdb.com'),
('tmdb','TMDB','x','https://www.themoviedb.org') ON CONFLICT DO NOTHING`,
		`INSERT INTO source_records (source, external_id, payload, hash, fetched_at, last_run_id, normalized_at)
VALUES ('wikidata','Q71','{}', repeat('a',64), now() - interval '5 days', 0, now() - interval '4 days'),
       ('igdb','Q71','{}', repeat('b',64), now() - interval '30 minutes', $1, now())`,
		`INSERT INTO items (id, type, slug, wikidata_qid, original_title, created_at, cover_key, cover_thumb_key,
  cover_source, cover_attribution)
VALUES ('`+itemGapA+`', 'game', 'gap-a', 'Q71', 'Gap A', now() - interval '5 days', 'covers/gap-a.webp',
  'covers/gap-a-t.webp', 'igdb', '{"by":"igdb"}')`,
		// another source's cover on another item stays
		`INSERT INTO source_records (source, external_id, payload, hash, fetched_at, last_run_id, normalized_at)
VALUES ('igdb','Q72','{}', repeat('c',64), now() - interval '30 minutes', $1, now())`,
		`INSERT INTO items (id, type, slug, wikidata_qid, original_title, created_at, cover_key, cover_source)
VALUES ('`+itemGapB+`', 'game', 'gap-b', 'Q72', 'Gap B', now() - interval '5 days', 'covers/gap-b.webp', 'tmdb')`,
	)

	runRollback(t, db, target)

	if n := count(t, db, `SELECT count(*) FROM source_records WHERE source = 'igdb'`); n != 0 {
		t.Fatalf("igdb records left = %d", n)
	}

	if n := count(t, db, `SELECT count(*) FROM source_records WHERE source = 'wikidata' AND external_id = 'Q71'
AND normalized_at IS NULL`); n != 1 {
		t.Error("the wikidata record of the QID must be pending again")
	}

	if n := count(t, db, `SELECT count(*) FROM items WHERE id = $1 AND cover_key IS NULL AND cover_thumb_key IS NULL
AND cover_source IS NULL AND cover_attribution IS NULL`, itemGapA); n != 1 {
		t.Error("the igdb cover must be cleared")
	}

	if n := count(
		t,
		db,
		`SELECT count(*) FROM cover_gc WHERE key IN ('covers/gap-a.webp','covers/gap-a-t.webp')`,
	); n != 2 {
		t.Errorf("released cover keys = %d, want 2", n)
	}

	if n := count(t, db, `SELECT count(*) FROM items WHERE id = $1 AND cover_key IS NOT NULL`, itemGapB); n != 1 {
		t.Error("a cover of another source must stay")
	}
}

// A restored record goes back to the pre-run state: its checked_at is the one it had then, not the bad run's, and
// the cover that source gave stays cleared so the covers job re-picks it from the restored payload.
func TestRollbackRestoreKeepsPreRunCheckedAt(t *testing.T) {
	db := newDB(t)
	target := seedSourceRun(t, db, "tmdb")

	execAll(
		t,
		db,
		target,
		`INSERT INTO data_sources (code, name, licence, url) VALUES
('tmdb','TMDB','x','https://www.themoviedb.org') ON CONFLICT DO NOTHING`,
		`INSERT INTO source_records (source, external_id, payload, hash, fetched_at, checked_at, last_run_id, normalized_at)
VALUES ('tmdb','Q73','{"v":"new"}', repeat('b',64), now() - interval '30 minutes', now() - interval '30 minutes', $1, now())`,
		`INSERT INTO source_record_history (source, external_id, hash, payload, fetched_at, checked_at, superseded_at,
  superseded_by_run)
VALUES ('tmdb','Q73', repeat('a',64), '{"v":"old"}', now() - interval '100 days', now() - interval '60 days',
  now() - interval '30 minutes', $1)`,
		`INSERT INTO items (id, type, slug, wikidata_qid, original_title, created_at, cover_key, cover_source)
VALUES ('`+itemGapC+`', 'movie', 'gap-c', 'Q73', 'Gap C', now() - interval '100 days', 'covers/gap-c.webp', 'tmdb')`,
	)

	runRollback(t, db, target)

	if n := count(t, db, `SELECT count(*) FROM source_records WHERE source = 'tmdb' AND external_id = 'Q73'
AND payload->>'v' = 'old' AND checked_at < now() - interval '59 days'`); n != 1 {
		t.Error("the restored record must carry its pre-run checked_at")
	}

	if n := count(t, db, `SELECT count(*) FROM items WHERE id = $1 AND cover_key IS NULL AND cover_source IS NULL`,
		itemGapC); n != 1 {
		t.Error("the tmdb cover must be cleared")
	}
}

// History written before the checked_at column existed has none; the fetch time stands in for it.
func TestRollbackRestoreFallsBackToFetchedAt(t *testing.T) {
	db := newDB(t)
	target := seedSourceRun(t, db, "tmdb")

	execAll(
		t,
		db,
		target,
		`INSERT INTO data_sources (code, name, licence, url) VALUES
('tmdb','TMDB','x','https://www.themoviedb.org') ON CONFLICT DO NOTHING`,
		`INSERT INTO source_records (source, external_id, payload, hash, fetched_at, checked_at, last_run_id, normalized_at)
VALUES ('tmdb','Q74','{"v":"new"}', repeat('b',64), now() - interval '30 minutes', now() - interval '30 minutes', $1, now())`,
		`INSERT INTO source_record_history (source, external_id, hash, payload, fetched_at, superseded_at, superseded_by_run)
VALUES ('tmdb','Q74', repeat('a',64), '{"v":"old"}', now() - interval '100 days', now() - interval '30 minutes', $1)`,
	)

	runRollback(t, db, target)

	if n := count(t, db, `SELECT count(*) FROM source_records WHERE external_id = 'Q74'
AND checked_at < now() - interval '99 days'`); n != 1 {
		t.Error("without a stored checked_at the restored record falls back to its fetched_at")
	}
}
