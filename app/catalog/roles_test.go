package catalog

import "testing"

func TestRoleCompatibility(t *testing.T) {
	cases := []struct {
		role, kind, itemType string
		want                 bool
	}{
		{"director", "person", "movie", true},
		{"director", "person", "book", false},
		{"director", "organization", "movie", false},
		{"author", "person", "book", true},
		{"author", "person", "movie", false},
		{"developer", "organization", "game", true},
		{"developer", "organization", "book", false},
		{"publisher", "organization", "book", true},
		{"publisher", "organization", "movie", false},
		{"network", "organization", "series", true},
		{"network", "organization", "movie", false},
		{"platform", "platform", "game", true},
		{"platform", "platform", "movie", false},
		{"franchise", "franchise", "book", true},
		{"genre", "genre", "series", true},
		{"genre", "person", "series", false},
		{"original_language", "language", "movie", true},
		{"translator", "person", "book", false},
		{"engine", "organization", "game", false},
	}
	for _, c := range cases {
		if got := RoleCompatible(c.role, c.kind, c.itemType); got != c.want {
			t.Errorf("RoleCompatible(%q, %q, %q) = %v, want %v", c.role, c.kind, c.itemType, got, c.want)
		}
	}
}
