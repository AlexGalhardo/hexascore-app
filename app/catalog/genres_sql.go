package catalog

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// GenreStatement is one SQL statement of the genre repair migration with its arguments.
type GenreStatement struct {
	SQL  string
	Args []any
}

// genreGroups are the rule lists by the item groups GenreFor serves.
func genreGroups() []struct {
	Name  string
	Rules []genreRule
} {
	return []struct {
		Name  string
		Rules []genreRule
	}{{"screen", screenRules}, {"game", gameRules}, {"book", bookRules}}
}

// values builds "VALUES ($1::text, ...), (...)" for rows of arguments and returns the flat argument list.
func values(rows [][]any, casts ...string) (string, []any) {
	var (
		b    strings.Builder
		args []any
	)

	for i, row := range rows {
		if i > 0 {
			b.WriteString(", ")
		}

		b.WriteString("(")

		for j, v := range row {
			if j > 0 {
				b.WriteString(", ")
			}

			args = append(args, v)
			fmt.Fprintf(&b, "$%d::%s", len(args), casts[j])
		}

		b.WriteString(")")
	}

	return "VALUES " + b.String(), args
}

// GenreRepairStatements is genres:canonicalize as plain PostgreSQL, so a migration can apply it without Go code that
// reads rows: the canonical genres and their names, then every source genre link of every item moved to its
// canonical genre (or hidden when no rule maps it), exactly as CanonicalizeGenres does. The rules are the Go ones,
// matched with the same regular expressions (RE2 and PostgreSQL agree on this subset). Only the canonical genres some
// item needs are created, the rest are created by the normaliser when an item needs them. Idempotent.
func GenreRepairStatements() []GenreStatement {
	var names [][]any

	for _, code := range slices.Sorted(maps.Keys(Genres)) {
		n := Genres[code]
		names = append(names, []any{code, n.En, n.PtBR, n.Es})
	}

	var rules [][]any

	for _, g := range genreGroups() {
		for i, r := range g.Rules {
			rules = append(rules, []any{g.Name, i, r.re.String(), r.code})
		}
	}

	namesSQL, namesArgs := values(names, "text", "text", "text", "text")
	rulesSQL, rulesArgs := values(rules, "text", "int", "text", "text")

	return []GenreStatement{
		{SQL: `INSERT INTO data_sources (code, name, licence, url)
VALUES ('hexascore', 'HexaScore', 'proprietary', 'https://hexascore.app') ON CONFLICT (code) DO NOTHING`},
		{SQL: `CREATE TEMP TABLE genre_names (code text PRIMARY KEY, en text, pt text, es text)`},
		{SQL: `INSERT INTO genre_names (code, en, pt, es) ` + namesSQL, Args: namesArgs},
		{SQL: `CREATE TEMP TABLE genre_rules (grp text, ord int, pattern text, code text)`},
		{SQL: `INSERT INTO genre_rules (grp, ord, pattern, code) ` + rulesSQL, Args: rulesArgs},
		// The first matching rule of the item group wins; a label no rule maps keeps a NULL code (hidden).
		{SQL: `CREATE TEMP TABLE genre_relink AS
SELECT DISTINCT ON (s.item_id, s.entity_id) s.item_id, s.entity_id, s.source, s.position, r.code
FROM (
  SELECT ie.item_id, ie.entity_id, ie.source, ie.position, e.default_name AS label,
    CASE WHEN i.type IN ('movie', 'series') THEN 'screen' ELSE i.type::text END AS grp
  FROM item_entities ie
  JOIN entities e ON e.id = ie.entity_id
  JOIN items i ON i.id = ie.item_id
  WHERE ie.role = 'genre' AND e.provenance->>'canonical' IS DISTINCT FROM 'genre'
) s
LEFT JOIN genre_rules r ON r.grp = s.grp AND lower(btrim(s.label)) ~ r.pattern
ORDER BY s.item_id, s.entity_id, r.ord`},
		// Take over the genre entity whose code or slug is the canonical code, the one with the code first.
		{SQL: `UPDATE entities e SET code = m.code, default_name = m.en,
  provenance = e.provenance || '{"canonical":"genre"}', updated_at = now()
FROM (
  SELECT DISTINCT ON (n.code) n.code, n.en, ent.id
  FROM genre_names n JOIN entities ent ON ent.kind = 'genre' AND (ent.code = n.code OR ent.slug = n.code)
  WHERE n.code IN (SELECT code FROM genre_relink WHERE code IS NOT NULL)
  ORDER BY n.code, (ent.code = n.code) DESC NULLS LAST
) m
WHERE e.id = m.id AND (e.code IS DISTINCT FROM m.code OR e.default_name <> m.en
  OR e.provenance->>'canonical' IS DISTINCT FROM 'genre')`},
		{SQL: `INSERT INTO entities (id, kind, slug, code, default_name, provenance)
SELECT gen_random_uuid(), 'genre', n.code, n.code, n.en, '{"canonical":"genre","default_name":"hexascore"}'::jsonb
FROM genre_names n
WHERE n.code IN (SELECT code FROM genre_relink WHERE code IS NOT NULL)
  AND NOT EXISTS (SELECT 1 FROM entities e WHERE e.kind = 'genre' AND (e.code = n.code OR e.slug = n.code))`},
		{SQL: `INSERT INTO entity_translations (entity_id, locale, name, name_source)
SELECT e.id, l.locale, l.name, 'hexascore'
FROM genre_names n
JOIN entities e ON e.kind = 'genre' AND e.code = n.code
CROSS JOIN LATERAL (VALUES ('en', n.en), ('pt-BR', n.pt), ('es', n.es)) AS l(locale, name)
WHERE n.code IN (SELECT code FROM genre_relink WHERE code IS NOT NULL)
ON CONFLICT (entity_id, locale) DO UPDATE SET name = excluded.name, name_source = excluded.name_source`},
		{SQL: `INSERT INTO item_entities (item_id, entity_id, role, position, source)
SELECT x.item_id, e.id, 'genre', x.position, x.source
FROM genre_relink x JOIN entities e ON e.kind = 'genre' AND e.code = x.code
WHERE x.code IS NOT NULL ON CONFLICT DO NOTHING`},
		{SQL: `INSERT INTO item_genre_sources (item_id, entity_id, source, position, code)
SELECT item_id, entity_id, source, position, code FROM genre_relink
ON CONFLICT (item_id, entity_id) DO UPDATE SET code = excluded.code`},
		{SQL: `DELETE FROM item_entities ie USING genre_relink x
WHERE ie.item_id = x.item_id AND ie.entity_id = x.entity_id AND ie.role = 'genre'`},
		{SQL: `DROP TABLE genre_relink`},
		{SQL: `DROP TABLE genre_rules`},
		{SQL: `DROP TABLE genre_names`},
	}
}
