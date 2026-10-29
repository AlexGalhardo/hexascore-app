// Package accounts wires authentication: guards, tokens, password resets
// and email verification. It is your code (vukem install:auth): change the
// rules freely.
package accounts

import (
	"cmp"
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/auth/hash"
	"github.com/AlexGalhardo/vukem/cache"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/events"
	"github.com/AlexGalhardo/vukem/i18n"
	"github.com/AlexGalhardo/vukem/mail"
	"github.com/AlexGalhardo/vukem/session"
	"github.com/AlexGalhardo/vukem/signed"

	"github.com/AlexGalhardo/hexascore-app/app/billing"
	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// Accounts holds the authentication services.
type Accounts struct {
	DB     *database.DB
	Mailer *mail.Mailer
	Users  *auth.ORMProvider[models.User]
	Web    *auth.SessionGuard[models.User]
	API    *auth.TokenGuard[models.User]
	Tokens *auth.Tokens
	Hasher hash.Hasher
	Signer *signed.Signer
	// Events receives UserBanned and UserUnbanned after commit (nil: none).
	Events *events.Bus
	// ProblemBase prefixes problem type URLs (APP_PROBLEM_BASE_URL).
	ProblemBase string
	// Subscriptions cancels live paid subscriptions before a purge (P1.H.3.1 replaces the default).
	Subscriptions SubscriptionCanceller
	// Messages holds the auth messages of lang/ (see T).
	Messages *i18n.Bundle
	// URL is the public base URL used in emails (APP_URL).
	URL string
	// ResetTTL is how long password reset links work.
	ResetTTL time.Duration
	// VerifyTTL is how long email verification links work.
	VerifyTTL time.Duration
	now       func() time.Time
	sessions  func() (*session.Manager, error)
	cache     func() (*cache.Cache, error)
	log       *slog.Logger
}

// RevokeOtherSessions signs userID out of every browser session except
// keep (the current session id, "" for all). Call it whenever the
// credentials change: password reset or change, two-factor enabled. It
// bumps the user's session version, which ends sessions on every
// SESSION_DRIVER (encrypted cookies included), and deletes server-side
// sessions and remember-me tokens (tokens issued before the bump no
// longer sign in anyway). With keep set, also call a.Web.RefreshSession(c)
// so the current session survives the bump and its device gets a new
// remember-me token.
func (a *Accounts) RevokeOtherSessions(ctx context.Context, userID int64, keep string) error {
	if err := a.bumpSessionVersion(ctx, a.DB, userID); err != nil {
		return err
	}

	if _, err := a.DB.ExecContext(ctx, "DELETE FROM remember_tokens WHERE user_id = "+a.DB.Dialect().Placeholder(1),
		strconv.FormatInt(userID, 10)); err != nil {
		return err
	}

	return a.revokeServerSessions(ctx, userID, keep)
}

// bumpSessionVersion ends every session of userID on the next request.
func (a *Accounts) bumpSessionVersion(ctx context.Context, ex database.Executor, userID int64) error {
	_, err := ex.ExecContext(ctx,
		"UPDATE users SET session_version = session_version + 1 WHERE id = "+ex.Dialect().Placeholder(1), userID)

	return err
}

// revokeServerSessions deletes the server-side sessions of userID except
// keep, so they end at once and leave the store. Encrypted cookie
// sessions have no record to delete: the version bump ends them.
func (a *Accounts) revokeServerSessions(ctx context.Context, userID int64, keep string) error {
	m, err := a.sessions()
	if err != nil {
		return err
	}

	err = m.RevokeOthers(ctx, a.Web.SessionUserID(strconv.FormatInt(userID, 10)), keep)
	if errors.Is(err, session.ErrRevokeUnsupported) {
		return nil
	}

	return err
}

// New builds Accounts from the app.
func New(app *vukem.App) (*Accounts, error) {
	db, err := app.DB()
	if err != nil {
		return nil, err
	}

	mailer, err := app.Mailer()
	if err != nil {
		return nil, err
	}

	inner := hash.Default()
	if app.Kind == vukem.KindTest {
		inner = hash.Fake() // fast hashing keeps tests quick
	}

	// Better Auth scrypt hashes verify and are rehashed to argon2id on login (ADR 0008).
	hasher := NewCompatHasher(inner)

	messages, err := loadMessages()
	if err != nil {
		return nil, err
	}

	users := auth.NewORMProvider[models.User](db)
	tokens := auth.NewTokens(db)
	// A recognizable prefix lets secret scanners catch leaked tokens; make
	// it specific to this app (SDD 03 §6.2).
	tokens.Prefix = "hx_app_"
	tokens.Now = app.Now // a fake app clock also drives token expiry
	tokens.Log = app.Log

	signingKey, err := app.DeriveKey("signed.urls", 32)
	if err != nil {
		return nil, err
	}

	web := auth.NewSessionGuard("web", users, hasher, auth.NewRememberStore(db))
	web.IdentifierField = "identifier"
	web.InsecureCookies = !app.Config.IsProduction() && !strings.HasPrefix(app.Config.URL, "https://")

	canceller, err := billing.NewCanceller(app, db)
	if err != nil {
		return nil, err
	}

	return &Accounts{
		Subscriptions: canceller,
		DB:            db, Mailer: mailer, Users: users, Web: web, Tokens: tokens, Hasher: hasher,
		API: auth.NewTokenGuard("api", users, tokens),
		// Links signed before an APP_KEY rotation verify until the old key leaves APP_PREVIOUS_KEYS.
		Signer: signed.New(signingKey).
			WithPreviousKeys(app.DerivePreviousKeys("signed.urls", 32)...).
			WithClock(app.Now),
		Events:      app.Events,
		ProblemBase: cmp.Or(app.Source.Get("APP_PROBLEM_BASE_URL"), "https://hexascore.app/problems/"),
		Messages:    messages,
		URL:         strings.TrimSuffix(app.Config.URL, "/"),
		ResetTTL:    time.Hour,
		VerifyTTL:   time.Hour,
		now:         time.Now,
		sessions:    app.Sessions,
		cache:       app.Cache,
		log:         app.Log,
	}, nil
}
