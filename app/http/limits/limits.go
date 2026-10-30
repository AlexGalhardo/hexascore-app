// Package limits is the rate-limit table of SDD 13 §7.1: one named rule per row, so every route group
// takes its budget from here and no route builds an ad-hoc limiter. Each rule has its own name, because
// limiters with the same name share buckets.
package limits

import (
	"errors"
	"fmt"
	"time"

	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/ratelimit"
)

// Rule is one row of the table: Limit requests per Window for each key.
type Rule struct {
	Name   string
	Limit  int
	Window time.Duration
}

// From builds the rule's limiter on the app's store.
func (r Rule) From(f ratelimit.Factory) ratelimit.Limiter { return f(r.Name, r.Limit, r.Window) }

// The table. Keys (IP, user, account, email hash) are chosen where each rule is installed; the names
// predate this table and are kept so stored budgets carry over.
var (
	Login        = Rule{"login", 5, time.Minute}                // identifier + IP
	LoginIP      = Rule{"login.ip", 10, time.Minute}            // IP across identifiers
	LoginAccount = Rule{"login.account", 20, 10 * time.Minute}  // account
	Signup       = Rule{"register", 10, 10 * time.Minute}       // IP
	SignupEmail  = Rule{"register.email", 5, time.Hour}         // email hash
	Forgot       = Rule{"password.email", 10, 10 * time.Minute} // IP
	ForgotEmail  = Rule{"password.email.address", 5, time.Hour} // email hash
	Mails        = Rule{"mails", 3, time.Minute}                // IP: verification resend
	TwoFactor    = Rule{"two-factor.challenge", 5, time.Minute}
	TwoFactorSet = Rule{"two-factor.confirm", 5, time.Minute}
	Deletion     = Rule{"deletion", 5, time.Hour}
	Contact      = Rule{"contact", 5, time.Hour}
	Ratings      = Rule{"rating", 30, time.Minute}
	Status       = Rule{"status", 60, time.Minute}
	TimeToBeat   = Rule{"time-to-beat", 30, time.Minute}
	Lists        = Rule{"lists", 60, time.Minute}
	Editions     = Rule{"edition", 60, time.Minute}
	Notify       = Rule{"notifications", 120, time.Minute}
	Reports      = Rule{"reports", 10, time.Hour}
	Search       = Rule{"search", 60, time.Minute}
	Suggest      = Rule{"search.suggest", 300, time.Minute}
	SearchEvents = Rule{"search.events", 30, time.Minute}
	APIIP        = Rule{"api-ip", 120, time.Minute}
	AppTokens    = Rule{"app.tokens", 10, time.Minute}
	AppTokensAcc = Rule{"app.tokens.account", 20, 10 * time.Minute}
	AppTokensUsr = Rule{"app.tokens.user", 60, time.Minute}
	App2FA       = Rule{"app.2fa", 5, 15 * time.Minute}
	Challenge    = Rule{"challenge", 30, time.Minute}
	Exports      = Rule{"exports", 10, time.Hour}
	CSPReports   = Rule{"csp-reports", 30, time.Minute}
)

// APIPlan is the per-key rate of a public API plan (SDD 05): rps requests per second.
func APIPlan(plan string, rps int) Rule { return Rule{"api-" + plan, rps, time.Second} }

// All lists the fixed rules, for the table test.
func All() []Rule {
	return []Rule{
		Login, LoginIP, LoginAccount, Signup, SignupEmail, Forgot, ForgotEmail, Mails, TwoFactor, TwoFactorSet,
		Deletion, Contact, Ratings, Status, TimeToBeat, Lists, Editions, Notify, Reports, Search, Suggest,
		SearchEvents, APIIP, AppTokens, AppTokensAcc, AppTokensUsr, App2FA, Challenge, Exports, CSPReports,
	}
}

// ErrMemoryStore refuses a production boot whose budgets would reset on every restart and differ per
// instance (SDD 13 §7.1, ADR 0016).
var ErrMemoryStore = errors.New(
	"limits: production needs RATE_LIMIT_STORE=cache on a shared cache (CACHE_STORE=database)",
)

// Factory is the app's limiter factory; in production it must be the shared cache store. vukem's limiters
// already deny with 503 when that store fails (https://vukem.dev/docs/security/rate-limiting "Failure
// behaviour").
func Factory(app *vukem.App) (ratelimit.Factory, error) {
	if app.Config.IsProduction() && app.Source.Get("RATE_LIMIT_STORE") != "cache" {
		return nil, ErrMemoryStore
	}

	f, err := app.RateLimiters()
	if err != nil {
		return nil, fmt.Errorf("limits: %w", err)
	}

	return f, nil
}
