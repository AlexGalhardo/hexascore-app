package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029140008CreateSearchEventsTable is SDD 07 §5: one row per performed search or result click, kept 30 days (SDD 07 §7.8; the daily purge is `search:purge`). query_norm never holds e-mail, phone or CPF patterns (the writer redacts them).
func m20261029140008CreateSearchEventsTable() schema.Migration {
	return postgresMigration("20261029140008_create_search_events_table",
		[]string{
			`CREATE TABLE search_events (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  at timestamptz NOT NULL DEFAULT now(),
  search_id uuid NOT NULL,
  surface text NOT NULL CHECK (surface IN ('web','app','suggest')),
  query_norm text NOT NULL,
  user_id bigint,
  locale text NOT NULL DEFAULT '',
  types text[] NOT NULL DEFAULT '{}',
  result_count integer NOT NULL DEFAULT 0,
  latency_ms integer NOT NULL DEFAULT 0,
  page integer NOT NULL DEFAULT 1,
  clicked_ref uuid,
  click_position integer,
  click_group text
)`,
			`CREATE INDEX search_events_at_idx ON search_events (at)`,
			`CREATE INDEX search_events_search_id_idx ON search_events (search_id)`,
		},
		[]string{`DROP TABLE IF EXISTS search_events`},
	)
}
