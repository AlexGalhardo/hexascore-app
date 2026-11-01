package catalog

import "slices"

// roleRule is one row of the SDD 01 §5.2 role compatibility table.
type roleRule struct {
	kinds     []string
	itemTypes []string // nil: every item type
}

var (
	allTypes  = []string{"game", "movie", "series", "book"}
	filmTypes = []string{"movie", "series"}
	roleRules = map[string]roleRule{
		"director":           {[]string{"person"}, filmTypes},
		"writer":             {[]string{"person"}, filmTypes},
		"cast":               {[]string{"person"}, filmTypes},
		"creator":            {[]string{"person"}, filmTypes},
		"author":             {[]string{"person"}, []string{"book"}},
		"developer":          {[]string{"organization"}, []string{"game"}},
		"publisher":          {[]string{"organization"}, []string{"game", "book"}},
		"production_company": {[]string{"organization"}, filmTypes},
		"network":            {[]string{"organization"}, []string{"series"}},
		"platform":           {[]string{"platform"}, []string{"game"}},
		"game_mode":          {[]string{"game_mode"}, []string{"game"}},
		"franchise":          {[]string{"franchise"}, allTypes},
		"series":             {[]string{"franchise"}, allTypes},
		"genre":              {[]string{"genre"}, allTypes},
		"country":            {[]string{"country"}, allTypes},
		"original_language":  {[]string{"language"}, allTypes},
	}
)

// RoleCompatible reports whether an entity of entityKind may hold role on an
// item of itemType. Unknown roles are never compatible. The database only
// checks the role and kind enums, so writers must call this first (DQ-08
// audits existing rows).
func RoleCompatible(role, entityKind, itemType string) bool {
	rule, ok := roleRules[role]
	if !ok {
		return false
	}
	return slices.Contains(rule.kinds, entityKind) && slices.Contains(rule.itemTypes, itemType)
}

// KnownRole reports whether role is an item_entities role.
func KnownRole(role string) bool {
	_, ok := roleRules[role]

	return ok
}

// RoleKind returns the entity kind a role holds, for creating a native entity from a credit.
func RoleKind(role string) (string, bool) {
	rule, ok := roleRules[role]
	if !ok {
		return "", false
	}

	return rule.kinds[0], true
}
