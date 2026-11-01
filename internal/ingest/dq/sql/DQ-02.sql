SELECT id::text FROM editions WHERE isbn13 IS NOT NULL AND NOT isbn13_is_valid(isbn13)
UNION ALL
SELECT (array_agg(id::text))[1] FROM editions WHERE isbn13 IS NOT NULL GROUP BY isbn13 HAVING count(*) > 1
