package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029100010CreateIsbn13Function is SDD 01 §5.2 `isbn13_is_valid`: 13 digits with a 978/979 prefix and a correct check digit (ADR 0012). The Go mirror is catalog.ValidISBN13.
func m20261029100010CreateIsbn13Function() schema.Migration {
	return schema.Migration{
		ID: "20261029100010_create_isbn13_function",
		Up: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil // scaffold tests run on SQLite (VF-033)
			}
			return s.Raw(`CREATE FUNCTION isbn13_is_valid(t text) RETURNS boolean LANGUAGE sql IMMUTABLE AS $$
  SELECT t ~ '^97[89][0-9]{10}$' AND (
    SELECT sum(substr(t, i, 1)::int * CASE WHEN i % 2 = 0 THEN 3 ELSE 1 END)
    FROM generate_series(1, 13) i) % 10 = 0 $$`)
		},
		Down: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil
			}
			return s.Raw(`DROP FUNCTION IF EXISTS isbn13_is_valid(text)`)
		},
	}
}
