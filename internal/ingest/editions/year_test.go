package editions

import "testing"

func TestPlausibleYearKeepsOnlyWhatTheCheckAccepts(t *testing.T) {
	for in, want := range map[int]int{0: 0, 1399: 0, 1400: 1400, 1999: 1999, 2100: 2100, 2101: 0, 9999: 0, -5: 0} {
		if got := plausibleYear(in); got != want {
			t.Errorf("plausibleYear(%d) = %d, want %d", in, got, want)
		}
	}
}
