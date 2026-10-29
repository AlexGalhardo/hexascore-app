package legacyimport

import "time"

// RoundRating is the legacy rounding of one category (SDD 02 FR-02-13, §5): the nearest multiple of 5,
// clamped to 5..100. Integer inputs never tie. roundSQL is the same rule in SQL; the tests check both.
func RoundRating(c int) int {
	return min(100, max(5, 5*((c+2)/5)))
}

func roundSQL(col string) string {
	return `least(100, greatest(5, 5 * round(` + col + ` / 5.0)))::smallint`
}

// socialSQL copies the social rows of one batch of legacy users (SDD 02 §5, SDD 14 §5). Owners map
// through users.legacy_id, so rows of users not imported (purged, P2.F.1.2) are skipped.
var socialSQL = []userSQL{
	{"ratings", `INSERT INTO ratings (user_id, item_id, c1, c2, c3, c4, c5, c6, created_at, updated_at)
SELECT n.id, r.item_id, ` + roundSQL("r.c1") + `, ` + roundSQL("r.c2") + `, ` + roundSQL("r.c3") + `, ` +
		roundSQL("r.c4") + `, ` + roundSQL("r.c5") + `, ` + roundSQL("r.c6") + `, r.created_at, r.updated_at
FROM legacy_import.ratings r JOIN users n ON n.legacy_id = r.user_id
WHERE r.user_id > $1 AND r.user_id <= $2
ON CONFLICT (user_id, item_id) DO NOTHING`, false},
	{"rating_migration_audit", `INSERT INTO rating_migration_audit (user_id, item_id, old_c, new_c, migrated_at)
SELECT n.id, r.item_id, ARRAY[r.c1, r.c2, r.c3, r.c4, r.c5, r.c6],
  ARRAY[` + roundSQL("r.c1") + `, ` + roundSQL("r.c2") + `, ` + roundSQL("r.c3") + `, ` + roundSQL("r.c4") + `, ` +
		roundSQL("r.c5") + `, ` + roundSQL("r.c6") + `], $3
FROM legacy_import.ratings r JOIN users n ON n.legacy_id = r.user_id
WHERE r.user_id > $1 AND r.user_id <= $2
  AND ARRAY[r.c1, r.c2, r.c3, r.c4, r.c5, r.c6] <> ARRAY[` + roundSQL("r.c1") + `, ` + roundSQL("r.c2") + `, ` +
		roundSQL("r.c3") + `, ` + roundSQL("r.c4") + `, ` + roundSQL("r.c5") + `, ` + roundSQL("r.c6") + `]
ON CONFLICT (user_id, item_id) DO NOTHING`, true},
	{
		"lists",
		`INSERT INTO lists (id, user_id, type, title, description, is_public, hidden, item_count, created_at, updated_at)
SELECT l.id, n.id, l.type::text, l.title, l.description, l.is_public, l.hidden,
  (SELECT count(*) FROM legacy_import.list_items li WHERE li.list_id = l.id), l.created_at, l.updated_at
FROM legacy_import.lists l JOIN users n ON n.legacy_id = l.user_id
WHERE l.user_id > $1 AND l.user_id <= $2
ON CONFLICT (id) DO NOTHING`,
		false,
	},
	// Legacy rows have no add date (SDD 08 §5); positions become dense 1..n in their legacy order. Items of
	// another type than the list (lists older than legacy migration 0005_list_type) are kept.
	{"list_items", `INSERT INTO list_items (list_id, item_id, position, note, added_at)
SELECT li.list_id, li.item_id, row_number() OVER (PARTITION BY li.list_id ORDER BY li.position, li.item_id),
  li.note, l.created_at
FROM legacy_import.list_items li JOIN legacy_import.lists l ON l.id = li.list_id
WHERE l.user_id > $1 AND l.user_id <= $2 AND EXISTS (SELECT 1 FROM users n WHERE n.legacy_id = l.user_id)
ON CONFLICT (list_id, item_id) DO NOTHING`, false},
	// A user report now targets the user's numeric id (app/library reports); reports whose target was not
	// imported are dropped, since an admin could not act on them.
	{"reports", `INSERT INTO reports (id, reporter_id, target_type, target_id, reason, details, status, created_at)
SELECT r.id, n.id, r.target_type::text, coalesce(t.id::text, r.target_id), r.reason, r.details, r.status::text, r.created_at
FROM legacy_import.reports r JOIN users n ON n.legacy_id = r.reporter_id
LEFT JOIN users t ON r.target_type::text = 'user' AND t.legacy_id = r.target_id
WHERE r.reporter_id > $1 AND r.reporter_id <= $2
  AND CASE r.target_type::text WHEN 'user' THEN t.id IS NOT NULL
    ELSE EXISTS (SELECT 1 FROM legacy_import.lists l JOIN users o ON o.legacy_id = l.user_id WHERE l.id::text = r.target_id) END
ON CONFLICT (id) DO NOTHING`, false},
	{"contact_messages", `INSERT INTO contact_messages (id, user_id, name, email, subject, message, handled, created_at)
SELECT c.id, n.id, c.name, c.email, c.subject, c.message, c.handled, c.created_at
FROM legacy_import.contact_messages c JOIN users n ON n.legacy_id = c.user_id
WHERE c.user_id > $1 AND c.user_id <= $2
ON CONFLICT (id) DO NOTHING`, false},
}

// socialFirst copies the messages sent without an account (or by a deleted one).
var socialFirst = []userSQL{
	{
		"contact_messages anonymous",
		`INSERT INTO contact_messages (id, name, email, subject, message, handled, created_at)
SELECT c.id, c.name, c.email, c.subject, c.message, c.handled, c.created_at FROM legacy_import.contact_messages c
WHERE c.user_id IS NULL OR NOT EXISTS (SELECT 1 FROM legacy_import."user" u WHERE u.id = c.user_id)
ON CONFLICT (id) DO NOTHING`,
		false,
	},
}

// socialLast writes the one legacy_rounding summary (SDD 02 §5): ratings read, ratings changed and, per
// category, how many ratings moved by each delta.
var socialLast = []userSQL{
	{"legacy_rounding summary", `INSERT INTO score_actions (at, kind, after, note)
SELECT $1, 'legacy_rounding', jsonb_build_object(
  'rows_read', (SELECT count(*) FROM legacy_import.ratings r JOIN users n ON n.legacy_id = r.user_id),
  'rows_changed', (SELECT count(*) FROM rating_migration_audit),
  'delta_histogram', (SELECT coalesce(jsonb_object_agg(k, h), '{}') FROM (
    SELECT 'c' || k AS k, jsonb_object_agg(d, cnt ORDER BY d) AS h FROM (
      SELECT k, a.new_c[k] - a.old_c[k] AS d, count(*) AS cnt
      FROM rating_migration_audit a, generate_series(1, 6) k GROUP BY k, d) x GROUP BY k) y)),
  'legacy import: ratings rounded to multiples of 5 (FR-02-13)'
WHERE NOT EXISTS (SELECT 1 FROM score_actions WHERE kind = 'legacy_rounding')`, true},
}

func socialBatch(now func() time.Time) BatchFunc {
	return userBatches("social", socialSQL, socialFirst, socialLast, now)
}
