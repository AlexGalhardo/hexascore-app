DO $$ DECLARE c text; BEGIN
  FOR c IN SELECT conname FROM pg_constraint WHERE conrelid = 'editions'::regclass AND pg_get_constraintdef(oid) ILIKE '%isbn13_is_valid%' LOOP
    EXECUTE format('ALTER TABLE editions DROP CONSTRAINT %I', c);
  END LOOP;
END $$;
UPDATE editions SET isbn13 = '9780306406158'
