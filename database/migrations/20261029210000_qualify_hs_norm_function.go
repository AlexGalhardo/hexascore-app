package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029210000QualifyHsNormFunction schema-qualifies the hs_unaccent call
// inside hs_norm: pg_restore runs with an empty search_path, so the
// unqualified call failed while restoring the generated search columns and
// no dump could be restored (found by the P2.E.3.1 drill test). Same result
// for every input, so columns and indexes stay valid; no SET search_path, which
// would stop the planner from inlining the function.
func m20261029210000QualifyHsNormFunction() schema.Migration {
	return postgresMigration("20261029210000_qualify_hs_norm_function",
		[]string{
			`CREATE OR REPLACE FUNCTION hs_norm(text) RETURNS text LANGUAGE sql IMMUTABLE PARALLEL SAFE STRICT
  AS $$ SELECT btrim(regexp_replace(lower(public.hs_unaccent($1)), '[[:punct:][:space:]]+', ' ', 'g')) $$`,
		},
		[]string{
			`CREATE OR REPLACE FUNCTION hs_norm(text) RETURNS text LANGUAGE sql IMMUTABLE PARALLEL SAFE STRICT
  AS $$ SELECT btrim(regexp_replace(lower(hs_unaccent($1)), '[[:punct:][:space:]]+', ' ', 'g')) $$`,
		},
	)
}
