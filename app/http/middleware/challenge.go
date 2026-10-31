package middleware

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/cache"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/i18n"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/challenge"
)

// HeaderChallenge carries "<c>:<n>": the challenge and the nonce that solves it.
const HeaderChallenge = "X-Challenge"

// ChallengeRule is when a route asks for proof of work (SDD 03 section 7.3):
// the request that takes one IP past Limit within Window, or the route as a
// whole past GlobalLimit, must carry a solution. A global overload also
// raises the difficulty.
type ChallengeRule struct {
	Name        string
	Limit       int
	Window      time.Duration
	GlobalLimit int
}

// ChallengeRules are the protected routes, by the name /challenge takes.
// Login counts attempts, not failures: the handler's failure is a redirect
// the middleware cannot tell from success, and a challenge only costs CPU.
var ChallengeRules = map[string]ChallengeRule{
	"signup": {Name: "signup", Limit: 3, Window: time.Hour, GlobalLimit: 30},
	"login":  {Name: "login", Limit: 5, Window: 10 * time.Minute},
	"reset":  {Name: "reset", Limit: 3, Window: time.Hour},
	"token":  {Name: "token", Limit: 5, Window: 10 * time.Minute},
}

// Challenger enforces ChallengeRules.
type Challenger struct {
	PoW         *challenge.PoW
	Cache       func() (*cache.Cache, error)
	ProblemBase string
	// GlobalLimit, when positive, replaces the global threshold of rules that have one.
	// Production leaves it zero. The e2e stack sets CHALLENGE_GLOBAL_LIMIT so a parallel
	// suite does not trip the 21-bit attack difficulty.
	GlobalLimit int
}

// GlobalLimitOverride parses CHALLENGE_GLOBAL_LIMIT. Empty keeps the production default.
// Any other value that is not a positive integer is refused so a typo cannot boot.
func GlobalLimitOverride(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}

	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		return 0, fmt.Errorf("CHALLENGE_GLOBAL_LIMIT must be a positive integer, got %q", raw)
	}

	return n, nil
}

func (ch *Challenger) key(rule ChallengeRule, ip string) string {
	return "challenge:" + rule.Name + ":" + ip
}

// TripHold is how long a tripped flag (one IP or the whole route) keeps the challenge on after the
// counter window ends (SDD 13 §7.2).
const TripHold = time.Hour

// FallbackDrop is how many bits a slow client may drop, once per TripHold per IP, by asking
// /challenge again with fallback=1 after 10 s (SDD 13 §7.2).
const FallbackDrop = 2

// state reads the counters (and adds one when count is set) and reports
// whether the caller needs a challenge and how hard.
func (ch *Challenger) state(
	ctx context.Context,
	rule ChallengeRule,
	ip string,
	count bool,
) (required bool, bits int, err error) {
	c, err := ch.Cache()
	if err != nil {
		return false, 0, err
	}

	required, err = tripped(ctx, c, ch.key(rule, ip), rule.Limit, rule.Window, count)
	if err != nil {
		return false, 0, err
	}

	bits = challenge.DefaultBits

	globalLimit := rule.GlobalLimit
	if globalLimit > 0 && ch.GlobalLimit > 0 {
		globalLimit = ch.GlobalLimit
	}

	if globalLimit > 0 {
		global, err := tripped(ctx, c, ch.key(rule, "all"), globalLimit, rule.Window, count)
		if err != nil {
			return false, 0, err
		}

		if global {
			required, bits = true, challenge.AttackBits
		}
	}

	return required, bits, nil
}

// tripped counts one attempt on key (or only peeks when count is false) and reports whether the
// attempts passed limit within window or a flag set by an earlier trip still holds.
func tripped(
	ctx context.Context,
	c *cache.Cache,
	key string,
	limit int,
	window time.Duration,
	count bool,
) (bool, error) {
	held, _, err := cache.Get[bool](ctx, c, key+":trip")
	if err != nil || held {
		return held, err
	}

	// A read-only probe reports the answer for the attempt it announces, so it counts that attempt
	// too; otherwise the client is told "not required" and the post that crosses the limit gets a 428.
	var n int64
	if count {
		n, err = c.Increment(ctx, key, 1, window)
	} else {
		n, _, err = cache.Get[int64](ctx, c, key)
		n++
	}

	if err != nil || n <= int64(limit) || !count {
		return err == nil && n > int64(limit), err
	}

	return true, c.Put(ctx, key+":trip", true, TripHold)
}

