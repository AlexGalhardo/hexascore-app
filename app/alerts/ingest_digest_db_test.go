//go:build integration

package alerts

import (
	"testing"
	"time"
)

// The summary counts only the window, lists the top errors, and a quiet window sends nothing (except at 08:00).
func TestIngestDigestFromTables(t *testing.T) {
	db := openDigestDB(t)
	ctx := t.Context()
	at := time.Date(2026, 11, 2, 17, 30, 0, 0, time.UTC) // 14:30 in Sao Paulo
	f := newFakeTelegram(t)
	n := newNotifier(t, f, errorChat, nil, at)

	if sent, err := SendIngestDigest(ctx, db, n, at, false); err != nil || sent {
		t.Fatalf("quiet window: sent %v, err %v", sent, err)
	}

	for _, q := range []string{
		`INSERT INTO data_sources (code, name, licence, url) VALUES ('wikidata','Wikidata','CC0-1.0','https://x')`,
		`INSERT INTO items (id, type, slug, wikidata_qid, created_at) VALUES (gen_random_uuid(), 'game','new-in-window','Q1', '2026-11-02T17:10:00Z'),
  (gen_random_uuid(), 'game','old','Q2','2026-11-02T10:00:00Z')`,
		`INSERT INTO job_runs (job, source, mode, status, finished_at, changed, failed, rate_limited, triggered_by)
VALUES ('fetch', 'wikidata', 'incremental', 'succeeded', '2026-11-02T17:20:00Z', 40, 2, 1, 't'),
  ('covers', NULL, 'incremental', 'succeeded', '2026-11-02T17:25:00Z', 7, 0, 0, 't'),
  ('fetch', 'wikidata', 'full', 'succeeded', '2026-11-02T09:00:00Z', 500, 9, 9, 't')`,
		`INSERT INTO ingest_failures (source, external_id, last_run_id, error_class, error, last_seen) VALUES
  ('wikidata','Q1',1,'http_5xx','x','2026-11-02T17:15:00Z'), ('wikidata','Q2',1,'http_5xx','x','2026-11-02T17:16:00Z'),
  ('tmdb','9',1,'timeout','x','2026-11-02T17:17:00Z'), ('tmdb','8',1,'timeout','x','2026-11-02T01:00:00Z')`,
		`INSERT INTO ingest_cost_daily (day, llm_usd) VALUES ('2026-11-01', 1.0), ('2026-11-02', 1.0), ('2026-10-31', 50)`,
	} {
		if _, err := db.ExecContext(ctx, q); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}

	s, err := BuildIngestSummary(ctx, db, at)
	if err != nil {
		t.Fatal(err)
	}

	if s.Created != 1 || s.Updated != 40 || s.Failed != 2 || s.Covers != 7 || s.RateLimited != 1 {
		t.Errorf("counts = %+v", s)
	}

	if len(s.Errors) != 2 || s.Errors[0] != (IngestError{"wikidata", "http_5xx", 2}) || s.Errors[1].Source != "tmdb" {
		t.Errorf("errors = %+v", s.Errors)
	}

	if s.LLMMonthUSD != 2 || s.LLMProjectedUSD != 30 { // 2 USD in 2 days of a 30-day month
		t.Errorf("llm = %v / %v", s.LLMMonthUSD, s.LLMProjectedUSD)
	}

	if sent, err := SendIngestDigest(ctx, db, n, at, false); err != nil || !sent {
		t.Fatalf("busy window: sent %v, err %v", sent, err)
	}

	drain(t, n)

	if got := f.messages(); len(got) != 1 || !got[0].Silent {
		t.Fatalf("messages = %+v", got)
	}
}
