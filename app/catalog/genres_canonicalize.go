package catalog

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/uuid"
)

// taxonomy is a set of canonical entities that source labels are folded into: the genres of every category and the
// game modes of the games. Items link to the canonical entities; the source labels (Wikidata, IGDB) are kept in
// item_genre_sources and never shown.
type taxonomy struct {
	kind  string                                 // entities.kind, item_entities.role and the provenance "canonical" value
	names map[string]GenreName                   // canonical code -> display names in the three locales
	match func(typ, label string) (string, bool) // the rules: first match wins, false hides the label
}

var (
	genreTaxonomy = taxonomy{"genre", Genres, GenreFor}
	modeTaxonomy  = taxonomy{"game_mode", GameModes, GameModeFor}
)

// GenreReport is what genres:canonicalize did (or would do with --dry-run).
type GenreReport struct {
	Created, Adopted int            // canonical genre entities inserted, or existing slugs taken over
	Links            int            // source genre links read
	Mapped, Hidden   int            // links moved to a canonical genre / hidden without one
	Unmapped         map[string]int // hidden labels by "type: label", for widening the rules
}

var errDryRun = errors.New("dry run")

// CanonicalizeGenres replaces every source genre link (role genre on a non-canonical entity) with a link to its
// canonical genre and moves the source link to item_genre_sources; a label no rule maps is moved there too, so it
// stops showing (owner 2026-11-02). Idempotent: a second run finds no source links. dryRun rolls everything back.
func CanonicalizeGenres(ctx context.Context, db *database.DB, dryRun bool) (GenreReport, error) {
	rep := GenreReport{Unmapped: map[string]int{}}

	err := db.Tx(ctx, func(tx *database.Tx) error {
		ids, err := ensureCanonicalGenres(ctx, tx, genreTaxonomy, &rep)
		if err != nil {
			return err
		}

		if err := relinkGenres(ctx, tx, genreTaxonomy, ids, &rep); err != nil {
			return err
		}

		if dryRun {
			return errDryRun
		}

		return nil
	})
	if errors.Is(err, errDryRun) {
		err = nil
	}

	return rep, err
}

// ensureCanonicalGenres creates the canonical genre entities with their three names, or takes over an existing
// genre entity whose slug is the code, and returns their ids by code.
func ensureCanonicalGenres(
	ctx context.Context, tx database.Executor, tax taxonomy, rep *GenreReport,
) (map[string]string, error) {
	if _, err := tx.ExecContext(ctx, `INSERT INTO data_sources (code, name, licence, url)
VALUES ('hexascore', 'HexaScore', 'proprietary', 'https://hexascore.app') ON CONFLICT (code) DO NOTHING`); err != nil {
		return nil, fmt.Errorf("genres: data source: %w", err)
	}

	ids := map[string]string{}
	for _, code := range slices.Sorted(maps.Keys(tax.names)) {
		id, err := ensureCanonicalGenre(ctx, tx, tax, code, rep)
		if err != nil {
			return nil, err
		}

		ids[code] = id
	}

	return ids, nil
}

// ensureCanonicalGenre creates one canonical genre with its three names, or takes over an existing genre entity whose
// slug is the code, and returns its id. The data source row is created by the caller.
func ensureCanonicalGenre(
	ctx context.Context, tx database.Executor, tax taxonomy, code string, rep *GenreReport,
) (string, error) {
	n := tax.names[code]

	var id string
	err := tx.QueryRowContext(ctx, `SELECT id::text FROM entities WHERE kind = $2 AND (code = $1 OR slug = $1)
ORDER BY (code = $1) DESC NULLS LAST LIMIT 1`, code, tax.kind).Scan(&id)

	switch {
	case err == nil:
		res, err := tx.ExecContext(ctx, `UPDATE entities SET code = $2, default_name = $3,
  provenance = provenance || jsonb_build_object('canonical', $4::text), updated_at = now()
WHERE id = $1 AND (code IS DISTINCT FROM $2 OR default_name <> $3 OR provenance->>'canonical' IS DISTINCT FROM $4)`,
			id, code, n.En, tax.kind)
		if err != nil {
			return "", fmt.Errorf("genres: adopt %s: %w", code, err)
		}

		if c, _ := res.RowsAffected(); c > 0 {
			rep.Adopted++
		}
	case errors.Is(err, sql.ErrNoRows):
		id = uuid.NewV7().String()
		if _, err := tx.ExecContext(ctx, `INSERT INTO entities (id, kind, slug, code, default_name, provenance)
VALUES ($1, $4, $2, $2, $3, jsonb_build_object('canonical', $4::text, 'default_name', 'hexascore'))`,
			id, code, n.En, tax.kind); err != nil {
			return "", fmt.Errorf("genres: create %s: %w", code, err)
		}

		rep.Created++
	default:
		return "", fmt.Errorf("genres: find %s: %w", code, err)
	}

	for locale, name := range map[string]string{"en": n.En, "pt-BR": n.PtBR, "es": n.Es} {
		if _, err := tx.ExecContext(ctx, `INSERT INTO entity_translations (entity_id, locale, name, name_source)
VALUES ($1, $2, $3, 'hexascore')
ON CONFLICT (entity_id, locale) DO UPDATE SET name = excluded.name, name_source = excluded.name_source`,
			id, locale, name); err != nil {
			return "", fmt.Errorf("genres: name %s %s: %w", code, locale, err)
		}
	}

	return id, nil
}

