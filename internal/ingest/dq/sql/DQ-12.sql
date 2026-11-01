SELECT entity_id::text || ':' || type FROM entity_stats WHERE refreshed_at < now() - interval '2 hours'
