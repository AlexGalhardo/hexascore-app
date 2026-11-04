-- Removes every user except two, and the interactions of the removed users, in one transaction.
-- It is a dry run unless -v commit=1 is passed: the script prints the counts before and after, then ROLLBACKs.
--
--   psql "$DATABASE_URL" -v owner_email=you@example.com -v admin_email=admin@example.com \
--        -v admin_hash='<argon2id hash>' -f scripts/ops/reset-users.sql            # dry run
--   psql ... -v commit=1 -f scripts/ops/reset-users.sql                              # applies it
--
-- The hash comes from `PASSWORD=... go run ./scripts/ops/hashpw`; the password never appears in a file or in argv.
-- Read docs/runbooks/reset-users.md first: take a backup, and check Stripe for live customers.
\set ON_ERROR_STOP on

\if :{?owner_email}
\else
	\echo 'missing -v owner_email=...'
	\quit
\endif
\if :{?admin_email}
\else
	\echo 'missing -v admin_email=...'
	\quit
\endif
\if :{?admin_hash}
\else
	\echo 'missing -v admin_hash=...'
	\quit
\endif

BEGIN;

SELECT set_config('reset.owner', lower(:'owner_email'), true) AS owner,
       set_config('reset.admin', lower(:'admin_email'), true) AS admin,
       set_config('reset.hash', :'admin_hash', true) IS NOT NULL AS hash_set \gset

\echo '== before'
SELECT (SELECT count(*) FROM users) AS users, (SELECT count(*) FROM ratings) AS ratings,
       (SELECT count(*) FROM reviews) AS reviews, (SELECT count(*) FROM comments) AS comments,
       (SELECT count(*) FROM lists) AS lists, (SELECT count(*) FROM follows) AS follows,
       (SELECT count(*) FROM sessions) AS sessions, (SELECT count(*) FROM reports) AS reports,
       (SELECT count(*) FROM contact_messages) AS contact_messages, (SELECT count(*) FROM suggestions) AS suggestions,
       (SELECT count(*) FROM subscriptions) AS subscriptions, (SELECT count(*) FROM api_keys) AS api_keys;

-- The owner must exist: a typo here would otherwise leave nobody to sign in.
DO $$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM users WHERE lower(email) = current_setting('reset.owner')) THEN
		RAISE EXCEPTION 'owner account % not found', current_setting('reset.owner');
	END IF;
END $$;

CREATE TEMP TABLE keep_users ON COMMIT DROP AS
SELECT id FROM users WHERE lower(email) IN (current_setting('reset.owner'), current_setting('reset.admin'));

-- Items whose score came from the ratings about to disappear; their aggregates are reset further down.
CREATE TEMP TABLE affected_items ON COMMIT DROP AS
SELECT DISTINCT item_id FROM ratings WHERE user_id NOT IN (SELECT id FROM keep_users);

-- Community suggestions of the removed users. Their events and reviews block the delete (no cascade), and the
-- catalogue audit keeps its row with the link to the suggestion cleared.
CREATE TEMP TABLE gone_suggestions ON COMMIT DROP AS
SELECT id FROM suggestions WHERE author_id IS NOT NULL AND author_id NOT IN (SELECT id FROM keep_users);
UPDATE catalog_audit SET suggestion_id = NULL WHERE suggestion_id IN (SELECT id FROM gone_suggestions);
DELETE FROM suggestion_events WHERE suggestion_id IN (SELECT id FROM gone_suggestions);
DELETE FROM suggestion_reviews WHERE suggestion_id IN (SELECT id FROM gone_suggestions);
DELETE FROM suggestions WHERE id IN (SELECT id FROM gone_suggestions);

-- Rows that mention a user without a cascading foreign key. A table the database does not have is skipped.
DO $$
DECLARE
	pair text[];
BEGIN
	FOREACH pair SLICE 1 IN ARRAY ARRAY[
		['rating_events', 'user_id'], ['rating_migration_audit', 'user_id'], ['access_tokens', 'user_id'],
		['remember_tokens', 'user_id'], ['sessions', 'user_id'], ['search_events', 'user_id']
	] LOOP
		IF to_regclass(pair[1]) IS NOT NULL THEN
			EXECUTE format('DELETE FROM %I WHERE %I IS NOT NULL AND %I::text NOT IN (SELECT id::text FROM keep_users)',
				pair[1], pair[2], pair[2]);
		END IF;
	END LOOP;
END $$;

DELETE FROM password_resets WHERE lower(email) NOT IN (current_setting('reset.owner'), current_setting('reset.admin'));
DELETE FROM reports WHERE reporter_id IS NULL OR reporter_id NOT IN (SELECT id FROM keep_users);
DELETE FROM contact_messages WHERE user_id IS NULL OR user_id NOT IN (SELECT id FROM keep_users);

-- The rest (ratings, reviews, comments, lists, follows, status, notifications, billing, API keys, passkeys,
-- sessions of the account, exports) cascades from the user row.
DELETE FROM users WHERE id NOT IN (SELECT id FROM keep_users);

-- Items that lost their ratings go back to unscored; the daily 04:00 recompute adds the kept accounts' ratings.
UPDATE items SET rating_count = 0, rating_weight = 0, score = NULL, score_c1 = NULL, score_c2 = NULL,
	score_c3 = NULL, score_c4 = NULL, score_c5 = NULL, score_c6 = NULL, score_se = NULL, score_computed_at = NULL
WHERE id IN (SELECT item_id FROM affected_items)
	AND NOT EXISTS (SELECT 1 FROM ratings r WHERE r.item_id = items.id);

-- The admin account: created when missing, otherwise given the new password and the admin role.
INSERT INTO users (name, email, password, email_verified_at, password_set_at, username, role, locale)
SELECT 'Admin', current_setting('reset.admin'), current_setting('reset.hash'), now(), now(),
	'hexaadmin', 'admin', 'en'
WHERE NOT EXISTS (SELECT 1 FROM users WHERE lower(email) = current_setting('reset.admin'));

UPDATE users SET password = current_setting('reset.hash'), password_set_at = now(), role = 'admin',
	email_verified_at = coalesce(email_verified_at, now()), banned_at = NULL, session_version = session_version + 1
WHERE lower(email) = current_setting('reset.admin');

\echo '== after'
SELECT (SELECT count(*) FROM users) AS users, (SELECT count(*) FROM ratings) AS ratings,
       (SELECT count(*) FROM reviews) AS reviews, (SELECT count(*) FROM comments) AS comments,
       (SELECT count(*) FROM lists) AS lists, (SELECT count(*) FROM follows) AS follows,
       (SELECT count(*) FROM sessions) AS sessions, (SELECT count(*) FROM reports) AS reports,
       (SELECT count(*) FROM contact_messages) AS contact_messages, (SELECT count(*) FROM suggestions) AS suggestions,
       (SELECT count(*) FROM subscriptions) AS subscriptions, (SELECT count(*) FROM api_keys) AS api_keys;
SELECT id, email, role, email_verified_at IS NOT NULL AS verified FROM users ORDER BY id;

\if :{?commit}
	\echo 'COMMIT'
	COMMIT;
\else
	\echo 'DRY RUN: rolled back. Pass -v commit=1 to apply.'
	ROLLBACK;
\endif
