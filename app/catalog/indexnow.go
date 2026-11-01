package catalog

import (
	"context"
	"fmt"
	"time"

	"github.com/AlexGalhardo/vukem/database"
)

// IndexNowBatch caps the api_changes rows one sweep reads; the rest wait for the next run.
const IndexNowBatch = 5000

// IndexNowCursor is how far the IndexNow sweep has read (table indexnow_cursor).
type IndexNowCursor struct {
	LastChangeID  int64
	EntitiesSince time.Time
}

// itemSection is the web section of an item type ("" when unknown).
func itemSection(typ string) string {
	for _, s := range sitemapSources {
		if s.family == "items" && s.key == typ {
			return s.path
		}
	}

	return ""
}

// ChangedURLs returns the public page URLs (en, pt-br, es) of the items in api_changes after c.LastChangeID and of
// the indexable entities updated after c.EntitiesSince and at or before now, plus the cursor to store. Score-only
// updates are left out: ratings move scores all day and the catalogue text did not change. A deleted or hidden item
// is included while its row still has a slug, so engines recrawl the 404.
func ChangedURLs(ctx context.Context, ex database.Executor, base string, c IndexNowCursor, now time.Time) (
	[]string, IndexNowCursor, error,
) {
	next := IndexNowCursor{LastChangeID: c.LastChangeID, EntitiesSince: now}

	rows, err := ex.QueryContext(ctx, `SELECT c.id, coalesce(i.type, ''), coalesce(i.slug, '')
FROM (SELECT id, item_id, action, fields FROM api_changes WHERE id > $1 ORDER BY id LIMIT $2) c
LEFT JOIN items i ON i.id = c.item_id
WHERE NOT (c.action = 'updated' AND c.fields <@ ARRAY['score', 'confidence', 'ratingCount']::text[])
ORDER BY c.id`, c.LastChangeID, IndexNowBatch)
	if err != nil {
		return nil, c, fmt.Errorf("indexnow: read api_changes: %w", err)
	}

	var paths []string

	for rows.Next() {
		var (
			id        int64
			typ, slug string
		)
		if err := rows.Scan(&id, &typ, &slug); err != nil {
			_ = rows.Close()

			return nil, c, err
		}

		if section := itemSection(typ); section != "" && slug != "" {
			paths = append(paths, "/"+section+"/"+slug)
		}
	}

	if err := rows.Close(); err != nil {
		return nil, c, err
	}

	if err := ex.QueryRowContext(ctx, `SELECT coalesce(max(id), $1) FROM (
  SELECT id FROM api_changes WHERE id > $1 ORDER BY id LIMIT $2) c`, c.LastChangeID, IndexNowBatch).
		Scan(&next.LastChangeID); err != nil {
		return nil, c, fmt.Errorf("indexnow: api_changes cursor: %w", err)
	}

	for _, s := range sitemapSources {
		if s.family != "entities" {
			continue
		}

		keys, err := changedEntityKeys(ctx, ex, s, c.EntitiesSince, now)
		if err != nil {
			return nil, c, err
		}

		for _, k := range keys {
			paths = append(paths, "/"+s.path+"/"+k)
		}
	}

	urls := make([]string, 0, 3*len(paths))
	for _, p := range paths {
		urls = append(urls, base+p, base+"/pt-br"+p, base+"/es"+p)
	}

	return urls, next, nil
}

// changedEntityKeys lists the URL keys of the indexable entities of one kind updated in (since, now]; s.column is a
// constant of sitemapSources, never input.
func changedEntityKeys(ctx context.Context, ex database.Executor, s sitemapSource, since, now time.Time) (
	[]string, error,
) {
	rows, err := ex.QueryContext(ctx, fmt.Sprintf(`SELECT e.%[1]s FROM entities e
JOIN item_entities ie ON ie.entity_id = e.id
JOIN items i ON i.id = ie.item_id AND i.visibility = 'public'
WHERE e.kind = $1 AND e.%[1]s IS NOT NULL AND e.updated_at > $2 AND e.updated_at <= $3
GROUP BY e.id HAVING count(DISTINCT ie.item_id) >= %[2]d`, s.column, MinIndexableItems), s.key, since, now)
	if err != nil {
		return nil, fmt.Errorf("indexnow: changed %s entities: %w", s.key, err)
	}
	defer rows.Close()

	var keys []string

	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			return nil, err
		}

		keys = append(keys, k)
	}

	return keys, rows.Err()
}
