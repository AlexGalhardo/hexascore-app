package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029140001CreateHsNormFunction is SDD 07 §5: `unaccent()` is STABLE, so IMMUTABLE wrappers are what generated columns and index expressions may call. Go never normalises; SQL calls hs_norm.
func m20261029140001CreateHsNormFunction() schema.Migration {
	return postgresMigration("20261029140001_create_hs_norm_function",
		[]string{
			`CREATE FUNCTION hs_unaccent(text) RETURNS text LANGUAGE sql IMMUTABLE PARALLEL SAFE STRICT
  AS $$ SELECT public.unaccent('public.unaccent'::regdictionary, $1) $$`,
			`CREATE FUNCTION hs_norm(text) RETURNS text LANGUAGE sql IMMUTABLE PARALLEL SAFE STRICT
  AS $$ SELECT btrim(regexp_replace(lower(hs_unaccent($1)), '[[:punct:][:space:]]+', ' ', 'g')) $$`,
		},
		[]string{
			`DROP FUNCTION IF EXISTS hs_norm(text)`,
			`DROP FUNCTION IF EXISTS hs_unaccent(text)`,
		},
	)
}
