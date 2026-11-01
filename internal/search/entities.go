package search

// EntityNameMatch is the entities.default_name predicate the entities_name_norm_idx index covers, for the credit
// matcher in internal/ingest/entities; param is the placeholder of the already normalised name.
func EntityNameMatch(param string) string {
	return "hs_norm(default_name) = " + param
}
