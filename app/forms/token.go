// Package forms signs the render time of public forms, so a submit can enforce the minimum fill time
// of SDD 13 §7.2 without server state.
package forms

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strconv"
	"strings"
	"time"
)

// Fill-time bounds shared by the honeypot forms (sign-up, contact).
const (
	MinFill = 3 * time.Second
	TTL     = 2 * time.Hour
)

// Token issues and checks "<unix render time>.<HMAC>" tokens bound to a form name.
type Token struct {
	Key []byte
	Now func() time.Time
}

func (t Token) mac(form, ts string) string {
	m := hmac.New(sha256.New, t.Key)
	m.Write([]byte(form + ":" + ts))

	return base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}

// Issue returns the token for form rendered now.
func (t Token) Issue(form string) string { return t.IssueAt(form, t.Now()) }

// IssueAt returns the token for form rendered at at.
func (t Token) IssueAt(form string, at time.Time) string {
	ts := strconv.FormatInt(at.Unix(), 10)

	return ts + "." + t.mac(form, ts)
}

// Valid reports whether token was issued for form, at least MinFill and at most TTL ago.
func (t Token) Valid(form, token string) bool {
	ts, sig, ok := strings.Cut(token, ".")
	if !ok || !hmac.Equal([]byte(sig), []byte(t.mac(form, ts))) {
		return false
	}

	sec, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return false
	}

	age := t.Now().Sub(time.Unix(sec, 0))

	return age >= MinFill && age <= TTL
}