// Require returns the middleware for rule: below the thresholds requests pass;
// past them a valid, unused solution is needed or the answer is 428 with a
// fresh challenge. A broken counter or solution store answers 503 (fail closed).
func (ch *Challenger) Require(name string) func(http.Handler) http.Handler {
	rule := ChallengeRules[name]

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := ratelimit.ByIP(r)

			required, bits, err := ch.state(r.Context(), rule, ip, true)
			// A broken counter store must not take sign-up and login down with it
			// (NFR-03-9): the rate limits still apply. A broken solution store
			// below does fail closed.
			if err != nil {
				web.SecurityEvent(r, "challenge.counter_failed", "route", rule.Name)
			}

			if err != nil || !required {
				next.ServeHTTP(w, r)

				return
			}

			c, n, _ := strings.Cut(r.Header.Get(HeaderChallenge), ":")

			// Accepts a fallback challenge; Issue hands those out at most once per hour per IP.
			switch err := ch.PoW.Verify(r.Context(), rule.Name, c, n, bits-FallbackDrop); {
			case err == nil:
				next.ServeHTTP(w, r)
			case errors.Is(err, challenge.ErrStore):
				web.Fail(w, r, errs.Wrap(http.StatusServiceUnavailable, err))
			default:
				if c != "" {
					web.SecurityEvent(r, "challenge.failed", "route", rule.Name)
				}

				if isPlainFormPost(r) {
					challengePage(w, r)

					return
				}

				ch.problem(w, rule, bits)
			}
		})
	}
}

// isPlainFormPost reports a browser's own form submission: urlencoded or multipart body, HTML
// accepted, no Inertia visit. Only such a client cannot read the JSON problem or solve the challenge.
func isPlainFormPost(r *http.Request) bool {
	ct := r.Header.Get("Content-Type")
	form := strings.HasPrefix(ct, "application/x-www-form-urlencoded") || strings.HasPrefix(ct, "multipart/form-data")

	return form && r.Header.Get("X-Inertia") == "" && strings.Contains(r.Header.Get("Accept"), "text/html")
}

// challengePage is the 428 for a no-JS form post: it stays fail closed and points to Google
// sign-in, which needs no proof of work (SDD 13 section 7.2). No script, no inline style (CSP).
func challengePage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	lang := html.EscapeString(cmp.Or(i18n.Locale(ctx), "en"))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusPreconditionRequired)

	_, _ = fmt.Fprintf(w,
		`<!doctype html><html lang="%s"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>%s</title></head><body><main><h1>%s</h1><p>%s <a href="/auth/google/redirect">%s</a></p></main></body></html>`,
		lang,
		html.EscapeString(i18n.T(ctx, "auth.challenge.checking")),
		html.EscapeString(i18n.T(ctx, "auth.challenge.checking")),
		html.EscapeString(i18n.T(ctx, "auth.challenge.noscript")),
		html.EscapeString(i18n.T(ctx, "auth.google")),
	)
}

// problem writes 428 challenge-required with a fresh challenge to solve.
func (ch *Challenger) problem(w http.ResponseWriter, rule ChallengeRule, bits int) {
	fresh, err := ch.PoW.Issue(rule.Name, bits)
	if err != nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)

		return
	}

	body, _ := json.Marshal(map[string]any{
		"type":      strings.TrimSuffix(ch.ProblemBase, "/") + "/challenge-required",
		"title":     "Challenge required",
		"status":    http.StatusPreconditionRequired,
		"code":      "challenge-required",
		"challenge": fresh,
	})

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(http.StatusPreconditionRequired)
	_, _ = w.Write(body)
}

// Issue serves GET /challenge?route=<name>[&fallback=1]: a challenge for the route and
// whether this client needs one now; fallback asks for FallbackDrop fewer bits, granted once per hour. The counters are only read, so asking
// does not count as an attempt.
func (ch *Challenger) Issue(c *web.Ctx) error {
	rule, ok := ChallengeRules[c.Query("route")]
	if !ok {
		return errs.New(http.StatusUnprocessableEntity, "Unknown route.")
	}

	ip := ratelimit.ByIP(c.Request())

	required, bits, err := ch.state(c.Context(), rule, ip, false)
	if err != nil {
		return errs.Wrap(http.StatusServiceUnavailable, err)
	}

	if required && c.Query("fallback") == "1" {
		store, err := ch.Cache()
		if err != nil {
			return errs.Wrap(http.StatusServiceUnavailable, err)
		}

		first, err := cache.Add(c.Context(), store, "challenge:fallback:"+ip, true, TripHold)
		if err != nil {
			return errs.Wrap(http.StatusServiceUnavailable, err)
		}

		if first {
			bits -= FallbackDrop
		}
	}

	fresh, err := ch.PoW.Issue(rule.Name, bits)
	if err != nil {
		return err
	}

	c.Writer().Header().Set("Cache-Control", "no-store")

	return c.JSON(http.StatusOK, map[string]any{
		"required": required, "c": fresh.C, "bits": fresh.Bits, "exp": fresh.Exp,
	})
}
