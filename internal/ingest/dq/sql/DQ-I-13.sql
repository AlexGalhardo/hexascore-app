SELECT source || ':' || external_id FROM source_records
WHERE source = 'igdb' AND checked_at < now() - interval '7 days'
