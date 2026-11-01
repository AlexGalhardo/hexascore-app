SELECT id::text FROM items
WHERE (release_status = 'released' AND release_date IS NULL AND release_precision <> 'unknown')
   OR (release_status = 'announced' AND release_precision = 'day' AND release_date < current_date - 30)
