SELECT id::text FROM items WHERE slug !~ '^[a-z0-9]+(-[a-z0-9]+)*$'
UNION ALL
SELECT i.id::text FROM items i
JOIN catalog_redirects r ON r.scope = 'item' AND r.namespace = i.type AND r.old_slug = i.slug
UNION ALL
SELECT id::text FROM entities WHERE slug !~ '^[a-z0-9]+(-[a-z0-9]+)*$'
UNION ALL
SELECT e.id::text FROM entities e
JOIN catalog_redirects r ON r.scope = 'entity' AND r.namespace = e.kind AND r.old_slug = e.slug
