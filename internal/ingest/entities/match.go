package entities

import (
	"context"
	"slices"
)

// Credit is one role of an item to resolve. QID and the external id are optional; Year is 0 when unknown.
type Credit struct {
	ItemID        string
	Role          string
	Kind          string
	Source        string
	QID           string
	ExternalIDKey string
	ExternalID    string
	Name          string
	Year          int
	Position      int
	Character     string
}

// Entity is a stored candidate; Year is 0 when it has no birth or founding year.
type Entity struct {
	ID   string
	Kind string
	Year int
}

// Outcome says what Resolve did with a credit.
type Outcome string

const (
	// Linked means the credit points to an existing entity.
	Linked Outcome = "linked"
	// Reviewed means the match is ambiguous and waits in entity_match_reviews.
	Reviewed Outcome = "reviewed"
	// Created means no entity matched and a native one was made.
	Created Outcome = "created"
	// Skipped means the credit carries no id and no usable name.
	Skipped Outcome = "skipped"
)

// Store is the persistence Resolve needs. Link writes the item_entities row (and keeps ids the entity lacked).
type Store interface {
	ByQID(ctx context.Context, qid string) (Entity, bool, error)
	ByExternalID(ctx context.Context, key, id, kind string) ([]Entity, error)
	ByName(ctx context.Context, normalized, kind string) ([]Entity, error)
	Create(ctx context.Context, c Credit) (string, error)
	Link(ctx context.Context, entityID string, c Credit) error
	OpenReview(ctx context.Context, c Credit, candidates []string) error
}

// Resolve matches one credit: (1) QID; (2) source external id; (3) normalised name + kind + year ±1 or unknown.
// One candidate links; two or more, or a QID stored under another kind, open a review; none creates a native entity.
func Resolve(ctx context.Context, s Store, c Credit) (Outcome, error) {
	if c.QID != "" {
		e, ok, err := s.ByQID(ctx, c.QID)
		if err != nil {
			return "", err
		}

		if ok && e.Kind != c.Kind {
			return Reviewed, s.OpenReview(ctx, c, []string{e.ID})
		}

		if ok {
			return Linked, s.Link(ctx, e.ID, c)
		}
	}

	if c.ExternalIDKey != "" && c.ExternalID != "" {
		cands, err := s.ByExternalID(ctx, c.ExternalIDKey, c.ExternalID, c.Kind)
		if err != nil {
			return "", err
		}

		if out, done, err := decide(ctx, s, c, cands); done {
			return out, err
		}
	}

	key := NormalizeName(c.Name)
	if key == "" && c.QID == "" && c.ExternalID == "" {
		return Skipped, nil
	}

	if key != "" {
		cands, err := s.ByName(ctx, key, c.Kind)
		if err != nil {
			return "", err
		}

		cands = slices.DeleteFunc(cands, func(e Entity) bool { return !sameYear(c.Year, e.Year) })

		if out, done, err := decide(ctx, s, c, cands); done {
			return out, err
		}
	}

	id, err := s.Create(ctx, c)
	if err != nil {
		return "", err
	}

	return Created, s.Link(ctx, id, c)
}

// decide handles the one-candidate and many-candidate cases; done is false when there is none.
func decide(ctx context.Context, s Store, c Credit, cands []Entity) (Outcome, bool, error) {
	switch len(cands) {
	case 0:
		return "", false, nil
	case 1:
		return Linked, true, s.Link(ctx, cands[0].ID, c)
	}

	ids := make([]string, len(cands))
	for i, e := range cands {
		ids[i] = e.ID
	}

	slices.Sort(ids)

	return Reviewed, true, s.OpenReview(ctx, c, ids)
}

func sameYear(a, b int) bool {
	return a == 0 || b == 0 || a-b <= 1 && b-a <= 1
}
