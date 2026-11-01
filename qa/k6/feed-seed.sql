-- Seeds a LOCAL load-test database (never production) for qa/k6/feed.js (SDD 14 NFR-14-7): every DemoCatalog user
-- opts in to the feed and follows every other one, and 10,000 activity events are spread over them on public items.
-- Needs FLAG_SOCIAL=true on the app. Idempotent: it clears activity_events and the demo follows first.
--   docker exec -i hexascore-postgres-1 psql -U hexascore -d <db> -v ON_ERROR_STOP=1 < qa/k6/feed-seed.sql
BEGIN;

UPDATE users SET feed_opt_in = true WHERE username LIKE 'demo%';

DELETE FROM follows WHERE follower_id IN (SELECT id FROM users WHERE username LIKE 'demo%');
INSERT INTO follows (follower_id, followee_id, state, accepted_at)
SELECT a.id, b.id, 'accepted', now() FROM users a JOIN users b ON a.id <> b.id
WHERE a.username LIKE 'demo%' AND b.username LIKE 'demo%';

DELETE FROM activity_events;
WITH actors AS (SELECT id, row_number() OVER (ORDER BY id) - 1 AS n, count(*) OVER () AS total
  FROM users WHERE username LIKE 'demo%'),
items AS (SELECT id, row_number() OVER (ORDER BY id) - 1 AS n, count(*) OVER () AS total
  FROM items WHERE visibility = 'public'),
g AS (SELECT s FROM generate_series(0, 9999) s)
INSERT INTO activity_events (id, actor_id, kind, item_id, payload, created_at)
SELECT gen_random_uuid(), a.id,
  CASE WHEN g.s % 2 = 0 THEN 'rating_saved' ELSE 'status_changed' END,
  i.id,
  CASE WHEN g.s % 2 = 0 THEN '{}'::jsonb ELSE '{"status":"completed"}'::jsonb END,
  now() - make_interval(mins => g.s)
FROM g
JOIN actors a ON a.n = g.s % a.total
JOIN items i ON i.n = (g.s * 7) % i.total;

COMMIT;

SELECT count(*) AS activity_events FROM activity_events;
