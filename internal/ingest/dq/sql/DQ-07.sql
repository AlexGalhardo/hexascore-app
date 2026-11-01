SELECT i.id::text FROM items i
LEFT JOIN data_sources s ON s.code = i.cover_source
WHERE i.cover_key IS NOT NULL
  AND (i.cover_source IS NULL OR (s.attribution IS NOT NULL AND i.cover_attribution IS NULL))
