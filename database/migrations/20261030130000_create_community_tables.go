package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261030130000CreateCommunityTables is SDD 04 §5. users.id is bigint in this
// app, so the user foreign keys are bigint. The append-only trigger is
// statement-level: an UPDATE of a missing id still raises P0001. When the
// hexascore_app role exists it receives only SELECT and INSERT (SDD 11).
// Existing catalog_merge_audit rows are copied as source=merge; that table
// stays, because the merge service still writes it.
func m20261030130000CreateCommunityTables() schema.Migration {
	return postgresMigration("20261030130000_create_community_tables", communityUp, communityDown)
}

var communityUp = []string{
	`CREATE TYPE suggestion_kind AS ENUM ('edit','create','merge')`,
	`CREATE TYPE suggestion_target AS ENUM ('item','edition','entity')`,
	`CREATE TYPE suggestion_status AS ENUM ('draft','submitted','in_review','approved','applied','rejected','reverted','withdrawn','expired')`,
	`CREATE TYPE reject_reason AS ENUM ('incorrect','no_source','bad_source','duplicate','policy','spam','stale','target_gone','other')`,
	`CREATE TYPE actor_role AS ENUM ('author','admin','moderator','system')`,
	`CREATE TABLE suggestions (
  id uuid PRIMARY KEY,
  author_id bigint REFERENCES users(id) ON DELETE SET NULL,
  kind suggestion_kind NOT NULL,
  target_type suggestion_target NOT NULL,
  target_id uuid,
  target_domain text NOT NULL CHECK (target_domain IN ('game','movie','series','book')),
  locale text NOT NULL CHECK (locale IN ('en','pt-BR','es')),
  changes jsonb NOT NULL,
  note text CHECK (char_length(note) <= 1000),
  source_urls text[] NOT NULL DEFAULT '{}' CHECK (cardinality(source_urls) <= 3),
  status suggestion_status NOT NULL DEFAULT 'draft',
  assigned_to bigint REFERENCES users(id) ON DELETE SET NULL,
  claimed_at timestamptz,
  decided_by bigint REFERENCES users(id) ON DELETE SET NULL,
  decided_at timestamptz,
  reject_reason reject_reason,
  decision_note text CHECK (char_length(decision_note) <= 1000),
  submitted_at timestamptz,
  submitted_day date GENERATED ALWAYS AS ((submitted_at AT TIME ZONE 'UTC')::date) STORED,
  depends_on uuid REFERENCES suggestions(id),
  meta jsonb NOT NULL DEFAULT '{}',
  version integer NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CHECK ((kind = 'create') = (target_id IS NULL))
)`,
	`CREATE INDEX suggestions_queue_idx ON suggestions (submitted_at) WHERE status IN ('submitted','in_review')`,
	`CREATE INDEX suggestions_quota_idx ON suggestions (author_id, submitted_day)`,
	`CREATE INDEX suggestions_target_idx ON suggestions (target_type, target_id, status)`,
	`CREATE INDEX suggestions_changes_gin ON suggestions USING gin (changes)`,
	`CREATE TABLE suggestion_events (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  suggestion_id uuid NOT NULL REFERENCES suggestions(id),
  from_status suggestion_status,
  to_status suggestion_status NOT NULL,
  actor_id bigint REFERENCES users(id) ON DELETE SET NULL,
  actor_role actor_role NOT NULL,
  reason text,
  meta jsonb NOT NULL DEFAULT '{}',
  created_at timestamptz NOT NULL DEFAULT now()
)`,
	`CREATE INDEX suggestion_events_sid_idx ON suggestion_events (suggestion_id, id)`,
	`CREATE TYPE audit_source AS ENUM ('community','admin','revert','merge','create')`,
	`CREATE TABLE catalog_audit (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  occurred_at timestamptz NOT NULL DEFAULT now(),
  actor_id bigint REFERENCES users(id) ON DELETE SET NULL,
  actor_role actor_role NOT NULL,
  source audit_source NOT NULL,
  suggestion_id uuid REFERENCES suggestions(id),
  target_type suggestion_target NOT NULL,
  target_id uuid NOT NULL,
  field text NOT NULL,
  before jsonb,
  after jsonb,
  prev_provenance jsonb,
  reason text,
  reverts_audit_id bigint REFERENCES catalog_audit(id),
  request_id text
)`,
	`CREATE INDEX catalog_audit_target_idx ON catalog_audit (target_type, target_id, occurred_at DESC)`,
	`CREATE INDEX catalog_audit_field_idx ON catalog_audit (target_type, target_id, field, id DESC)`,
	`CREATE INDEX catalog_audit_actor_idx ON catalog_audit (actor_id, occurred_at DESC)`,
	`CREATE UNIQUE INDEX catalog_audit_one_revert ON catalog_audit (reverts_audit_id) WHERE reverts_audit_id IS NOT NULL`,
	`CREATE FUNCTION catalog_audit_append_only() RETURNS trigger AS $$
BEGIN
  RAISE EXCEPTION 'catalog_audit is append-only';
END $$ LANGUAGE plpgsql`,
	`CREATE TRIGGER catalog_audit_append_only
  BEFORE UPDATE OR DELETE ON catalog_audit
  FOR EACH STATEMENT EXECUTE FUNCTION catalog_audit_append_only()`,
	`DO $hexascore_grant$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'hexascore_app') THEN
    REVOKE ALL ON TABLE catalog_audit FROM hexascore_app;
    GRANT SELECT, INSERT ON TABLE catalog_audit TO hexascore_app;
  END IF;
END
$hexascore_grant$`,
	`CREATE TYPE contributor_tier AS ENUM ('new','standard','low','blocked')`,
	`CREATE TABLE contributor_stats (
  user_id bigint PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  submitted_total integer NOT NULL DEFAULT 0,
  approved_total integer NOT NULL DEFAULT 0,
  rejected_total integer NOT NULL DEFAULT 0,
  reverted_total integer NOT NULL DEFAULT 0,
  withdrawn_total integer NOT NULL DEFAULT 0,
  accuracy numeric(4,3) GENERATED ALWAYS AS ((approved_total + 1)::numeric / (approved_total + rejected_total + reverted_total + 2)) STORED,
  tier contributor_tier NOT NULL DEFAULT 'new',
  cooldown_until timestamptz,
  blocked_until timestamptz,
  trusted_eligible boolean NOT NULL DEFAULT false,
  updated_at timestamptz NOT NULL DEFAULT now()
)`,
	`CREATE TABLE catalog_locks (
  id uuid PRIMARY KEY,
  target_type suggestion_target NOT NULL,
  target_id uuid NOT NULL,
  field text NOT NULL DEFAULT '*',
  reason text NOT NULL,
  locked_by bigint REFERENCES users(id) ON DELETE SET NULL,
  expires_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (target_type, target_id, field)
)`,
	`INSERT INTO catalog_audit (occurred_at, actor_id, actor_role, source, target_type, target_id, field, before, after)
SELECT occurred_at, actor_id, 'system', 'merge', 'item', target_item_id, '*merge*',
  jsonb_build_object('sourceItemId', source_item_id, 'itemType', item_type), snapshot
FROM catalog_merge_audit`,
}

var communityDown = []string{
	`DROP TABLE IF EXISTS catalog_locks`,
	`DROP TABLE IF EXISTS contributor_stats`,
	`DROP TRIGGER IF EXISTS catalog_audit_append_only ON catalog_audit`,
	`DROP TABLE IF EXISTS catalog_audit`,
	`DROP FUNCTION IF EXISTS catalog_audit_append_only()`,
	`DROP TABLE IF EXISTS suggestion_events`,
	`DROP TABLE IF EXISTS suggestions`,
	`DROP TYPE IF EXISTS contributor_tier`,
	`DROP TYPE IF EXISTS audit_source`,
	`DROP TYPE IF EXISTS actor_role`,
	`DROP TYPE IF EXISTS reject_reason`,
	`DROP TYPE IF EXISTS suggestion_status`,
	`DROP TYPE IF EXISTS suggestion_target`,
	`DROP TYPE IF EXISTS suggestion_kind`,
}
