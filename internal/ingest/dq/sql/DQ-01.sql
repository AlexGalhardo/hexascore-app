SELECT i.id::text FROM items i
WHERE NOT EXISTS (SELECT 1 FROM item_translations t WHERE t.item_id = i.id AND t.locale = 'en' AND length(btrim(t.title)) > 0)
