package library

import (
	"net/http"

	"github.com/AlexGalhardo/vukem/errs"
)

// Problem codes of the list item writes (SDD 14 §6.3).
const (
	CodeItemLimit       = "list-item-limit-reached"
	CodeVersionConflict = "list-version-conflict"
	CodeOrderMismatch   = "order-mismatch"
)

func errVersionConflict() error {
	return errs.NewKey(http.StatusConflict, "lists.version_conflict", "The list changed; reload it and retry.", nil).
		WithExtension("code", CodeVersionConflict)
}

func errOrderMismatch() error {
	return errs.NewKey(http.StatusUnprocessableEntity, "lists.order_mismatch",
		"The order must contain every item of the list exactly once.", nil).WithExtension("code", CodeOrderMismatch)
}

// PlanOrder validates a full-permutation reorder (SDD 14 §7.1): a stale version is a 409, and ids that are not
// exactly the current set (missing, extra, unknown or duplicated) a 422. It returns the dense positions 1..n.
func PlanOrder(current []string, have, want int, ids []string) (map[string]int, error) {
	if have != want {
		return nil, errVersionConflict()
	}
	if len(ids) != len(current) {
		return nil, errOrderMismatch()
	}

	known := make(map[string]struct{}, len(current))
	for _, id := range current {
		known[id] = struct{}{}
	}

	pos := make(map[string]int, len(ids))
	for i, id := range ids {
		if _, ok := known[id]; !ok {
			return nil, errOrderMismatch()
		}
		if _, dup := pos[id]; dup {
			return nil, errOrderMismatch()
		}
		pos[id] = i + 1
	}

	return pos, nil
}
