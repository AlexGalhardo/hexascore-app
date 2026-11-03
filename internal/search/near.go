package search

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/AlexGalhardo/vukem/database"
)

// NearTitle finds the closest other item of the type whose English title is almost the same as title and whose release
// year is within maxYearGap years of year (RFC-001 section 5 C, entity resolution). Titles are compared normalised
// (hs_norm) by trigram similarity, at least minSimilarity; the trigram index prefilters. It returns the item id and the
// similarity, and ok false when nothing qualifies. The item with the QID excludeQID is ignored.
func NearTitle(
	ctx context.Context,
	ex database.Executor,
	itemType, title, excludeQID string,
	year, maxYearGap int,
	minSimilarity float64,
) (id string, similarity float64, ok bool, err error) {
	err = ex.QueryRowContext(ctx, `SELECT i.id::text, similarity(t.title_norm, hs_norm($2))::float8
FROM item_translations t JOIN items i ON i.id = t.item_id
WHERE t.locale = 'en' AND i.type = $1 AND i.wikidata_qid <> $3 AND t.title_norm % hs_norm($2)
  AND i.release_date IS NOT NULL AND abs(extract(year FROM i.release_date)::int - $4) <= $5
  AND similarity(t.title_norm, hs_norm($2)) >= $6
ORDER BY 2 DESC, i.created_at, i.id LIMIT 1`, itemType, title, excludeQID, year, maxYearGap, minSimilarity).
		Scan(&id, &similarity)
	if errors.Is(err, sql.ErrNoRows) {
		return "", 0, false, nil
	}

	if err != nil {
		return "", 0, false, fmt.Errorf("search: near title of %s: %w", excludeQID, err)
	}

	return id, similarity, true, nil
}
