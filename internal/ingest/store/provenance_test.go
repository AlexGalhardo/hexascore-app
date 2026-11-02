package store

import "testing"

func TestConfidenceFormula(t *testing.T) {
	for _, c := range []struct {
		source string
		s      Signals
		want   int
	}{
		{"wikidata", Signals{}, 90},
		{"wikidata", Signals{Agree: true}, 100},
		{"tmdb", Signals{Disagree: true}, 75},
		{"gog", Signals{Disagree: true, FailedDQ: true}, 40},
		{"llm", Signals{}, 50},
		{"mystery", Signals{}, 50},
		{"steam", Signals{FailedDQ: true, Disagree: true}, 45},
	} {
		if got := Confidence(c.source, c.s); got != c.want {
			t.Errorf("Confidence(%s, %+v) = %d, want %d", c.source, c.s, got, c.want)
		}
	}

	if !(FieldProv{Confidence: 59}).NeedsReview() || (FieldProv{Confidence: 60}).NeedsReview() {
		t.Error("the review floor is 60")
	}
}
