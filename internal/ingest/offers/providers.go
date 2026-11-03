package offers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"
)

// The TMDB watch-provider terms (RFC-003 sections 3 and 4, ADR 0006).
const (
	// ProviderTTL is how long a provider offer lives: no TMDB value may be kept past 180 days.
	ProviderTTL = 180 * 24 * time.Hour
	// ProviderRefresh is how often an item is asked about again.
	ProviderRefresh = 30 * 24 * time.Hour
	// ProviderAttribution is shown wherever these offers are: the data comes from JustWatch through TMDB.
	ProviderAttribution = "JustWatch"
	// SourceTMDBProviders is offers.source of these rows.
	SourceTMDBProviders = "tmdb_providers"
)

// Countries are the countries the stage keeps: the product's three languages first (RFC-003 section 1).
var Countries = []string{"BR", "US", "PT", "ES", "MX", "AR"}

// Provider is one place a title is available in one country, as TMDB reports it.
type Provider struct {
	Country string
	// Kind is the offers.kind: stream, rent, buy or free.
	Kind string
	// ID is the TMDB provider id; the store code is tmdb_<id>.
	ID   int
	Name string
	// Link is the TMDB watch page of the title in that country, which carries the JustWatch attribution.
	Link string
}

func storeCode(id int) string { return "tmdb_" + strconv.Itoa(id) }

// SaveProviders replaces what TMDB says about one item: it upserts a store per provider and an offer per country and
// kind (expiring ProviderTTL from now), deactivates the earlier provider offers that TMDB no longer lists, and marks the
// item checked. Offers from other sources are never touched. An empty list is a valid answer.
func SaveProviders(ctx context.Context, tx *database.Tx, itemID string, now time.Time, ps []Provider) error {
	keys := make([]string, 0, len(ps))

	for _, p := range ps {
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO stores (code, name, kind, attribution) VALUES ($1, $2, 'streaming', $3)
ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name WHERE stores.kind = 'streaming' AND stores.name <> EXCLUDED.name`,
			storeCode(p.ID),
			p.Name,
			ProviderAttribution,
		); err != nil {
			return fmt.Errorf("offers: save store %d: %w", p.ID, err)
		}

		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO offers (item_id, country, store, kind, url, external_id, source, licence,
  confidence, last_verified_at, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, 'tmdb-justwatch', 80, $8, $9)
ON CONFLICT ON CONSTRAINT `+offerKey+` DO UPDATE SET url = EXCLUDED.url, external_id = EXCLUDED.external_id,
  last_verified_at = EXCLUDED.last_verified_at, expires_at = EXCLUDED.expires_at, active = true
WHERE offers.source = $7`,
			itemID,
			p.Country,
			storeCode(p.ID),
			p.Kind,
			p.Link,
			strconv.Itoa(p.ID),
			SourceTMDBProviders,
			now,
			now.Add(ProviderTTL),
		); err != nil {
			return fmt.Errorf("offers: save offer %s %s %d: %w", itemID, p.Country, p.ID, err)
		}

		keys = append(keys, p.Country+"|"+storeCode(p.ID)+"|"+p.Kind)
	}

	if _, err := tx.ExecContext(ctx, `UPDATE offers SET active = false
WHERE item_id = $1 AND source = $2 AND active AND NOT ((country || '|' || store || '|' || kind) = ANY($3::text[]))`,
		itemID, SourceTMDBProviders, textArray(keys)); err != nil {
		return fmt.Errorf("offers: deactivate gone providers of %s: %w", itemID, err)
	}

	if _, err := tx.ExecContext(ctx, `INSERT INTO offer_source_checks (item_id, source, checked_at) VALUES ($1, $2, $3)
ON CONFLICT (item_id, source) DO UPDATE SET checked_at = EXCLUDED.checked_at`,
		itemID, SourceTMDBProviders, now); err != nil {
		return fmt.Errorf("offers: mark %s checked: %w", itemID, err)
	}

	return nil
}

// textArray renders values as a Postgres text[] literal, quoting each so a separator inside one cannot split it.
func textArray(vals []string) string {
	quoted := make([]string, len(vals))
	for i, v := range vals {
		quoted[i] = strconv.Quote(v)
	}

	return "{" + strings.Join(quoted, ",") + "}"
}
