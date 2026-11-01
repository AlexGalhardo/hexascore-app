// Package logging holds the privacy-preserving log handler (SDD 12 §6.1).
package logging

import (
	"cmp"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net/netip"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/AlexGalhardo/vukem/web"
)

// RequestMessage is the message of vukem's request log line
// (middleware.Logger).
const RequestMessage = "request"

// Redacted replaces a sensitive value.
const Redacted = "[REDACTED]"

// sampleEvery keeps 1 in sampleEvery fast 2xx/3xx lines of the noisy routes.
const sampleEvery = 100

// slowRequest lines are always kept.
const slowRequest = time.Second

// sensitiveKeys never carry their value into a log line (FR-12-8): secrets,
// contact data and raw search text.
var sensitiveKeys = map[string]bool{
	"password": true, "password_confirmation": true, "current_password": true, "new_password": true,
	"token": true, "access_token": true, "refresh_token": true, "secret": true, "api_key": true, "apikey": true,
	"authorization": true, "cookie": true, "set-cookie": true, "email": true, "e-mail": true,
	"q": true, "query": true, "search": true, "term": true,
}

var (
	emailRe  = regexp.MustCompile(`[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`)
	bearerRe = regexp.MustCompile(`(?i)\b(bearer|basic)\s+[^\s,;"]+`)
	keyRe    = regexp.MustCompile(`\b(hx_live_|hx_app_|sk_live_|sk_test_|rk_live_|whsec_|re_)[A-Za-z0-9_]{8,}`)
	paramRe  = regexp.MustCompile(`(?i)\b(password|token|api_key|apikey|secret|email|q|query|search)=[^&\s"]*`)
)

// Scrub removes emails, bearer credentials, API keys and sensitive query
// parameters from free text (error messages, URLs).
func Scrub(s string) string {
	s = emailRe.ReplaceAllString(s, Redacted)
	s = bearerRe.ReplaceAllString(s, "$1 "+Redacted)
	s = keyRe.ReplaceAllString(s, "${1}"+Redacted)

	return paramRe.ReplaceAllString(s, "${1}="+Redacted)
}

// TruncateIP keeps the network part only: /24 for IPv4, /48 for IPv6.
// Anything unparsable is dropped.
func TruncateIP(s string) string {
	addr, err := netip.ParseAddr(s)
	if err != nil {
		return ""
	}

	addr = addr.Unmap()

	bits := 48
	if addr.Is4() {
		bits = 24
	}

	p, err := addr.Prefix(bits)
	if err != nil {
		return ""
	}

	return p.Addr().String()
}

// UserHash is hex(HMAC-SHA256(key, id))[:16] (SDD 12 §6.1, NFR-13-4).
func UserHash(key []byte, id string) string {
	m := hmac.New(sha256.New, key)
	_, _ = m.Write([]byte(id))

	return hex.EncodeToString(m.Sum(nil))[:16]
}

// userAttr carries the raw user id from SetUser to the request line, where Handle replaces it with user_hash.
const userAttr = "user_id"

// SetUser records the authenticated user id for the request line, which
// carries it hashed; a no-op outside a request logged by vukem's Logger.
func SetUser(ctx context.Context, id string) {
	web.AddLogAttrs(ctx, slog.String(userAttr, id))
}

// Handler redacts every record and reshapes vukem's request line: ip and
// peer truncated, duration in ms, route and user_hash added, fast 2xx/3xx
// lines of noisy routes sampled.
type Handler struct {
	next slog.Handler
	key  []byte
	seen *atomic.Uint64
}

// NewHandler wraps next; key is the user-hash key (app.DeriveKey("log", 32)).
func NewHandler(next slog.Handler, key []byte) *Handler {
	return &Handler{next: next, key: key, seen: new(atomic.Uint64)}
}

// Enabled implements slog.Handler.
func (h *Handler) Enabled(ctx context.Context, l slog.Level) bool { return h.next.Enabled(ctx, l) }

// WithAttrs implements slog.Handler.
func (h *Handler) WithAttrs(as []slog.Attr) slog.Handler {
	out := make([]slog.Attr, len(as))
	for i, a := range as {
		out[i] = redact(a)
	}

	return &Handler{next: h.next.WithAttrs(out), key: h.key, seen: h.seen}
}

// WithGroup implements slog.Handler.
func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{next: h.next.WithGroup(name), key: h.key, seen: h.seen}
}

// Handle implements slog.Handler.
func (h *Handler) Handle(ctx context.Context, rec slog.Record) error {
	out := slog.NewRecord(rec.Time, rec.Level, Scrub(rec.Message), rec.PC)

	if rec.Message != RequestMessage {
		rec.Attrs(func(a slog.Attr) bool {
			out.AddAttrs(redact(a))
			return true
		})

		return h.next.Handle(ctx, out)
	}

	var (
		path   string
		status int64
		dur    time.Duration
	)

	route, user := "unmatched", ""

	rec.Attrs(func(a slog.Attr) bool {
		switch a.Key {
		case "ip", "peer":
			a.Value = slog.StringValue(TruncateIP(a.Value.String()))
		case "path":
			path = a.Value.String()
		case "status":
			status = a.Value.Int64()
		case "duration":
			dur = a.Value.Duration()
			a = slog.Int64("duration_ms", dur.Milliseconds())
		case "route":
			route = cmp.Or(a.Value.String(), route)

			return true
		case userAttr:
			user = a.Value.String()

			return true
		}

		out.AddAttrs(redact(a))

		return true
	})

	if !h.keep(path, status, dur) {
		return nil
	}

	out.AddAttrs(slog.String("route", route))

	if user != "" {
		out.AddAttrs(slog.String("user_hash", UserHash(h.key, user)))
	}

	return h.next.Handle(ctx, out)
}

// keep applies the sampling policy: every 4xx/5xx and every slow request
// is kept; fast successes on noisy routes are kept 1 in sampleEvery.
func (h *Handler) keep(path string, status int64, dur time.Duration) bool {
	if status >= 400 || dur > slowRequest || !noisy(path) {
		return true
	}

	return h.seen.Add(1)%sampleEvery == 1
}

func noisy(path string) bool {
	for _, l := range []string{"/pt-br/", "/es/"} {
		if rest, ok := strings.CutPrefix(path, l); ok {
			path = "/" + rest
			break
		}
	}

	return path == "/up" || strings.HasPrefix(path, "/build/") ||
		path == "/search/suggest" || path == "/search/events"
}

func redact(a slog.Attr) slog.Attr {
	if sensitiveKeys[strings.ToLower(a.Key)] {
		return slog.String(a.Key, Redacted)
	}

	v := a.Value.Resolve()

	switch v.Kind() { //nolint:exhaustive // other kinds carry no free text
	case slog.KindString:
		return slog.String(a.Key, Scrub(v.String()))
	case slog.KindGroup:
		as := v.Group()
		out := make([]slog.Attr, len(as))

		for i, g := range as {
			out[i] = redact(g)
		}

		return slog.Attr{Key: a.Key, Value: slog.GroupValue(out...)}
	case slog.KindAny:
		// shortcut: only errors are scrubbed among KindAny values; structs and maps pass as is, scrub them if one ever carries PII.
		if err, ok := v.Any().(error); ok {
			return slog.String(a.Key, Scrub(err.Error()))
		}
	}

	return slog.Attr{Key: a.Key, Value: v}
}
