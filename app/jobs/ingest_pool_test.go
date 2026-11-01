package jobs

import "testing"

func TestCheckPool(t *testing.T) {
	for _, tc := range []struct {
		in string
		ok bool
	}{{"", true}, {"10", true}, {"2", true}, {"0", true}, {"1", false}, {"-3", false}, {"many", false}} {
		if err := checkPool(tc.in); (err == nil) != tc.ok {
			t.Errorf("checkPool(%q) = %v, want ok=%v", tc.in, err, tc.ok)
		}
	}
}
