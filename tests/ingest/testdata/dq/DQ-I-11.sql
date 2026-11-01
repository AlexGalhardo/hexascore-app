INSERT INTO entity_match_reviews (item_id, role, source, raw_name, candidates, created_at)
SELECT '0192f1c6-0000-7000-8000-0000000000d1', 'director', 'tmdb', 'n' || g, '{}', now() - interval '8 days' FROM generate_series(1, 501) g
