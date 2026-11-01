SELECT item_id::text || ':' || locale FROM item_translations
WHERE synopsis IS NOT NULL
  AND (synopsis_source IS NULL OR synopsis_licence IS NULL
       OR (synopsis_source LIKE 'wikipedia%' AND coalesce(synopsis_source_url, '') = ''))
