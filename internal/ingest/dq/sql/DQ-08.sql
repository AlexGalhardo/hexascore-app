WITH ok(role, kind, types) AS (VALUES
  ('director', 'person', ARRAY['movie', 'series']), ('writer', 'person', ARRAY['movie', 'series']),
  ('cast', 'person', ARRAY['movie', 'series']), ('creator', 'person', ARRAY['movie', 'series']),
  ('author', 'person', ARRAY['book']),
  ('developer', 'organization', ARRAY['game']), ('publisher', 'organization', ARRAY['game', 'book']),
  ('production_company', 'organization', ARRAY['movie', 'series']), ('network', 'organization', ARRAY['series']),
  ('platform', 'platform', ARRAY['game']), ('game_mode', 'game_mode', ARRAY['game']),
  ('franchise', 'franchise', ARRAY['game', 'movie', 'series', 'book']),
  ('series', 'franchise', ARRAY['game', 'movie', 'series', 'book']),
  ('genre', 'genre', ARRAY['game', 'movie', 'series', 'book']),
  ('country', 'country', ARRAY['game', 'movie', 'series', 'book']),
  ('original_language', 'language', ARRAY['game', 'movie', 'series', 'book']))
SELECT ie.item_id::text || ':' || ie.role || ':' || ie.entity_id::text
FROM item_entities ie
JOIN items i ON i.id = ie.item_id
JOIN entities e ON e.id = ie.entity_id
WHERE NOT EXISTS (SELECT 1 FROM ok WHERE ok.role = ie.role AND ok.kind = e.kind AND i.type = ANY (ok.types))
