package editions

import (
	"context"
	"slices"
)

// MaxPerWork caps the editions kept for one work (SDD 06 §7.4).
const MaxPerWork = 50

// Candidate is one edition found for a work. ISBN is raw on input and an ISBN-13 once Select has run.
type Candidate struct {
	ISBN          string
	QID           string
	OpenLibraryID string
	Title         string
	Language      string
	MarketCountry string
	Year          int
	HasCover      bool
	Source        string
	// PublisherQID and PublisherName are the edition's P123 publisher; PublisherID is the entity SaveAll resolved.
	PublisherQID  string
	PublisherName string
	PublisherID   string
}

// Language maps a MARC language code to the editions.language value; anything unknown is "und" (DQ-10).
func Language(marc string) string {
	switch marc {
	case "por":
		return "pt"
	case "eng":
		return "en"
	case "spa":
		return "es"
	}

	return "und"
}

func rank(lang string) int {
	if i := slices.Index([]string{"pt", "en", "es"}, lang); i >= 0 {
		return i
	}

	return 3
}

// Select orders the candidates pt, en, es first, normalises their ISBNs to ISBN-13, drops the repeated ones and
// keeps at most limit. An invalid ISBN is dropped and counted; its candidate survives only through a QID or an
// Open Library id (the editions CHECK needs one identifier).
func Select(in []Candidate, limit int) ([]Candidate, int) {
	sorted := slices.Clone(in)
	slices.SortStableFunc(sorted, func(a, b Candidate) int { return rank(a.Language) - rank(b.Language) })

	var (
		out     []Candidate
		dropped int
		seen    = map[string]bool{}
	)

	for _, c := range sorted {
		if c.ISBN != "" {
			isbn, ok := ToISBN13(c.ISBN)
			if !ok {
				dropped++
			}

			c.ISBN = isbn
		}

		if c.ISBN == "" && c.QID == "" && c.OpenLibraryID == "" {
			continue
		}

		if c.ISBN != "" {
			if seen[c.ISBN] {
				continue
			}

			seen[c.ISBN] = true
		}

		if len(out) < limit {
			out = append(out, c)
		}
	}

	return out, dropped
}

// PrimaryEdition returns the index of the earliest edition in the work's original language, or -1. An edition with a
// cover beats one without (SDD 06 §7.4); ingest writes no edition covers yet, so until it does the earliest edition in
// the original language is primary. An undated edition comes after every dated one.
func PrimaryEdition(in []Candidate, original string) int {
	best := -1

	var top Candidate

	for i, c := range in {
		if c.Language != original {
			continue
		}

		if best < 0 || better(c, top) {
			best, top = i, c
		}
	}

	return best
}

func better(a, b Candidate) bool {
	if a.HasCover != b.HasCover {
		return a.HasCover
	}

	return a.Year > 0 && (b.Year == 0 || a.Year < b.Year)
}

// Store is the persistence Save needs.
type Store interface {
	WorkOfISBN(ctx context.Context, isbn13 string) (work string, found bool, err error)
	WorkOfQID(ctx context.Context, qid string) (work string, found bool, err error)
	Upsert(ctx context.Context, work string, c Candidate) error
}

// Outcome says what Save did.
type Outcome string

const (
	// Saved means the edition was written for the work.
	Saved Outcome = "saved"
	// Conflict means the ISBN belongs to another work; nothing moved, the caller records the failure and a review.
	Conflict Outcome = "conflict"
)

// Save writes one edition unless its ISBN or edition QID is already held by a different work; neither is ever moved.
func Save(ctx context.Context, s Store, work string, c Candidate) (Outcome, error) {
	if c.ISBN != "" {
		owner, found, err := s.WorkOfISBN(ctx, c.ISBN)
		if err != nil {
			return "", err
		}

		if found && owner != work {
			return Conflict, nil
		}
	}

	if c.QID != "" {
		owner, found, err := s.WorkOfQID(ctx, c.QID)
		if err != nil {
			return "", err
		}

		if found && owner != work {
			return Conflict, nil
		}
	}

	return Saved, s.Upsert(ctx, work, c)
}
