package catalog

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// DQ levels (SDD 01 §7.6): an error fails `catalog:check`, a warning is listed.
const (
	LevelError = "error"
	LevelWarn  = "warn"
)

const (
	dqSampleSize = 20
	slugRegex    = `^[a-z0-9]+(-[a-z0-9]+)*$`
)

// DQAssert is one data-quality assert: Query returns the violating ids as a
// single text column.
type DQAssert struct {
	ID, Level, Description, Query string
}

// DQResult is the outcome of one assert; Sample holds up to 20 violating ids.
type DQResult struct {
	ID          string   `json:"id"`
	Level       string   `json:"level"`
	Description string   `json:"description"`
	Count       int      `json:"count"`
	Sample      []string `json:"sample,omitempty"`
}

// DQReport aggregates a run. Errors counts violations of error-level asserts
// and Warnings those of warn-level ones.
type DQReport struct {
	Errors   int        `json:"errors"`
	Warnings int        `json:"warnings"`
	Checks   []DQResult `json:"checks"`
}

// allowedRoleValues renders the §5.2 role table as a SQL VALUES list built
// from roleRules, so DQ-08 and RoleCompatible cannot drift apart.
func allowedRoleValues() string {
	var rows []string
	for role, rule := range roleRules {
		itemTypes := rule.itemTypes
		if itemTypes == nil {
			itemTypes = allTypes
		}
		for _, kind := range rule.kinds {
			for _, typ := range itemTypes {
				rows = append(rows, fmt.Sprintf("('%s','%s','%s')", role, kind, typ))
			}
		}
	}
	sort.Strings(rows)
	return strings.Join(rows, ",")
}

