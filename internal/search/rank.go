package search

import (
	"strings"
)

// Item SQL placeholders, fixed for every variant so each is always referenced:
// $1 raw text, $2 match text, $3 types text[] or NULL, $4 popularity p99, $5 UI locale, $6 allowed sources text[]
// (NULL = site, every source), $7 limit, $8 offset, $9 ISBN-13 or NULL, $10 excluded terms (or ”).
//
// Tiers are strict (SDD 07 §7.3): rank = 10·E + 5·P + (S >= 0.2 ? 2 + S : 0) + F + 0.3·Pop, with the non-exact
// maximum 9.3 < 10, non-prefix 4.3 < 5 and FTS-only 1.3 < 2.2.
const itemRankSQL = `10 * e::int + 5 * p::int + CASE WHEN sim >= 0.2 THEN 2 + sim ELSE 0 END + f
  + 0.3 * coalesce(least(ln(1 + i.popularity_prior) / nullif(ln(1 + $4::float8), 0), 1), 0)`

const itemOrder = `rank DESC, pop DESC, release_date DESC NULLS LAST, id`

// guard keeps every placeholder referenced whatever the variant, so Postgres can type them all.
const itemGuard = `AND coalesce($1::text, '') IS NOT NULL AND coalesce($7::int, 0) >= 0
   AND coalesce($8::int, 0) >= 0 AND coalesce($9::text, '') IS NOT NULL AND coalesce($10::text, '') IS NOT NULL`

var ftsConfigs = map[string]string{"en": "hs_en", "pt-BR": "hs_pt", "es": "hs_es"}

var searchLocales = []string{"en", "pt-BR", "es"}

const trigramGate = `length(hs_norm($2)) >= 3`

// itemCandidates builds the UNION ALL of the candidate rows: (item_id, n, t, k, eid, isbn, f, x) with n the
// normalised title, t the title, k the origin ('t' translation, 'a' alternative, 'e' edition), f the FTS rank and
// x a forced exact match (ISBN).
func itemCandidates(locale string, api, isbn bool) string {
	if isbn {
		return `SELECT work_id AS item_id, title_norm AS n, title AS t, 'e' AS k, id AS eid, isbn13::text AS isbn,
    0::real AS f, true AS x
  FROM editions WHERE isbn13 = $9::text AND ($6::text[] IS NULL OR source = ANY($6))`
	}

	col := "search_tsv"
	if api {
		col = "search_tsv_api"
	}

	locales := searchLocales
	if locale != "" {
		locales = []string{locale}
	}

	var branches []string

	for _, loc := range locales {
		cfg := ftsConfigs[loc]
		branches = append(
			branches,
			`SELECT item_id, title_norm AS n, title AS t, 't' AS k, NULL::uuid AS eid, NULL::text AS isbn,
    ts_rank_cd(`+col+`, websearch_to_tsquery('`+cfg+`', $1), 32) AS f, false AS x
  FROM item_translations
  WHERE locale = '`+loc+`' AND `+col+` @@ websearch_to_tsquery('`+cfg+`', $1)
    AND ($6::text[] IS NULL OR title_source = ANY($6))`,
		)
	}

	branches = append(branches,
		`SELECT item_id, title_norm, title, 't', NULL::uuid, NULL::text, 0::real, false
  FROM item_translations
  WHERE `+trigramGate+` AND hs_norm($2) <% title_norm AND ($6::text[] IS NULL OR title_source = ANY($6))`,
		`SELECT item_id, title_norm, title, 'a', NULL::uuid, NULL::text, 0::real, false
  FROM item_alt_titles
  WHERE `+trigramGate+` AND hs_norm($2) <% title_norm AND ($6::text[] IS NULL OR source = ANY($6))`,
		`SELECT work_id, title_norm, title, 'e', id, isbn13::text, 0::real, false
  FROM editions
  WHERE `+trigramGate+` AND hs_norm($2) <% title_norm AND ($6::text[] IS NULL OR source = ANY($6))`,
	)

	return strings.Join(branches, "\n  UNION ALL ")
}

