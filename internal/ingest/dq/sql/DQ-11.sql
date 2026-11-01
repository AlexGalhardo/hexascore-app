SELECT i.id::text || ':' || p.key FROM items i, jsonb_each_text(i.provenance) p
WHERE NOT EXISTS (SELECT 1 FROM data_sources WHERE code = p.value)
UNION ALL
SELECT e.id::text || ':' || p.key FROM entities e, jsonb_each_text(e.provenance) p
WHERE NOT EXISTS (SELECT 1 FROM data_sources WHERE code = p.value)
