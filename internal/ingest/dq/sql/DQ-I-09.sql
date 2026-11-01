SELECT id::text FROM items
WHERE release_date IS NOT NULL
  AND (release_date > current_date + interval '10 years'
       OR release_date < CASE WHEN type = 'book' THEN date '1400-01-01' ELSE date '1870-01-01' END)
