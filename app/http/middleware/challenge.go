package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/cache"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/security"
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
	PoW         *security.PoW
	Cache       func() (*cache.Cache, error)
	ProblemBase string
}

func (ch *Challenger) key(rule ChallengeRule, ip string) string {
	return "challenge:" + rule.Name + ":" + ip
}

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

	read := func(key string) (int64, error) {
		if count {
			return c.Increment(ctx, key, 1, rule.Window)
		}

		n, _, err := cache.Get[int64](ctx, c, key)

		return n, err
	}

	n, err := read(ch.key(rule, ip))
	if err != nil {
		return false, 0, err
	}

	bits = security.DefaultBits
	required = n > int64(rule.Limit)

	if rule.GlobalLimit > 0 {
		g, err := read(ch.key(rule, "all"))
		if err != nil {
			return false, 0, err
		}

		if g > int64(rule.GlobalLimit) {
			required, bits = true, security.AttackBits
		}
	}

	return required, bits, nil
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

			switch err := ch.PoW.Verify(r.Context(), rule.Name, c, n, bits); {
			case err == nil:
				next.ServeHTTP(w, r)
			case errors.Is(err, security.ErrStore):
				web.Fail(w, r, errs.Wrap(http.StatusServiceUnavailable, err))
			default:
				if c != "" {
					web.SecurityEvent(r, "challenge.failed", "route", rule.Name)
				}

				ch.problem(w, rule, bits)
			}
		})
	}
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

// Issue serves GET /challenge?route=<name>: a challenge for the route and
// whether this client needs one now. The counters are only read, so asking
// does not count as an attempt.
func (ch *Challenger) Issue(c *web.Ctx) error {
	rule, ok := ChallengeRules[c.Query("route")]
	if !ok {
		return errs.New(http.StatusUnprocessableEntity, "Unknown route.")
	}

	required, bits, err := ch.state(c.Context(), rule, ratelimit.ByIP(c.Request()), false)
	if err != nil {
		return errs.Wrap(http.StatusServiceUnavailable, err)
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
