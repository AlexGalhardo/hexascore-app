package library_test

import (
	"fmt"
	"math/rand/v2" //nolint:depguard // reproducible fixtures, not secrets
	"net/http"
	"slices"
	"testing"

	"github.com/AlexGalhardo/vukem/errs"

	"github.com/AlexGalhardo/hexascore-app/app/library"
)

func ids(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("item-%d", i)
	}

	return out
}

// TestReorderProperty: any permutation of the current set is accepted with dense positions 1..n and the same
// set; a stale version is a 409 and a changed set a 422.
func TestReorderProperty(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))

	for range 10000 {
		n := rng.IntN(40)
		current := ids(n)
		want := slices.Clone(current)
		rng.Shuffle(n, func(i, j int) { want[i], want[j] = want[j], want[i] })
		version := rng.IntN(1000)

		pos, err := library.PlanOrder(current, version, version, want)
		if err != nil {
			t.Fatalf("permutation of %d rejected: %v", n, err)
		}
		if len(pos) != n {
			t.Fatalf("set changed: %d positions for %d items", len(pos), n)
		}
		seen := make([]bool, n+1)
		for i, id := range want {
			p := pos[id]
			if p != i+1 || seen[p] {
				t.Fatalf("positions not dense: %s at %d, want %d", id, p, i+1)
			}
			seen[p] = true
		}

		if _, err := library.PlanOrder(current, version+1, version, want); errs.Status(err) != http.StatusConflict {
			t.Fatalf("stale version: %v", err)
		}

		if n == 0 {
			continue
		}
		cases := map[string][]string{
			"missing": want[1:],
			"extra":   append(slices.Clone(want), "item-new"),
			"unknown": append(slices.Clone(want[1:]), "item-new"),
		}
		if n >= 2 {
			dup := slices.Clone(want)
			dup[n-1] = dup[0]
			cases["duplicate"] = dup
		}
		for name, bad := range cases {
			if _, err := library.PlanOrder(
				current,
				version,
				version,
				bad,
			); errs.Status(
				err,
			) != http.StatusUnprocessableEntity {
				t.Fatalf("%s order of %d accepted: %v", name, n, err)
			}
		}
	}
}
