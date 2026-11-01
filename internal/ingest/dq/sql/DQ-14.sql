SELECT a.from_item_id::text || ':' || a.to_item_id::text FROM item_links a
JOIN item_links b ON b.from_item_id = a.to_item_id AND b.to_item_id = a.from_item_id AND b.relation = a.relation
WHERE a.from_item_id < a.to_item_id
