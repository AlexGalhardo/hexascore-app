//go:build integration

package ingest_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/community"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/normalize"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/store"
)

const itemID = "00000000-0000-0000-0000-0000000000aa"

var edited = time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

func seed(t *testing.T) *database.DB {
	t.Helper()

	env := map[string]string{"DATABASE_URL": pgtest.Database(t, "")}
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: env})

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	for _, q := range []string{
		`INSERT INTO data_sources (code, name, licence, url) VALUES
 ('wikidata','Wikidata','CC0-1.0','https://www.wikidata.org'),
 ('community','HexaScore community','HEXASCORE-API-TERMS','https://hexascore.app')`,
		`INSERT INTO items (id, type, slug, wikidata_qid) VALUES ('` + itemID + `','movie','matrix','Q83495')`,
		`INSERT INTO item_translations (item_id, locale, title, title_source) VALUES ('` + itemID + `','en','Matrix, The','community')`,
		`INSERT INTO catalog_audit (occurred_at, actor_role, source, target_type, target_id, field, before, after)
VALUES ('` + edited.Format(time.RFC3339) + `','admin','community','item','` + itemID + `','translations.en.title','"The Matrix"','"Matrix, The"')`,
	} {
		if _, err := db.ExecContext(t.Context(), q); err != nil {
			t.Fatal(err)
		}
	}

	return db
}

func built(t *testing.T) *normalize.Item {
	t.Helper()

	all := map[string]json.RawMessage{}

	raw, err := os.ReadFile("../../internal/ingest/testdata/normalize_wikidata_payloads.json")
	if err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal(raw, &all); err != nil {
		t.Fatal(err)
	}

	item, err := normalize.BuildItem(
		"Q83495",
		[]store.Record{{Source: "wikidata", ExternalID: "Q83495", Payload: all["Q83495"]}},
		edited,
	)
	if err != nil {
		t.Fatal(err)
	}

	return item
}

func lockCount(t *testing.T, db *database.DB, q string) int {
	t.Helper()

	var n int
	if err := db.QueryRowContext(context.Background(), q).Scan(&n); err != nil {
		t.Fatal(err)
	}

	return n
}

// Fake clock: day 179 the source is skipped and a system suggestion opens;
// day 181 a divergence still opens one and never overwrites (FR-06-4).
func TestCommunityFieldLock(t *testing.T) {
	db := seed(t)
	gate := community.IngestGate{DB: db}

	for _, days := range []int{179, 181} {
		item := built(t)

		res, err := normalize.ApplyCommunity(t.Context(), gate, itemID, item, edited.AddDate(0, 0, days))
		if err != nil {
			t.Fatalf("day %d: %v", days, err)
		}

		if !item.Skip["translations.en.title"] || len(res.Suggested) != 1 {
			t.Errorf("day %d: skip = %v, suggested = %v", days, item.Skip, res.Suggested)
		}
	}

	// Both runs propose the same value: one open suggestion, not two.
	if n := lockCount(t, db, `SELECT count(*) FROM suggestions WHERE author_id IS NULL AND status = 'submitted'
AND kind = 'edit' AND target_id = '`+itemID+`' AND changes -> 'translations.en.title' ->> 'after' = 'The Matrix'
AND changes -> 'translations.en.title' ->> 'before' = 'Matrix, The' AND submitted_at IS NOT NULL`); n != 1 {
		t.Errorf("open system suggestions = %d, want 1", n)
	}

	if n := lockCount(
		t,
		db,
		`SELECT count(*) FROM suggestion_events WHERE actor_role = 'system' AND to_status = 'submitted'`,
	); n != 1 {
		t.Errorf("system submit events = %d, want 1", n)
	}

	if n := lockCount(
		t,
		db,
		`SELECT count(*) FROM item_translations WHERE title = 'Matrix, The' AND title_source = 'community'`,
	); n != 1 {
		t.Error("the community title was overwritten")
	}
}

func TestCommunityFieldLockSameValueExpires(t *testing.T) {
	db := seed(t)
	if _, err := db.ExecContext(t.Context(), `UPDATE item_translations SET title = 'The Matrix'`); err != nil {
		t.Fatal(err)
	}

	gate := community.IngestGate{DB: db}

	for days, wantSkip := range map[int]bool{179: true, 181: false} {
		item := built(t)

		if _, err := normalize.ApplyCommunity(t.Context(), gate, itemID, item, edited.AddDate(0, 0, days)); err != nil {
			t.Fatal(err)
		}

		if item.Skip["translations.en.title"] != wantSkip {
			t.Errorf("day %d: skip = %v", days, item.Skip)
		}
	}

	if n := lockCount(t, db, `SELECT count(*) FROM suggestions`); n != 0 {
		t.Errorf("suggestions = %d, want 0", n)
	}
}