// itemSQL assembles the ranked item query. grouped returns the top 8 per type with per-type totals; otherwise a
// flat page with the overall total.
func itemSQL(locale string, api, isbn, grouped, excluded bool) string {
	exclusion := ""
	if excluded {
		col := "search_tsv"
		if api {
			col = "search_tsv_api"
		}

		var conds []string
		for _, cfg := range []string{"hs_en", "hs_pt", "hs_es"} {
			conds = append(conds, col+` @@ websearch_to_tsquery('`+cfg+`', $10)`)
		}

		exclusion = ` AND NOT EXISTS (SELECT 1 FROM item_translations x WHERE x.item_id = i.id AND (` +
			strings.Join(conds, " OR ") + `))`
	}

	ranked := `WITH c AS (
  ` + itemCandidates(locale, api, isbn) + `
), b AS (
  SELECT DISTINCT ON (item_id) item_id, t, k, eid, isbn
  FROM c
  ORDER BY item_id, (x OR n = hs_norm($2)) DESC, word_similarity(hs_norm($2), n) DESC, f DESC
), s AS (
  SELECT item_id, bool_or(x OR n = hs_norm($2)) AS e,
    bool_or(hs_norm($2) <> '' AND starts_with(n, hs_norm($2))) AS p,
    max(word_similarity(hs_norm($2), n)) AS sim, max(f) AS f
  FROM c GROUP BY item_id
), r AS (
  SELECT i.id, i.type, i.slug, i.popularity_prior AS pop, i.release_date, dt.title,
    b.t AS mt, b.k, b.eid, b.isbn, ` + itemRankSQL + ` AS rank
  FROM s
  JOIN b USING (item_id)
  JOIN items i ON i.id = s.item_id
  JOIN LATERAL (
    SELECT title FROM item_translations t
    WHERE t.item_id = i.id AND ($6::text[] IS NULL OR t.title_source = ANY($6))
    ORDER BY (t.locale = $5::text) DESC, (t.locale = 'en') DESC, t.locale
    LIMIT 1
  ) dt ON true
  WHERE i.visibility = 'public' AND ($3::text[] IS NULL OR i.type = ANY($3))
   ` + exclusion + `
   ` + itemGuard + `
)
`

	cols := `id::text, type, slug, title, mt, k, eid::text, isbn, rank`
	if grouped {
		return ranked + `SELECT ` + cols + `, total FROM (
  SELECT r.*, row_number() OVER (PARTITION BY type ORDER BY ` + itemOrder + `) AS rn,
    count(*) OVER (PARTITION BY type) AS total
  FROM r
) g WHERE rn <= 8 ORDER BY type, rn`
	}

	return ranked + `SELECT ` + cols + `, count(*) OVER () AS total FROM r ORDER BY ` + itemOrder + ` LIMIT $7 OFFSET $8`
}

// Entity placeholders: $1 match text, $2 kinds text[], $3 UI locale, $4 allowed sources (NULL = site), $5 limit,
// $6 offset. Same tiers as items without FTS (SDD 07 §7.4). entities.default_name has no recorded source, so the
// API skips it and only shows translations from allowed sources (ADR 0006, fail closed).
func entitySQL(api bool) string {
	defaultName := `
  UNION ALL SELECT id, name_norm, default_name FROM entities
  WHERE kind IN ('person','organization','franchise') AND ` + entityGate + ` AND hs_norm($1) <% name_norm`
	display := `coalesce((SELECT name FROM entity_translations t WHERE t.entity_id = e.id
      ORDER BY (t.locale = $3::text) DESC, (t.locale = 'en') DESC, t.locale LIMIT 1), e.default_name)`

	if api {
		defaultName = ""
		display = `(SELECT name FROM entity_translations t WHERE t.entity_id = e.id AND t.name_source = ANY($4)
      ORDER BY (t.locale = $3::text) DESC, (t.locale = 'en') DESC, t.locale LIMIT 1)`
	}

	return `WITH c AS (
  SELECT entity_id AS id, name_norm AS n, name AS t FROM entity_translations
  WHERE ` + entityGate + ` AND hs_norm($1) <% name_norm AND ($4::text[] IS NULL OR name_source = ANY($4))` + defaultName + `
), s AS (
  SELECT id, bool_or(n = hs_norm($1)) AS e, bool_or(hs_norm($1) <> '' AND starts_with(n, hs_norm($1))) AS p,
    max(word_similarity(hs_norm($1), n)) AS sim
  FROM c GROUP BY id
), r AS (
  SELECT e.id, e.kind, e.slug, ` + display + ` AS name,
    10 * s.e::int + 5 * s.p::int + CASE WHEN sim >= 0.2 THEN 2 + sim ELSE 0 END AS rank,
    (SELECT coalesce(sum(item_count), 0) FROM entity_stats st WHERE st.entity_id = e.id) AS cnt
  FROM s JOIN entities e ON e.id = s.id
  WHERE e.kind = ANY($2::text[])
)
SELECT id::text, kind, slug, name, rank, cnt, count(*) OVER () AS total FROM r WHERE name IS NOT NULL
ORDER BY rank DESC, cnt DESC, id LIMIT $5 OFFSET $6`
}

const entityGate = `length(hs_norm($1)) >= 3`
