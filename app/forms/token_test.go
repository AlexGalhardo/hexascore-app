package forms_test

import (
	"testing"
	"time"

	"github.com/AlexGalhardo/hexascore-app/app/forms"
)

func TestToken(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_800_000_000, 0)
	tok := forms.Token{Key: []byte("0123456789abcdef0123456789abcdef"), Now: func() time.Time { return now }}
	other := forms.Token{Key: []byte("fedcba9876543210fedcba9876543210"), Now: tok.Now}

	for name, c := range map[string]struct {
		token string
		want  bool
	}{
		"filled in 3 s":      {tok.IssueAt("signup", now.Add(-forms.MinFill)), true},
		"filled in 2 s":      {tok.IssueAt("signup", now.Add(-2*time.Second)), false},
		"older than the TTL": {tok.IssueAt("signup", now.Add(-forms.TTL-time.Second)), false},
		"other form":         {tok.IssueAt("contact", now.Add(-time.Minute)), false},
		"other key":          {other.IssueAt("signup", now.Add(-time.Minute)), false},
		"empty":              {"", false},
		"garbage":            {"x.y", false},
	} {
		if got := tok.Valid("signup", c.token); got != c.want {
			t.Errorf("%s: Valid = %v, want %v", name, got, c.want)
		}
	}
}
