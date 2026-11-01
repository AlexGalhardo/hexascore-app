SELECT r.scope || ':' || r.namespace || ':' || coalesce(r.old_slug, r.old_id::text) FROM catalog_redirects r
WHERE (r.scope = 'item' AND NOT EXISTS (SELECT 1 FROM items WHERE id = r.target_id))
   OR (r.scope = 'entity' AND NOT EXISTS (SELECT 1 FROM entities WHERE id = r.target_id))
   OR EXISTS (SELECT 1 FROM catalog_redirects o WHERE o.scope = r.scope AND o.old_id = r.target_id)
