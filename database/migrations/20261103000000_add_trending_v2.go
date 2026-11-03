package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261103000000AddTrendingV2 adds the home window to the trending tables (RFC-002 sections 4 and 5).
//
// trending_snapshots keeps one row set per (computed_at, item_type, window_days): 30 for the API list (score v1, 10
// ranks), 7 for the home carousel (score v2, 12 ranks, basis 'pinned' for admin pins). z is the per-type z-score,
// entered_at when the item joined the list, met_at when it last met the entry rule and weak_runs the consecutive runs
// with z below 0; the next run reads them from the previous snapshot to apply the cooldown. Written by
// trending.Snapshot; read by trending.Latest and the home controller.
//
// trending_overrides holds admin pins and bans (kind pinned|banned, until NULL = no expiry). admin_id is a bare id,
// like field_reviews.resolved_by: the account purge keeps the users row (SDD 03). Written by the admin item page.
//
// item_popularity_external holds one external popularity value per (item, day, source): TMDB popularity, IGDB follows,
// Open Library readers, Wikidata sitelinks. Written by the refresh job (RFC-001 CW-04); the home score reads the
// 7-day relative change and never exposes the value.
func m20261103000000AddTrendingV2() schema.Migration {
	return postgresMigration("20261103000000_add_trending_v2",
		[]string{
			`ALTER TABLE trending_snapshots
  ADD COLUMN window_days smallint NOT NULL DEFAULT 30 CHECK (window_days IN (7, 30)),
  ADD COLUMN z real,
  ADD COLUMN entered_at timestamptz,
  ADD COLUMN met_at timestamptz,
  ADD COLUMN weak_runs smallint NOT NULL DEFAULT 0`,
			`ALTER TABLE trending_snapshots
  DROP CONSTRAINT trending_snapshots_pkey,
  DROP CONSTRAINT trending_snapshots_computed_at_item_type_item_id_key,
  DROP CONSTRAINT trending_snapshots_rank_check,
  DROP CONSTRAINT trending_snapshots_basis_check`,
			`ALTER TABLE trending_snapshots
  ADD PRIMARY KEY (computed_at, item_type, window_days, rank),
  ADD CONSTRAINT trending_snapshots_item_uniq UNIQUE (computed_at, item_type, window_days, item_id),
  ADD CONSTRAINT trending_snapshots_rank_check CHECK (rank BETWEEN 1 AND 12),
  ADD CONSTRAINT trending_snapshots_basis_check CHECK (basis IN ('trend', 'prior', 'pinned'))`,
			`CREATE TABLE trending_overrides (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  item_id uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  kind text NOT NULL CHECK (kind IN ('pinned', 'banned')),
  until timestamptz,
  reason text NOT NULL DEFAULT '',
  admin_id bigint,
  created_at timestamptz NOT NULL DEFAULT now()
)`,
			`CREATE INDEX trending_overrides_item_idx ON trending_overrides (item_id, kind)`,
			`CREATE TABLE item_popularity_external (
  item_id uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  day date NOT NULL,
  source text NOT NULL CHECK (source IN ('tmdb', 'igdb', 'openlibrary', 'wikidata')),
  value double precision NOT NULL CHECK (value >= 0),
  PRIMARY KEY (item_id, day, source)
)`,
			`CREATE INDEX item_popularity_external_day_idx ON item_popularity_external (day)`,
		},
		[]string{
			`DROP TABLE IF EXISTS item_popularity_external`,
			`DROP TABLE IF EXISTS trending_overrides`,
			`DELETE FROM trending_snapshots WHERE window_days <> 30`,
			`ALTER TABLE trending_snapshots
  DROP CONSTRAINT IF EXISTS trending_snapshots_pkey,
  DROP CONSTRAINT IF EXISTS trending_snapshots_item_uniq,
  DROP CONSTRAINT IF EXISTS trending_snapshots_rank_check,
  DROP CONSTRAINT IF EXISTS trending_snapshots_basis_check`,
			`ALTER TABLE trending_snapshots
  DROP COLUMN IF EXISTS window_days, DROP COLUMN IF EXISTS z, DROP COLUMN IF EXISTS entered_at,
  DROP COLUMN IF EXISTS met_at, DROP COLUMN IF EXISTS weak_runs`,
			`ALTER TABLE trending_snapshots
  ADD PRIMARY KEY (computed_at, item_type, rank),
  ADD UNIQUE (computed_at, item_type, item_id),
  ADD CONSTRAINT trending_snapshots_rank_check CHECK (rank BETWEEN 1 AND 10),
  ADD CONSTRAINT trending_snapshots_basis_check CHECK (basis IN ('trend', 'prior'))`,
		},
	)
}
