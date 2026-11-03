package trending

import (
	"context"
	"fmt"
	"time"

	"github.com/AlexGalhardo/vukem/database"
)

// ExternalValue is one external popularity reading of an item (TMDB popularity, IGDB follows, Open Library readers or
// Wikidata sitelinks). The refresh job writes one per item, day and source; the value is never shown or exported.
type ExternalValue struct {
	ItemID string
	Source string
	Day    time.Time
	Value  float64
}

// RecordExternal stores the readings in one transaction; a reading of the same (item, day, source) replaces the earlier
// one, so a retried refresh writes the same rows.
func RecordExternal(ctx context.Context, db *database.DB, vals []ExternalValue) error {
	err := db.Tx(ctx, func(tx *database.Tx) error {
		for _, v := range vals {
			if _, err := tx.ExecContext(ctx, `INSERT INTO item_popularity_external (item_id, day, source, value)
VALUES ($1, $2::date, $3, $4)
ON CONFLICT (item_id, day, source) DO UPDATE SET value = EXCLUDED.value`,
				v.ItemID, v.Day.UTC(), v.Source, v.Value); err != nil {
				return fmt.Errorf("record external %s %s: %w", v.Source, v.ItemID, err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("trending record external: %w", err)
	}

	return nil
}
