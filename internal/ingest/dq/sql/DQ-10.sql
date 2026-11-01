SELECT i.id::text FROM items i WHERE i.type = 'book' AND NOT EXISTS (SELECT 1 FROM editions WHERE work_id = i.id)
UNION ALL
SELECT id::text FROM editions WHERE language = 'und'
