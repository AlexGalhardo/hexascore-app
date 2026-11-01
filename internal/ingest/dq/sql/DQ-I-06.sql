SELECT source || ':' || external_id FROM source_records
WHERE source = 'tmdb' AND checked_at < now() - interval '180 days'