type sourceGenre struct {
	item, entity, typ, label, source string
	position                         int
}

func relinkGenres(
	ctx context.Context, tx database.Executor, tax taxonomy, ids map[string]string, rep *GenreReport,
) error {
	links, err := sourceGenres(ctx, tx, tax, "")
	if err != nil {
		return err
	}

	rep.Links = len(links)
	for _, g := range links {
		mapped, err := relinkOne(ctx, tx, tax, ids, g)
		if err != nil {
			return err
		}

		if mapped {
			rep.Mapped++
		} else {
			rep.Hidden++
			rep.Unmapped[g.typ+": "+g.label]++
		}
	}

	return nil
}

// relinkOne moves one source genre link to its canonical genre, or hides it when no rule maps its label.
func relinkOne(
	ctx context.Context, tx database.Executor, tax taxonomy, ids map[string]string, g sourceGenre,
) (bool, error) {
	code, ok := tax.match(g.typ, g.label)
	var codeArg any // NULL marks a hidden label
	if ok {
		codeArg = code
		if _, err := tx.ExecContext(ctx, `INSERT INTO item_entities (item_id, entity_id, role, position, source)
VALUES ($1, $2, $5, $3, $4) ON CONFLICT DO NOTHING`, g.item, ids[code], g.position, g.source, tax.kind); err != nil {
			return false, fmt.Errorf("genres: link %s: %w", g.item, err)
		}
	}

	if _, err := tx.ExecContext(ctx, `INSERT INTO item_genre_sources (item_id, entity_id, source, position, code)
VALUES ($1, $2, $3, $4, $5) ON CONFLICT (item_id, entity_id) DO UPDATE SET code = excluded.code`,
		g.item, g.entity, g.source, g.position, codeArg); err != nil {
		return false, fmt.Errorf("genres: keep source %s: %w", g.item, err)
	}

	if _, err := tx.ExecContext(
		ctx,
		`DELETE FROM item_entities WHERE item_id = $1 AND entity_id = $2 AND role = $3`,
		g.item,
		g.entity,
		tax.kind,
	); err != nil {
		return false, fmt.Errorf("genres: unlink %s: %w", g.item, err)
	}

	return ok, nil
}

// CanonicalizeItem is what the normaliser runs after it saves an item's relations: the source genres and game
// modes it just linked (Wikidata labels such as "comedy film" or "single-player video game") become canonical links
// at once, so a new ingest never shows them and nobody has to run anything again. It creates only the canonical
// entities the item needs.
func CanonicalizeItem(ctx context.Context, ex database.Executor, itemID string) error {
	for _, tax := range []taxonomy{genreTaxonomy, modeTaxonomy} {
		if err := canonicalizeItem(ctx, ex, tax, itemID); err != nil {
			return err
		}
	}

	return nil
}

func canonicalizeItem(ctx context.Context, ex database.Executor, tax taxonomy, itemID string) error {
	links, err := sourceGenres(ctx, ex, tax, itemID)
	if err != nil || len(links) == 0 {
		return err
	}

	if _, err := ex.ExecContext(ctx, `INSERT INTO data_sources (code, name, licence, url)
VALUES ('hexascore', 'HexaScore', 'proprietary', 'https://hexascore.app') ON CONFLICT (code) DO NOTHING`); err != nil {
		return fmt.Errorf("genres: data source: %w", err)
	}

	ids := map[string]string{}

	for _, g := range links {
		code, ok := tax.match(g.typ, g.label)
		if !ok || ids[code] != "" {
			continue
		}

		if ids[code], err = ensureCanonicalGenre(ctx, ex, tax, code, &GenreReport{}); err != nil {
			return err
		}
	}

	for _, g := range links {
		if _, err := relinkOne(ctx, ex, tax, ids, g); err != nil {
			return err
		}
	}

	return nil
}

// sourceGenres reads the genre links that still point at a non-canonical entity.
func sourceGenres(ctx context.Context, tx database.Executor, tax taxonomy, item string) ([]sourceGenre, error) {
	rows, err := tx.QueryContext(
		ctx,
		`SELECT ie.item_id::text, ie.entity_id::text, i.type, e.default_name, ie.source, ie.position
FROM item_entities ie
JOIN entities e ON e.id = ie.entity_id
JOIN items i ON i.id = ie.item_id
WHERE ie.role = $2 AND e.provenance->>'canonical' IS DISTINCT FROM $2
  AND ($1 = '' OR ie.item_id = NULLIF($1, '')::uuid)
ORDER BY ie.item_id, ie.position`,
		item,
		tax.kind,
	)
	if err != nil {
		return nil, fmt.Errorf("genres: source links: %w", err)
	}
	defer rows.Close()

	var links []sourceGenre
	for rows.Next() {
		var g sourceGenre
		if err := rows.Scan(&g.item, &g.entity, &g.typ, &g.label, &g.source, &g.position); err != nil {
			return nil, fmt.Errorf("genres: scan: %w", err)
		}
		links = append(links, g)
	}

	return links, rows.Err()
}
