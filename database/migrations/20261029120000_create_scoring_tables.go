package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029120000CreateScoringTables is SDD 02 §5: ratings with the generated
// unrounded overall, status, time-to-beat, anti-abuse (bursts, flags,
// actions), rating events for SDD 08, formula versions, priors and the A-04
// rounding audit, plus the score columns on items. user_id is bigint (SDD 03
// §5 amendment); item types are text with a CHECK because items.type is text.
func m20261029120000CreateScoringTables() schema.Migration {
	return schema.Migration{
		ID: "20261029120000_create_scoring_tables",
		Up: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil // generated columns, enums, arrays: PostgreSQL only (VF-009, VF-033)
			}
			return s.Raw(scoringUp)
		},
		Down: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil
			}
			return s.Raw(scoringDown)
		},
	}
}

const scoringUp = `CREATE TABLE ratings (
  user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  item_id uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  c1 smallint NOT NULL, c2 smallint NOT NULL, c3 smallint NOT NULL,
  c4 smallint NOT NULL, c5 smallint NOT NULL, c6 smallint NOT NULL,
  overall numeric GENERATED ALWAYS AS ((c1 + c2 + c3 + c4 + c5 + c6)::numeric / 6) STORED,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, item_id),
  CONSTRAINT ratings_half_star CHECK (
    c1 BETWEEN 5 AND 100 AND c1 % 5 = 0 AND c2 BETWEEN 5 AND 100 AND c2 % 5 = 0 AND
    c3 BETWEEN 5 AND 100 AND c3 % 5 = 0 AND c4 BETWEEN 5 AND 100 AND c4 % 5 = 0 AND
    c5 BETWEEN 5 AND 100 AND c5 % 5 = 0 AND c6 BETWEEN 5 AND 100 AND c6 % 5 = 0)
);
CREATE INDEX ratings_item_idx ON ratings (item_id, created_at);
CREATE INDEX ratings_user_idx ON ratings (user_id, created_at);

CREATE TYPE item_status_value AS ENUM ('want','in_progress','completed','dropped');
CREATE TABLE item_status (
  user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  item_id uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  status item_status_value NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, item_id)
);
CREATE INDEX item_status_user_idx ON item_status (user_id, status);

CREATE TABLE time_to_beat (
  user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  item_id uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  main_minutes int, extras_minutes int, completionist_minutes int,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, item_id),
  CHECK ((main_minutes IS NULL OR main_minutes BETWEEN 30 AND 600000) AND
         (extras_minutes IS NULL OR extras_minutes BETWEEN 30 AND 600000) AND
         (completionist_minutes IS NULL OR completionist_minutes BETWEEN 30 AND 600000)),
  CHECK (coalesce(main_minutes, extras_minutes, completionist_minutes) IS NOT NULL)
);
CREATE INDEX time_to_beat_item_idx ON time_to_beat (item_id);

CREATE TABLE score_bursts (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  item_id uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  started_at timestamptz NOT NULL,
  limited_until timestamptz NOT NULL,
  trigger text NOT NULL CHECK (trigger IN ('volume','skew')),
  stats jsonb NOT NULL,
  resolution text CHECK (resolution IN ('approved','excluded','auto_released','kept_held')),
  resolved_at timestamptz,
  resolved_by bigint REFERENCES users(id) ON DELETE SET NULL
);
CREATE INDEX score_bursts_open_idx ON score_bursts (item_id) WHERE resolution IS NULL;

-- Keyed by (user, item), not a FK to ratings: deleting and re-creating a
-- rating must not launder its flag (SDD 02 §5).
CREATE TABLE rating_flags (
  user_id bigint NOT NULL,
  item_id uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  shadow_weight numeric(3,2) NOT NULL CHECK (shadow_weight BETWEEN 0 AND 1),
  reason text NOT NULL CHECK (reason IN ('burst_hold','admin_exclude','admin_downweight')),
  burst_id bigint REFERENCES score_bursts(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, item_id)
);

CREATE TABLE score_actions (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  at timestamptz NOT NULL DEFAULT now(),
  actor bigint,
  kind text NOT NULL,
  item_id uuid, user_id bigint, burst_id bigint,
  before jsonb, after jsonb, note text
);

CREATE TYPE rating_event_kind AS ENUM ('rating_created','rating_updated','rating_deleted',
  'status_changed','time_to_beat_reported','time_to_beat_cleared');
CREATE TABLE rating_events (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  at timestamptz NOT NULL DEFAULT now(),
  kind rating_event_kind NOT NULL,
  user_id bigint NOT NULL,
  item_id uuid NOT NULL,
  item_type text NOT NULL CHECK (item_type IN ('game','movie','series','book')),
  overall_before numeric, overall_after numeric,
  status_before item_status_value, status_after item_status_value,
  weight numeric(4,3)
);
CREATE INDEX rating_events_item_idx ON rating_events (item_id, at);
CREATE INDEX rating_events_user_idx ON rating_events (user_id, at);

CREATE TABLE score_formula_versions (
  version smallint PRIMARY KEY,
  m_game smallint NOT NULL, m_movie smallint NOT NULL, m_series smallint NOT NULL, m_book smallint NOT NULL,
  params jsonb NOT NULL,
  activated_at timestamptz NOT NULL,
  changelog text NOT NULL
);

CREATE TABLE score_priors (
  item_type text NOT NULL CHECK (item_type IN ('game','movie','series','book')),
  formula_version smallint NOT NULL REFERENCES score_formula_versions(version),
  prior_c numeric[] NOT NULL CHECK (array_length(prior_c, 1) = 6),
  items_used int NOT NULL,
  computed_at timestamptz NOT NULL,
  PRIMARY KEY (item_type, formula_version)
);

CREATE TABLE rating_migration_audit (
  user_id bigint NOT NULL,
  item_id uuid NOT NULL,
  old_c smallint[] NOT NULL CHECK (array_length(old_c, 1) = 6),
  new_c smallint[] NOT NULL CHECK (array_length(new_c, 1) = 6),
  migrated_at timestamptz NOT NULL,
  PRIMARY KEY (user_id, item_id)
);

ALTER TABLE items
  ADD rating_count int NOT NULL DEFAULT 0,
  ADD rating_weight numeric(12,3) NOT NULL DEFAULT 0,
  ADD score numeric(7,4),
  ADD score_c1 numeric(7,4), ADD score_c2 numeric(7,4), ADD score_c3 numeric(7,4),
  ADD score_c4 numeric(7,4), ADD score_c5 numeric(7,4), ADD score_c6 numeric(7,4),
  ADD score_se numeric(7,4),
  ADD score_confidence text CHECK (score_confidence IN ('low','medium','high')),
  ADD score_formula_version smallint,
  ADD score_computed_at timestamptz,
  ADD ratings_limited_until timestamptz,
  ADD unusual_activity_until timestamptz`

const scoringDown = `ALTER TABLE items
  DROP COLUMN IF EXISTS rating_count, DROP COLUMN IF EXISTS rating_weight, DROP COLUMN IF EXISTS score,
  DROP COLUMN IF EXISTS score_c1, DROP COLUMN IF EXISTS score_c2, DROP COLUMN IF EXISTS score_c3,
  DROP COLUMN IF EXISTS score_c4, DROP COLUMN IF EXISTS score_c5, DROP COLUMN IF EXISTS score_c6,
  DROP COLUMN IF EXISTS score_se, DROP COLUMN IF EXISTS score_confidence,
  DROP COLUMN IF EXISTS score_formula_version, DROP COLUMN IF EXISTS score_computed_at,
  DROP COLUMN IF EXISTS ratings_limited_until, DROP COLUMN IF EXISTS unusual_activity_until;
DROP TABLE IF EXISTS rating_migration_audit, score_priors, score_formula_versions, rating_events,
  score_actions, rating_flags, score_bursts, time_to_beat, item_status, ratings;
DROP TYPE IF EXISTS rating_event_kind;
DROP TYPE IF EXISTS item_status_value`
