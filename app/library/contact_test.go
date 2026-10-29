package library

import (
	"testing"
	"time"
)

func TestContactHoneypot(t *testing.T) {
	// A nil DB proves nothing is stored: the honeypot answers before any query.
	c := &Contact{Key: make([]byte, 32), Now: time.Now}
	if err := c.Submit(t.Context(), ContactInput{Honeypot: "http://spam"}); err != nil {
		t.Fatalf("honeypot must look like success, got %v", err)
	}
}

func TestContactTokenMinimumFill(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	c := &Contact{Key: make([]byte, 32), Now: func() time.Time { return now }}
	tok := c.Token()

	cases := map[string]struct {
		at   time.Duration
		tok  string
		want bool
	}{
		"too fast":  {time.Second, tok, false},
		"on time":   {ContactMinFill, tok, true},
		"expired":   {3 * time.Hour, tok, false},
		"tampered":  {time.Minute, "1.abc", false},
		"empty":     {time.Minute, "", false},
		"forged ts": {time.Minute, "1700000000." + c.mac("1800000000"), false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c.Now = func() time.Time { return now.Add(tc.at) }
			if got := c.tokenOK(tc.tok); got != tc.want {
				t.Fatalf("tokenOK = %v, want %v", got, tc.want)
			}
		})
	}
}
