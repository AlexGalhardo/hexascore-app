-- Local preview only: keeps the first 1000 items of each type (by id) and removes orphaned child rows.
-- Runs after the full catalogue restore so every foreign key is still in place when the schema is built.
-- Foreign-key triggers are disabled only for the delete; the orphan sweep below restores referential integrity.
BEGIN;

CREATE TEMP TABLE keep_items ON COMMIT DROP AS
SELECT id FROM (
    SELECT id, row_number() OVER (PARTITION BY type ORDER BY id) AS rn FROM items
) ranked
WHERE rn <= 1000;

SET LOCAL session_replication_role = replica;
DELETE FROM items WHERE id NOT IN (SELECT id FROM keep_items);

-- Several passes: removing a child row can orphan its own children.
DO $$
DECLARE
    c record;
    pass int;
BEGIN
    FOR pass IN 1..4 LOOP
        FOR c IN
            SELECT k.conrelid::regclass AS child,
                   k.confrelid::regclass AS parent,
                   a.attname AS child_col,
                   pa.attname AS parent_col
            FROM pg_constraint k
            JOIN pg_attribute a ON a.attrelid = k.conrelid AND a.attnum = k.conkey[1]
            JOIN pg_attribute pa ON pa.attrelid = k.confrelid AND pa.attnum = k.confkey[1]
            WHERE k.contype = 'f' AND cardinality(k.conkey) = 1
              AND k.connamespace = 'public'::regnamespace
        LOOP
            EXECUTE format(
                'DELETE FROM %s WHERE %I IS NOT NULL AND %I NOT IN (SELECT %I FROM %s)',
                c.child, c.child_col, c.child_col, c.parent_col, c.parent
            );
        END LOOP;
    END LOOP;
END $$;

COMMIT;
