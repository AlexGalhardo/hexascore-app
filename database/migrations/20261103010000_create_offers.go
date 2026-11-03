package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261103010000CreateOffers adds the where-to-watch and where-to-buy model of RFC-003 section 2.
//
// stores lists the places an offer can point to. url_template carries a {id} or {isbn13} placeholder when the source
// gives identifiers instead of URLs; the affiliate tag is never stored, it is appended at render time from
// affiliate_param and the AFFILIATE_* environment variables. Streaming stores are created by the TMDB stage on demand;
// the rows seeded here are the ones the no-network stages derive.
//
// offers is keyed by item × edition × country × store × kind. country is an ISO 3166 alpha-2 code, or 'ZZ' for an
// offer that is valid everywhere (admin links, digital stores). edition_id is set only for book offers. URLs are
// canonical (no affiliate tag); source 'admin' rows migrated from affiliate_links keep the URL the admin entered.
// verify_status follows the offers.verify state machine (RFC-003 section 3); expires_at is set for TMDB data
// (last_verified_at + 180 days, ADR 0006). Written by the offers stages, the verify job and the admin item page.
//
// offer_clicks counts outbound clicks per day, offer and country with no user id and no IP (LGPD); written by
// GET /go/{id}, pruned after 400 days.
func m20261103010000CreateOffers() schema.Migration {
	return postgresMigration("20261103010000_create_offers",
		[]string{
			`CREATE TABLE stores (
  code text PRIMARY KEY,
  name text NOT NULL,
  kind text NOT NULL CHECK (kind IN ('streaming', 'digital_store', 'mobile_store', 'retailer', 'library')),
  url_template text,
  affiliate_program text,
  affiliate_param text,
  logo_key text,
  attribution text,
  active boolean NOT NULL DEFAULT true
)`,
			`CREATE TABLE offers (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  item_id uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  edition_id uuid REFERENCES editions(id) ON DELETE CASCADE,
  country char(2) NOT NULL CHECK (country ~ '^[A-Z]{2}$'),
  store text NOT NULL REFERENCES stores(code),
  kind text NOT NULL CHECK (kind IN ('stream', 'rent', 'buy', 'free', 'buy_digital', 'mobile', 'physical', 'read')),
  url text NOT NULL CHECK (url ~ '^https://'),
  external_id text,
  source text NOT NULL CHECK (source IN ('tmdb_providers', 'wikidata', 'editions', 'admin', 'watchmode')),
  licence text NOT NULL,
  confidence smallint NOT NULL DEFAULT 80 CHECK (confidence BETWEEN 0 AND 100),
  first_seen timestamptz NOT NULL DEFAULT now(),
  last_verified_at timestamptz,
  verify_status text NOT NULL DEFAULT 'unverified'
    CHECK (verify_status IN ('unverified', 'ok', 'redirected', 'broken', 'gone')),
  verify_http_status int,
  broken_runs smallint NOT NULL DEFAULT 0,
  expires_at timestamptz,
  active boolean NOT NULL DEFAULT true,
  UNIQUE NULLS NOT DISTINCT (item_id, edition_id, country, store, kind)
)`,
			// serves the item page: the active offers of an item in the visitor's country (and 'ZZ')
			`CREATE INDEX offers_item_country_idx ON offers (item_id, country) WHERE active`,
			// serves offers.verify: the least recently verified active offers first
			`CREATE INDEX offers_verify_idx ON offers (last_verified_at NULLS FIRST) WHERE active`,
			// serves the expiry sweep of TMDB data
			`CREATE INDEX offers_expires_idx ON offers (expires_at) WHERE expires_at IS NOT NULL AND active`,
			`CREATE TABLE offer_clicks (
  day date NOT NULL,
  offer_id uuid NOT NULL REFERENCES offers(id) ON DELETE CASCADE,
  country char(2) NOT NULL,
  clicks int NOT NULL DEFAULT 0,
  PRIMARY KEY (day, offer_id, country)
)`,
			`INSERT INTO stores (code, name, kind, url_template, affiliate_program, affiliate_param) VALUES
  ('steam', 'Steam', 'digital_store', 'https://store.steampowered.com/app/{id}', NULL, NULL),
  ('gog', 'GOG', 'digital_store', 'https://www.gog.com/en/{id}', NULL, NULL),
  ('epic', 'Epic Games Store', 'digital_store', 'https://store.epicgames.com/en-US/p/{id}', NULL, NULL),
  ('google_play', 'Google Play', 'mobile_store', 'https://play.google.com/store/apps/details?id={id}', NULL, NULL),
  ('app_store', 'App Store', 'mobile_store', 'https://apps.apple.com/app/id{id}', NULL, NULL),
  ('amazon_br', 'Amazon Brasil', 'retailer', 'https://www.amazon.com.br/s?k={isbn13}', 'amazon_associates', 'tag'),
  ('amazon_us', 'Amazon', 'retailer', 'https://www.amazon.com/s?k={isbn13}', 'amazon_associates', 'tag'),
  ('amazon', 'Amazon', 'retailer', NULL, 'amazon_associates', 'tag'),
  ('mercadolivre', 'Mercado Livre', 'retailer', NULL, NULL, NULL),
  ('openlibrary', 'Open Library', 'library', 'https://archive.org/details/{id}', NULL, NULL)`,
		},
		[]string{
			`DROP TABLE IF EXISTS offer_clicks`,
			`DROP TABLE IF EXISTS offers`,
			`DROP TABLE IF EXISTS stores`,
		},
	)
}

// m20261103010100MigrateAffiliateLinksToOffers copies the admin affiliate links into offers (expand step of RFC-003
// section 2; affiliate_links stays, read-only, for two releases). The store follows the URL host for Amazon
// (amazon.com.br → amazon_br / BR, amazon.com → amazon_us / US, anything else → amazon / ZZ); mercadolivre is BR; steam
// and gog are worldwide. Steam and GOG links are digital purchases, the rest plain purchases. Rerunning copies nothing
// twice (ON CONFLICT DO NOTHING on the offer key).
func m20261103010100MigrateAffiliateLinksToOffers() schema.Migration {
	return postgresMigration("20261103010100_migrate_affiliate_links_to_offers",
		[]string{
			`INSERT INTO offers (item_id, country, store, kind, url, source, licence, confidence, active)
SELECT a.item_id,
  CASE WHEN a.store = 'amazon' AND a.url ~* '^https://([a-z0-9.-]+\.)?amazon\.com\.br(/|$)' THEN 'BR'
       WHEN a.store = 'amazon' AND a.url ~* '^https://([a-z0-9.-]+\.)?amazon\.com(/|$)' THEN 'US'
       WHEN a.store = 'mercadolivre' THEN 'BR'
       ELSE 'ZZ' END,
  CASE WHEN a.store = 'amazon' AND a.url ~* '^https://([a-z0-9.-]+\.)?amazon\.com\.br(/|$)' THEN 'amazon_br'
       WHEN a.store = 'amazon' AND a.url ~* '^https://([a-z0-9.-]+\.)?amazon\.com(/|$)' THEN 'amazon_us'
       ELSE a.store END,
  CASE WHEN a.store IN ('steam', 'gog') THEN 'buy_digital' ELSE 'buy' END,
  a.url, 'admin', 'proprietary', 100, a.active
FROM affiliate_links a
ON CONFLICT DO NOTHING`,
		},
		[]string{`DELETE FROM offers WHERE source = 'admin'`},
	)
}
