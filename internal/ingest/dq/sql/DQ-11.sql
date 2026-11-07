SELECT i.id::text || ':' || p.key FROM items i, jsonb_each_text(i.provenance) p
WHERE NOT EXISTS (SELECT 1 FROM data_sources WHERE code = p.value)
UNION ALL
-- provenance.canonical marks a canonical taxonomy entity (genre, game mode); it names no data source.
SELECT e.id::text || ':' || p.key FROM entities e, jsonb_each_text(e.provenance) p
WHERE p.key <> 'canonical' AND NOT EXISTS (SELECT 1 FROM data_sources WHERE code = p.value)
