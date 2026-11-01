//go:build integration

package legacyimport_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/database/legacyimport"
)

// seedSynthetic adds n items to the rehearsal fixture with overlapping names, slugs that collide
// (Ação / Acao), QID-backed and native entities sharing a name, non-Latin names, blank names and books
// with valid, duplicated and invalid ISBNs.
func seedSynthetic(t *testing.T, db *database.DB, n int) {
	t.Helper()

	genres := []string{"Action", "Ação", "Acao", "RPG", "shooter", " Action ", "", "Sci-Fi", "action-2"}
	creators := []string{"Valve", "Ana Writer", "宮崎駿", "Bob  Smith", "Bob Smith", "Studio X", "Q Corp"}
	platforms := []string{"Windows", "Linux", "Switch", "windows"}
	countries := []string{"United States", "Brazil", "日本"}
	types := []string{"game", "movie", "series", "book"}

	var items, titles, records strings.Builder

	pick := func(list []string, g, k int) string { return list[(g*k+g/3)%len(list)] }

	for g := 1; g <= n; g++ {
		id := fmt.Sprintf("01910000-0000-7000-8000-%012d", g)
		meta := fmt.Sprintf(`{"genres": [%q, %q], "platforms": [%q], "creators": [%q, %q], "countries": [%q]}`,
			pick(genres, g, 3), pick(genres, g, 5), pick(platforms, g, 7), pick(creators, g, 11), pick(creators, g, 13),
			pick(countries, g, 17))

		ext := `{}`

		switch g % 5 {
		case 1:
			ext = fmt.Sprintf(
				`{"isbn13": "978%09d%d"}`,
				100000000+g%40,
				isbnCheck(fmt.Sprintf("978%09d", 100000000+g%40)),
			)
		case 2:
			ext = `{"isbn13": "12345"}`
		}

		sep := ","
		if g == 1 {
			sep = ""
		}

		fmt.Fprintf(&items, "%s('%s', '%s', 'syn-%d', 'Q%d', '%s', '%s')", sep, id, types[g%4], g, 5000+g, ext, meta)
		fmt.Fprintf(&titles, "%s('%s', 'en', 'Synthetic %d')", sep, id, g)

		if g%3 != 0 {
			fmt.Fprintf(
				&records,
				`%s('wikidata', 'Q%d', '{"claims": {"P136": [{"value": "Q70"}], "P57": [{"value": "Q71"}],
 "P170": [{"value": "Q71"}], "P50": [{"value": "Q71"}], "P178": [{"value": "Q72"}], "P400": [{"value": "Q73"}]},
 "refs": {"Q70": "Action", "Q71": "Bob Smith", "Q72": "Valve", "Q73": "Windows"}}', repeat('d', 64), '%s')`,
				sep,
				5000+g,
				id,
			)
		}
	}

	stmts := []string{
		`INSERT INTO legacy_import.items (id, type, slug, wikidata_qid, external_ids, meta) VALUES ` + items.String(),
		`INSERT INTO legacy_import.item_translations (item_id, locale, title) VALUES ` + titles.String(),
	}
	if records.Len() > 0 {
		stmts = append(
			stmts,
			`INSERT INTO legacy_import.source_records (source, external_id, payload, hash, item_id) VALUES `+
				strings.TrimPrefix(records.String(), ","),
		)
	}

	for _, q := range stmts {
		if _, err := db.ExecContext(t.Context(), q); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
}

func isbnCheck(twelve string) int {
	sum := 0

	for i, c := range twelve {
		d := int(c - '0')
		if i%2 == 1 {
			d *= 3
		}

		sum += d
	}

	return (10 - sum%10) % 10
}

// entityDump renders what the entities stage wrote without surrogate ids. withSlugs adds the entity slugs,
// and the work behind each edition, which depend on the order batches commit in (a parallel run may hand out "-2"
// differently, and a shared ISBN13 goes to whichever work commits first).
func entityDump(t *testing.T, db *database.DB, withSlugs bool) string {
	t.Helper()

	slug, edWork := "''", "''"
	if withSlugs {
		slug, edWork = "e.slug", "w.slug || '|' || ed.title"
	}

	queries := []string{
		`SELECT coalesce(string_agg(concat_ws('|', e.kind, e.wikidata_qid, e.default_name, ` + slug + `), E'\n' ORDER BY
  e.kind, coalesce(e.wikidata_qid, ''), e.default_name, ` + slug + `), '') FROM entities e`,
		`SELECT coalesce(string_agg(concat_ws('|', i.slug, ie.role, ie.position, ie.source, e.kind, e.wikidata_qid,
  e.default_name, ` + slug + `), E'\n' ORDER BY i.slug, ie.role, e.kind, coalesce(e.wikidata_qid, ''), e.default_name), '')
FROM item_entities ie JOIN items i ON i.id = ie.item_id JOIN entities e ON e.id = ie.entity_id`,
		`SELECT coalesce(string_agg(concat_ws('|', ` + edWork + `, ed.isbn13, ed.language, ed.source), E'
'
  ORDER BY ed.isbn13), '') FROM editions ed JOIN items w ON w.id = ed.work_id`,
		`SELECT status || ' ' || rows || ' ' || coalesce(last_key, '') FROM legacy_import_runs WHERE stage = 'entities'`,
	}

	var out strings.Builder

	for _, q := range queries {
		var s string
		if err := db.QueryRowContext(t.Context(), q).Scan(&s); err != nil {
			t.Fatalf("dump: %v", err)
		}

		out.WriteString(s + "\n--\n")
	}

	return out.String()
}

func runEntities(t *testing.T, db *database.DB, batch, workers int, reference bool) {
	t.Helper()

	stages, err := legacyimport.Select(legacyimport.Stages(), []string{"catalog", "entities"})
	if err != nil {
		t.Fatal(err)
	}

	for i := range stages {
		if reference && stages[i].Name == "entities" {
			stages[i].Batch = legacyimport.ReferenceEntitiesBatch()
		}
	}

	r := &legacyimport.Runner{DB: db, Batch: batch, Out: io.Discard, Now: time.Now}
	r.Workers = workers

	if err := r.Run(t.Context(), stages); err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestEntitiesSetBasedMatchesReference(t *testing.T) {
	t.Parallel()

	_, oldDB := fixtureDB(t)
	_, newDB := fixtureDB(t)

	for _, db := range []*database.DB{oldDB, newDB} {
		seedSynthetic(t, db, 400)
	}

	runEntities(t, oldDB, 37, 1, true)
	runEntities(t, newDB, 37, 1, false)

	want, got := entityDump(t, oldDB, true), entityDump(t, newDB, true)
	if want != got {
		t.Fatalf("set-based stage differs from the per-item reference\n--- reference\n%s\n--- set-based\n%s", want, got)
	}

	if !strings.Contains(want, "|Action|") || !strings.Contains(want, "978") {
		t.Fatalf("fixture too weak to prove anything:\n%s", want)
	}
}

func TestEntitiesWorkersMatchSequential(t *testing.T) {
	t.Parallel()

	_, seqDB := fixtureDB(t)
	_, parDB := fixtureDB(t)

	for _, db := range []*database.DB{seqDB, parDB} {
		seedSynthetic(t, db, 400)
	}

	runEntities(t, seqDB, 25, 1, false)
	runEntities(t, parDB, 25, 4, false)

	// Slugs are left out: which of two colliding names gets the "-2" depends on who commits first.
	want, got := entityDump(t, seqDB, false), entityDump(t, parDB, false)
	if want != got {
		t.Fatalf("4 workers differ from 1\n--- 1 worker\n%s\n--- 4 workers\n%s", want, got)
	}

	var dup int
	if err := parDB.QueryRowContext(t.Context(), `SELECT count(*) FROM (SELECT kind, default_name FROM entities
WHERE wikidata_qid IS NULL GROUP BY 1, 2 HAVING count(*) > 1) d`).Scan(&dup); err != nil || dup != 0 {
		t.Fatalf("native entities duplicated by racing workers: %d (%v)", dup, err)
	}
}

// TestEntitiesWorkersResume kills a parallel run with a failing batch and resumes it: the checkpoint must
// never pass a batch that did not commit, and the finished import must equal an uninterrupted one.
func TestEntitiesWorkersResume(t *testing.T) {
	t.Parallel()

	_, wantDB := fixtureDB(t)
	_, db := fixtureDB(t)

	for _, d := range []*database.DB{wantDB, db} {
		seedSynthetic(t, d, 300)
	}

	runEntities(t, wantDB, 20, 1, false)

	stages, err := legacyimport.Select(legacyimport.Stages(), []string{"catalog", "entities"})
	if err != nil {
		t.Fatal(err)
	}

	var calls atomic.Int32

	for i := range stages {
		if stages[i].Name != "entities" {
			continue
		}

		inner := stages[i].Batch
		stages[i].Batch = func(ctx context.Context, tx *database.Tx, after string, n int) (string, int, error) {
			if calls.Add(1) == 6 {
				return "", 0, errors.New("killed")
			}

			return inner(ctx, tx, after, n)
		}
	}

	r := &legacyimport.Runner{DB: db, Batch: 20, Workers: 4, Out: io.Discard, Now: time.Now}
	if err := r.Run(t.Context(), stages); err == nil {
		t.Fatal("expected the killed run to fail")
	}

	// Every batch below the checkpoint must be in place.
	var key string
	if err := db.QueryRowContext(t.Context(), `SELECT coalesce(last_key, '') FROM legacy_import_runs WHERE stage = 'entities'`).
		Scan(&key); err != nil {
		t.Fatal(err)
	}

	var missing int
	if err := db.QueryRowContext(t.Context(), `SELECT count(*) FROM legacy_import.items i WHERE i.id::text <= $1
AND jsonb_array_length(coalesce(i.meta->'genres', '[]')) > 0 AND NOT EXISTS (SELECT 1 FROM item_entities ie WHERE ie.item_id = i.id)`,
		key).Scan(&missing); err != nil || missing != 0 {
		t.Fatalf("checkpoint %q is past %d unfinished items (%v)", key, missing, err)
	}

	runEntities(t, db, 20, 4, false)

	if want, got := entityDump(t, wantDB, false), entityDump(t, db, false); want != got {
		t.Fatalf("resumed run differs from an uninterrupted one\n--- want\n%s\n--- got\n%s", want, got)
	}
}

func userStagesDump(t *testing.T, db *database.DB) string {
	t.Helper()

	queries := []string{
		`SELECT coalesce(string_agg(concat_ws('|', u.legacy_id, r.item_id, r.c1, r.c2, r.c3, r.c4, r.c5, r.c6), E'
'
  ORDER BY u.legacy_id, r.item_id), '') FROM ratings r JOIN users u ON u.id = r.user_id`,
		`SELECT coalesce(string_agg(concat_ws('|', u.legacy_id, a.item_id, a.old_c, a.new_c), E'
'
  ORDER BY u.legacy_id, a.item_id), '') FROM rating_migration_audit a JOIN users u ON u.id = a.user_id`,
		`SELECT coalesce(string_agg(concat_ws('|', l.title, l.type, l.item_count, li.item_id, li.position), E'
'
  ORDER BY l.title, li.position), '') FROM lists l LEFT JOIN list_items li ON li.list_id = l.id`,
		`SELECT coalesce(string_agg(concat_ws('|', u.legacy_id, s.item_id, s.status), E'
' ORDER BY u.legacy_id, s.item_id), '')
FROM item_status s JOIN users u ON u.id = s.user_id`,
		`SELECT coalesce(string_agg(concat_ws('|', u.legacy_id, t.item_id, t.main_minutes), E'
' ORDER BY u.legacy_id, t.item_id), '')
FROM time_to_beat t JOIN users u ON u.id = t.user_id`,
		`SELECT coalesce(string_agg(concat_ws('|', c.subject, c.message, c.user_id IS NULL), E'
' ORDER BY c.subject, c.message), '')
FROM contact_messages c`,
		`SELECT coalesce(string_agg(after::text, E'
'), '') FROM score_actions WHERE kind = 'legacy_rounding'`,
		`SELECT status || ' ' || rows || ' ' || coalesce(last_key, '') FROM legacy_import_runs WHERE stage = 'social'`,
		`SELECT status || ' ' || rows || ' ' || coalesce(last_key, '') FROM legacy_import_runs WHERE stage = 'library'`,
	}

	var out strings.Builder

	for _, q := range queries {
		var s string
		if err := db.QueryRowContext(t.Context(), q).Scan(&s); err != nil {
			t.Fatalf("dump: %v", err)
		}

		out.WriteString(s + "\n--\n")
	}

	return out.String()
}

func TestSocialLibraryWorkersMatchSequential(t *testing.T) {
	t.Parallel()

	dumps := make([]string, 2)

	for i, workers := range []int{1, 4} {
		_, db := fixtureDB(t)

		stages, err := legacyimport.Select(legacyimport.Stages(), []string{"users", "catalog", "social", "library"})
		if err != nil {
			t.Fatal(err)
		}

		r := &legacyimport.Runner{DB: db, Batch: 1, Workers: workers, Out: io.Discard, Now: time.Now}
		if err := r.Run(t.Context(), stages); err != nil {
			t.Fatalf("run with %d workers: %v", workers, err)
		}

		dumps[i] = userStagesDump(t, db)
	}

	if dumps[0] != dumps[1] {
		t.Fatalf("4 workers differ from 1\n--- 1 worker\n%s\n--- 4 workers\n%s", dumps[0], dumps[1])
	}

	if !strings.Contains(dumps[0], "u01|") {
		t.Fatalf("fixture too weak:\n%s", dumps[0])
	}
}
