package similar

// CastPositions is how many billed cast members count as features (position < 5).
const CastPositions = 5

// ParentWeight is the multiplier of a franchise reached through a part_of parent.
const ParentWeight = 0.5

// Feature is a (role, entity) pair taken from item_entities.
type Feature struct {
	Role   string
	Entity string
}

// Entry is one item_entities row.
type Entry struct {
	Role     string
	Entity   string
	Position int
}

// ExtractFeatures builds F_i: every entry as a feature with multiplier 1, except cast beyond CastPositions (dropped).
// parents maps a franchise entity to its part_of parents (entity_links); each parent joins at ParentWeight unless the item already holds it directly.
func ExtractFeatures(entries []Entry, parents map[string][]string) map[Feature]float64 {
	out := make(map[Feature]float64, len(entries))

	for _, e := range entries {
		if e.Role == "cast" && e.Position >= CastPositions {
			continue
		}

		out[Feature{e.Role, e.Entity}] = 1
	}

	for _, e := range entries {
		if e.Role != "franchise" {
			continue
		}

		for _, p := range parents[e.Entity] {
			f := Feature{"franchise", p}
			if _, ok := out[f]; !ok {
				out[f] = ParentWeight
			}
		}
	}

	return out
}
