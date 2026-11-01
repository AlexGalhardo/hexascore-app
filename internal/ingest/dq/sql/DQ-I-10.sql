SELECT (array_agg(i.id::text))[1] FROM items i
JOIN item_translations t ON t.item_id = i.id AND t.locale = 'en'
GROUP BY i.type, lower(t.title), extract(year FROM i.release_date) HAVING count(*) > 1
UNION ALL
SELECT (array_agg(id::text))[1] FROM entities GROUP BY kind, lower(default_name) HAVING count(*) > 1
