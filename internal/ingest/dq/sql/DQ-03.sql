SELECT e.id::text FROM entities e
WHERE e.created_at < now() - interval '7 days'
  AND NOT EXISTS (SELECT 1 FROM item_entities ie WHERE ie.entity_id = e.id)
  AND NOT EXISTS (SELECT 1 FROM editions d WHERE d.publisher_id = e.id OR d.imprint_id = e.id)
  AND NOT EXISTS (SELECT 1 FROM entity_links l WHERE l.from_id = e.id OR l.to_id = e.id)
