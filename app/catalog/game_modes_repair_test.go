//go:build integration

package catalog_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

var modeLabels = []string{
	"single-player video game", "multiplayer video game", "co-op mode", "player versus player",
	"multiplayer online game", "shared/splitscreen multiplayer", "asymmetric unmapped mode",
}

// The migration SQL gives every game the canonical mode GameModeFor picks, hides a label no rule maps, keeps the
// source link and names each canonical mode in three locales.
func TestGameModeRepairSQLMatchesTheGoRules(t *testing.T) {
	ctx := context.Background()
	db := entityDB(t)

	type fixture struct{ item, entity string }

	fx := map[string]fixture{}

	for i, label := range modeLabels {
		item := newItem(t, db, "game", fmt.Sprintf("mode-game-%d", i), fmt.Sprintf("Q%d", 700+i), "public")
		ent := newEntity(t, db, "game_mode", fmt.Sprintf("source-mode-%d", i))

		if _, err := db.ExecContext(
			ctx,
			`UPDATE entities SET default_name = $2 WHERE id = $1`,
			ent,
			label,
		); err != nil {
			t.Fatal(err)
		}

		if _, err := db.ExecContext(ctx, `INSERT INTO item_entities (item_id, entity_id, role, position, source)
VALUES ($1, $2, 'game_mode', 0, 'wikidata')`, item, ent); err != nil {
			t.Fatal(err)
		}

		fx[label] = fixture{item, ent}
	}

	for range 2 { // the second pass proves it is idempotent
		for _, st := range catalog.GameModeRepairStatements() {
			if _, err := db.ExecContext(ctx, st.SQL, st.Args...); err != nil {
				t.Fatalf("%v\n%s", err, st.SQL)
			}
		}
	}

	needed := map[string]bool{}

	for _, label := range modeLabels {
		f := fx[label]
		want, mapped := catalog.GameModeFor("game", label)

		var got []string

		rows, err := db.QueryContext(ctx, `SELECT e.code FROM item_entities ie JOIN entities e ON e.id = ie.entity_id
WHERE ie.item_id = $1 AND ie.role = 'game_mode' AND e.code IS NOT NULL`, f.item)
		if err != nil {
			t.Fatal(err)
		}

		for rows.Next() {
			var code string
			if err := rows.Scan(&code); err != nil {
				t.Fatal(err)
			}

			got = append(got, code)
		}

		if err := rows.Close(); err != nil {
			t.Fatal(err)
		}

		switch {
		case mapped && (len(got) != 1 || got[0] != want):
			t.Errorf("%q: modes %v, want [%s]", label, got, want)
		case !mapped && len(got) != 0:
			t.Errorf("%q: modes %v, want none (hidden)", label, got)
		}

		if mapped {
			needed[want] = true
		}

		var kept int
		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM item_genre_sources WHERE item_id = $1 AND entity_id = $2`,
			f.item, f.entity).
			Scan(&kept); err != nil ||
			kept != 1 {
			t.Errorf("%q: source link kept %d times (%v), want once", label, kept, err)
		}
	}

	var pt string
	if err := db.QueryRowContext(ctx, `SELECT t.name FROM entity_translations t JOIN entities e ON e.id = t.entity_id
WHERE e.kind = 'game_mode' AND e.code = 'single-player' AND t.locale = 'pt-BR'`).Scan(&pt); err != nil || pt != "Um jogador" {
		t.Errorf("pt-BR name of single-player = %q (%v)", pt, err)
	}

	var canon int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM entities WHERE kind = 'game_mode' AND provenance->>'canonical' = 'game_mode'`).
		Scan(&canon); err != nil ||
		canon != len(needed) {
		t.Errorf(
			"canonical modes = %d (%v), want %d: only the modes some game needs are created",
			canon,
			err,
			len(needed),
		)
	}

	facets, err := catalog.FacetsOf(ctx, db, "game", "game_mode", "en", nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range facets {
		if f.Canonical && f.Name != catalog.GameModes[f.Slug].En {
			t.Errorf("facet %s named %q, want %q", f.Slug, f.Name, catalog.GameModes[f.Slug].En)
		}
	}
}

// The normaliser canonicalises the game modes of each item it saves.
func TestCanonicalizeItemFoldsGameModes(t *testing.T) {
	ctx := context.Background()
	db := entityDB(t)

	item := newItem(t, db, "game", "new-game", "Q990", "public")
	ent := newEntity(t, db, "game_mode", "single-player-video-game")

	if _, err := db.ExecContext(
		ctx,
		`UPDATE entities SET default_name = 'single-player video game' WHERE id = $1`,
		ent,
	); err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(ctx, `INSERT INTO item_entities (item_id, entity_id, role, position, source)
VALUES ($1, $2, 'game_mode', 0, 'wikidata')`, item, ent); err != nil {
		t.Fatal(err)
	}

	if err := catalog.CanonicalizeItem(ctx, db, item); err != nil {
		t.Fatal(err)
	}

	var name string
	if err := db.QueryRowContext(ctx, `SELECT coalesce(t.name, e.default_name) FROM item_entities ie
JOIN entities e ON e.id = ie.entity_id LEFT JOIN entity_translations t ON t.entity_id = e.id AND t.locale = 'en'
WHERE ie.item_id = $1 AND ie.role = 'game_mode'`, item).Scan(&name); err != nil || name != "Single player" {
		t.Errorf("mode = %q (%v), want Single player", name, err)
	}
}