// DQAsserts returns the SDD 01 §7.6 asserts DQ-01..DQ-14.
func DQAsserts() []DQAssert {
	return []DQAssert{
		{
			"DQ-01", LevelError, "item without an en translation with a non-empty title",
			`SELECT i.id::text FROM items i WHERE NOT EXISTS (
SELECT 1 FROM item_translations t WHERE t.item_id = i.id AND t.locale = 'en' AND btrim(t.title) <> '')`,
		},
		{
			"DQ-02", LevelError, "edition isbn13 invalid or duplicated",
			`SELECT id::text FROM editions WHERE isbn13 IS NOT NULL AND NOT isbn13_is_valid(isbn13)
UNION SELECT e.id::text FROM editions e WHERE e.isbn13 IS NOT NULL
AND (SELECT count(*) FROM editions o WHERE o.isbn13 = e.isbn13) > 1`,
		},
		{
			"DQ-03", LevelWarn, "orphan entity older than 7 days (no items, editions or entity links)",
			`SELECT e.id::text FROM entities e WHERE e.created_at < now() - interval '7 days'
AND NOT EXISTS (SELECT 1 FROM item_entities ie WHERE ie.entity_id = e.id)
AND NOT EXISTS (SELECT 1 FROM editions ed WHERE ed.publisher_id = e.id OR ed.imprint_id = e.id)
AND NOT EXISTS (SELECT 1 FROM entity_links l WHERE l.from_id = e.id OR l.to_id = e.id)`,
		},
		{
			"DQ-04", LevelError, "slug violates the format or equals a redirect old_slug in its namespace",
			`SELECT id::text FROM items WHERE slug !~ '` + slugRegex + `' OR length(slug) > 120
UNION SELECT id::text FROM entities WHERE slug !~ '` + slugRegex + `' OR length(slug) > 120
UNION SELECT i.id::text FROM items i JOIN catalog_redirects r
  ON r.scope = 'item' AND r.namespace = i.type AND r.old_slug = i.slug
UNION SELECT e.id::text FROM entities e JOIN catalog_redirects r
  ON r.scope = 'entity' AND r.namespace = e.kind AND r.old_slug = e.slug`,
		},
		{
			"DQ-05", LevelError, "redirect target missing or itself a redirect source (chain)",
			`SELECT r.target_id::text FROM catalog_redirects r
WHERE (r.scope = 'item' AND NOT EXISTS (SELECT 1 FROM items i WHERE i.id = r.target_id))
OR (r.scope = 'entity' AND NOT EXISTS (SELECT 1 FROM entities e WHERE e.id = r.target_id))
OR EXISTS (SELECT 1 FROM catalog_redirects n WHERE n.scope = r.scope AND n.old_id = r.target_id)`,
		},
		{
			"DQ-06", LevelError, "synopsis or description without source, licence or (Wikipedia) source URL",
			`SELECT item_id::text FROM item_translations WHERE synopsis IS NOT NULL
AND (synopsis_source IS NULL OR synopsis_licence IS NULL
  OR (synopsis_source = 'wikipedia' AND synopsis_source_url IS NULL))
UNION SELECT entity_id::text FROM entity_translations WHERE description IS NOT NULL
AND (description_source IS NULL OR description_licence IS NULL
  OR (description_source = 'wikipedia' AND description_source_url IS NULL))`,
		},
		{
			"DQ-07", LevelError, "cover without source, or attribution-licensed cover without attribution",
			`SELECT i.id::text FROM items i LEFT JOIN data_sources s ON s.code = i.cover_source
WHERE i.cover_key IS NOT NULL AND (i.cover_source IS NULL
  OR (i.cover_attribution IS NULL AND (s.attribution IS NOT NULL OR s.licence LIKE 'CC-BY%')))
UNION SELECT id::text FROM editions WHERE cover_key IS NOT NULL AND cover_source IS NULL`,
		},
		{
			"DQ-08", LevelError, "item_entities role incompatible with entity kind and item type",
			`SELECT ie.item_id::text || ':' || ie.role || ':' || ie.entity_id::text FROM item_entities ie
JOIN items i ON i.id = ie.item_id JOIN entities e ON e.id = ie.entity_id
WHERE (ie.role, e.kind, i.type) NOT IN (` + allowedRoleValues() + `)`,
		},
		{
			"DQ-09", LevelWarn, "native entity (no Wikidata QID), listed for review",
			`SELECT id::text FROM entities WHERE wikidata_qid IS NULL`,
		},
		{
			"DQ-10", LevelWarn, "book work without editions, or edition with defaulted language",
			`SELECT i.id::text FROM items i WHERE i.type = 'book'
AND NOT EXISTS (SELECT 1 FROM editions e WHERE e.work_id = i.id)
UNION SELECT id::text FROM editions WHERE language = 'und'`,
		},
		{
			"DQ-11", LevelError, "provenance value is not a data_sources code",
			`SELECT i.id::text FROM items i, jsonb_each_text(i.provenance) p
WHERE p.value NOT IN (SELECT code FROM data_sources)
UNION SELECT e.id::text FROM entities e, jsonb_each_text(e.provenance) p
WHERE p.value NOT IN (SELECT code FROM data_sources)`,
		},
		{
			"DQ-12", LevelWarn, "entity_stats refreshed more than 2 hours ago",
			`SELECT entity_id::text || ':' || type FROM entity_stats WHERE refreshed_at < now() - interval '2 hours'`,
		},
		{
			"DQ-13", LevelWarn, "released item with unexpected null date, or announced item long past its day date",
			`SELECT id::text FROM items WHERE release_status = 'released' AND release_date IS NULL
AND release_precision <> 'unknown'
UNION SELECT id::text FROM items WHERE release_status = 'announced' AND release_precision = 'day'
AND release_date < current_date - 30`,
		},
		{
			"DQ-14", LevelWarn, "based_on 2-cycle between two items",
			`SELECT a.from_item_id::text || ':' || a.to_item_id::text FROM item_links a
JOIN item_links b ON b.from_item_id = a.to_item_id AND b.to_item_id = a.from_item_id
AND b.relation = a.relation WHERE a.relation = 'based_on' AND a.from_item_id < a.to_item_id`,
		},
	}
}

// RunDQ executes every assert and reports counts and a sample of violating ids.
func RunDQ(ctx context.Context, db DQQuerier) (DQReport, error) {
	report := DQReport{Checks: []DQResult{}}
	for _, a := range DQAsserts() {
		q := `WITH v(id) AS (` + a.Query + `) SELECT (SELECT count(*) FROM v),
coalesce((SELECT string_agg(id, ',') FROM (SELECT id FROM v ORDER BY id LIMIT ` +
			strconv.Itoa(dqSampleSize) + `) s), '')`
		var sample string
		res := DQResult{ID: a.ID, Level: a.Level, Description: a.Description}
		if err := db.QueryRowContext(ctx, q).Scan(&res.Count, &sample); err != nil {
			return DQReport{}, fmt.Errorf("%s: %w", a.ID, err)
		}
		if sample != "" {
			res.Sample = strings.Split(sample, ",")
		}
		if a.Level == LevelError {
			report.Errors += res.Count
		} else {
			report.Warnings += res.Count
		}
		report.Checks = append(report.Checks, res)
	}
	return report, nil
}

// DQQuerier is the subset of *database.DB the asserts need.
type DQQuerier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}
