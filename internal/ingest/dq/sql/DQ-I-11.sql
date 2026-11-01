SELECT id::text FROM entity_match_reviews WHERE status = 'open' AND created_at < now() - interval '7 days'
