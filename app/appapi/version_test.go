package appapi

import "testing"

func TestVersionBelow(t *testing.T) {
	t.Parallel()

	cases := []struct {
		client, min string
		want        bool
	}{
		{"1.2.0", "1.2.0", false},
		{"1.2.1", "1.2.0", false},
		{"1.10.0", "1.2.0", false},
		{"1.1.9", "1.2.0", true},
		{"", "1.0.0", true},
		{"nope", "1.0.0", true},
		{"0.0.0", "0.0.0", false},
		{"1.0.0", "banana", true},
	}

	for _, tc := range cases {
		if got := versionBelow(tc.client, tc.min); got != tc.want {
			t.Errorf("versionBelow(%q, %q) = %v, want %v", tc.client, tc.min, got, tc.want)
		}
	}
}
