package backup

import (
	"errors"
	"slices"
	"testing"
)

func TestVerifyRejectsSmallDump(t *testing.T) {
	t.Parallel()

	week := []int64{1000, 1100, 900, 1050, 950, 1000, 1020} // median 1000

	for _, c := range []struct {
		size   int64
		recent []int64
		fail   bool
	}{
		{500, week, true}, // exactly 50 % fails
		{499, week, true},
		{501, week, false},
		{10, nil, false},                  // first dump: nothing to compare
		{1000, []int64{1000, 3000}, true}, // even count: median 2000
	} {
		err := CheckSize(c.size, c.recent)
		if got := errors.Is(err, ErrSmallDump); got != c.fail {
			t.Errorf("CheckSize(%d, %v) = %v, want fail=%v", c.size, c.recent, err, c.fail)
		}
	}
}

func TestMissingTables(t *testing.T) {
	t.Parallel()

	toc := []byte(`;
; Archive created at 2026-10-29 09:00:00 UTC
3401; 1259 16390 TABLE public items hexascore
3402; 1259 16400 TABLE public users hexascore
3403; 1259 16410 TABLE public user_ratings_view hexascore
3404; 0 16390 TABLE DATA public items hexascore
`)

	if got := MissingTables(toc); !slices.Equal(got, []string{"ratings"}) {
		t.Fatalf("missing = %v, want [ratings]", got)
	}
}
