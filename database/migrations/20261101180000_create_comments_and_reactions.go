package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261101180000CreateCommentsAndReactions adds comments on reviews and lists (one level of replies) and one
// reaction per user and target on reviews, comments and lists (SDD 14 §12.11). Typed nullable foreign keys
// instead of a polymorphic id, so deleting a target cascades; a trigger keeps the counters in the same transaction,
// cascades and account deletion included.
func m20261101180000CreateCommentsAndReactions() schema.Migration {
	return postgresMigration("20261101180000_create_comments_and_reactions",
		[]string{
			`ALTER TABLE reviews ADD COLUMN like_count integer NOT NULL DEFAULT 0`,
			`ALTER TABLE lists ADD COLUMN like_count integer NOT NULL DEFAULT 0,
  ADD COLUMN helpful_count integer NOT NULL DEFAULT 0`,
			`CREATE TABLE comments (
  id uuid PRIMARY KEY,
  user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  review_id uuid REFERENCES reviews(id) ON DELETE CASCADE,
  list_id uuid REFERENCES lists(id) ON DELETE CASCADE,
  parent_id uuid REFERENCES comments(id) ON DELETE CASCADE,
  body text NOT NULL CHECK (char_length(body) BETWEEN 1 AND 2000),
  like_count integer NOT NULL DEFAULT 0,
  helpful_count integer NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  CHECK (num_nonnulls(review_id, list_id) = 1))`,
			`CREATE INDEX comments_review_idx ON comments (review_id, created_at) WHERE review_id IS NOT NULL`,
			`CREATE INDEX comments_list_idx ON comments (list_id, created_at) WHERE list_id IS NOT NULL`,
			`CREATE INDEX comments_user_idx ON comments (user_id)`,
			`CREATE TABLE reactions (
  user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  review_id uuid REFERENCES reviews(id) ON DELETE CASCADE,
  comment_id uuid REFERENCES comments(id) ON DELETE CASCADE,
  list_id uuid REFERENCES lists(id) ON DELETE CASCADE,
  kind text NOT NULL CHECK (kind IN ('like','helpful')),
  created_at timestamptz NOT NULL DEFAULT now(),
  CHECK (num_nonnulls(review_id, comment_id, list_id) = 1))`,
			`CREATE UNIQUE INDEX reactions_review_uq ON reactions (user_id, review_id) WHERE review_id IS NOT NULL`,
			`CREATE UNIQUE INDEX reactions_comment_uq ON reactions (user_id, comment_id) WHERE comment_id IS NOT NULL`,
			`CREATE UNIQUE INDEX reactions_list_uq ON reactions (user_id, list_id) WHERE list_id IS NOT NULL`,
			`CREATE FUNCTION reactions_bump(r reactions, d integer) RETURNS void LANGUAGE sql AS $$
UPDATE reviews SET like_count = like_count + CASE WHEN r.kind = 'like' THEN d ELSE 0 END,
  helpful_count = helpful_count + CASE WHEN r.kind = 'helpful' THEN d ELSE 0 END WHERE id = r.review_id;
UPDATE comments SET like_count = like_count + CASE WHEN r.kind = 'like' THEN d ELSE 0 END,
  helpful_count = helpful_count + CASE WHEN r.kind = 'helpful' THEN d ELSE 0 END WHERE id = r.comment_id;
UPDATE lists SET like_count = like_count + CASE WHEN r.kind = 'like' THEN d ELSE 0 END,
  helpful_count = helpful_count + CASE WHEN r.kind = 'helpful' THEN d ELSE 0 END WHERE id = r.list_id;
$$`,
			`CREATE FUNCTION reactions_count() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  IF TG_OP IN ('DELETE', 'UPDATE') THEN PERFORM reactions_bump(OLD, -1); END IF;
  IF TG_OP IN ('INSERT', 'UPDATE') THEN PERFORM reactions_bump(NEW, 1); END IF;
  RETURN NULL;
END $$`,
			`CREATE TRIGGER reactions_count AFTER INSERT OR UPDATE OR DELETE ON reactions
  FOR EACH ROW EXECUTE FUNCTION reactions_count()`,
			`ALTER TABLE user_notifications DROP CONSTRAINT user_notifications_kind_check,
  ADD CONSTRAINT user_notifications_kind_check CHECK (kind IN ('suggestion_decided','suggestion_reverted',
    'suggestion_expired','report_resolved','rating_flagged','list_hidden','system','new_follower','follow_request',
    'new_comment','new_reaction'))`,
		},
		[]string{
			`DELETE FROM user_notifications WHERE kind IN ('new_comment','new_reaction')`,
			`ALTER TABLE user_notifications DROP CONSTRAINT user_notifications_kind_check,
  ADD CONSTRAINT user_notifications_kind_check CHECK (kind IN ('suggestion_decided','suggestion_reverted',
    'suggestion_expired','report_resolved','rating_flagged','list_hidden','system','new_follower','follow_request'))`,
			`DROP FUNCTION IF EXISTS reactions_count() CASCADE`,
			`DROP FUNCTION IF EXISTS reactions_bump(reactions, integer)`,
			`DROP TABLE IF EXISTS reactions`,
			`DROP TABLE IF EXISTS comments`,
			`ALTER TABLE lists DROP COLUMN IF EXISTS like_count, DROP COLUMN IF EXISTS helpful_count`,
			`ALTER TABLE reviews DROP COLUMN IF EXISTS like_count`,
		},
	)
}
