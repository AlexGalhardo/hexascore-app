// Package appapi holds the /app/v1 handlers used by the Android app (A-01, ADR 0009).
package appapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/auth/jwt"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/social"
	"github.com/AlexGalhardo/vukem/web"
	"github.com/AlexGalhardo/vukem/z"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/http/middleware"
	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// Guard is the name of the token guard shared with routes.
const Guard = "api"

// Tokens serves /app/v1/tokens (SDD 03 §6.2, §7.4).
type Tokens struct {
	a *accounts.Accounts
	// google holds one verifier per accepted client id (Android and web); an empty list refuses every ID token.
	google []*jwt.Verifier
	// twoFactor caps second-factor attempts per user on every issue path
	// (password and Google), whatever IP or identifier spelling is used.
	twoFactor ratelimit.Limiter
}

// NewTokens builds the handlers; google verifies Google ID tokens against the accepted audiences.
func NewTokens(a *accounts.Accounts, google []*jwt.Verifier, twoFactor ratelimit.Limiter) *Tokens {
	return &Tokens{a: a, google: google, twoFactor: twoFactor}
}

type passwordInput struct {
	Identifier   string `json:"identifier"`
	Password     string `json:"password"`
	DeviceName   string `json:"deviceName"`
	Code         string `json:"code"`
	RecoveryCode string `json:"recoveryCode"`
	// Challenge is reserved for the bot-protection token (P1.C.5.x); it is accepted and ignored until then.
	Challenge string `json:"challenge"`
}

var passwordSchema = z.Struct[passwordInput](z.Shape{
	"identifier":   z.String().Trim().Min(1).Max(254),
	"password":     z.String().Min(1),
	"deviceName":   z.String().Trim().Min(1).Max(200),
	"code":         z.String().Trim().Max(16).Optional(),
	"recoveryCode": z.String().Trim().Max(32).Optional(),
	"challenge":    z.String().Max(2048).Optional(),
})

type googleInput struct {
	IDToken      string `json:"idToken"`
	DeviceName   string `json:"deviceName"`
	Code         string `json:"code"`
	RecoveryCode string `json:"recoveryCode"`
}

var googleSchema = z.Struct[googleInput](z.Shape{
	"idToken":      z.String().Min(1).Max(8192),
	"deviceName":   z.String().Trim().Min(1).Max(200),
	"code":         z.String().Trim().Max(16).Optional(),
	"recoveryCode": z.String().Trim().Max(32).Optional(),
})

// problem writes an RFC 9457 problem with a kebab-case code (05-9).
func (h *Tokens) problem(c *web.Ctx, status int, code, title string) error {
	raw, err := json.Marshal(map[string]any{
		"type":   strings.TrimSuffix(h.a.ProblemBase, "/") + "/" + code,
		"title":  title,
		"status": status,
		"code":   code,
	})
	if err != nil {
		return err
	}

	return c.Blob(status, "application/problem+json", raw)
}

// Create exchanges a nickname or email and a password for an app token: POST /app/v1/tokens.
func (h *Tokens) Create(c *web.Ctx) error {
	in, err := web.Parse(c, passwordSchema)
	if err != nil {
		return err
	}

	device, ok := accounts.CleanDeviceName(in.DeviceName)
	if !ok {
		return errs.NewValidation().Add("deviceName", "Device name must be 1 to 64 characters.")
	}

	email, err := h.a.EmailForIdentifier(c.Context(), in.Identifier)
	if err != nil {
		return err
	}

	u, err := h.a.Web.Validate(c, map[string]string{"email": email, "password": in.Password})
	if err != nil {
		if _, isValidation := errors.AsType[*errs.ValidationError](err); isValidation {
			return h.problem(c, http.StatusUnprocessableEntity, "invalid-credentials", "Invalid credentials")
		}

		return err
	}

	return h.issue(c, u, device, in.Code, in.RecoveryCode)
}

// Google exchanges a Google ID token for an app token: POST /app/v1/tokens/google.
func (h *Tokens) Google(c *web.Ctx) error {
	in, err := web.Parse(c, googleSchema)
	if err != nil {
		return err
	}

	device, ok := accounts.CleanDeviceName(in.DeviceName)
	if !ok {
		return errs.NewValidation().Add("deviceName", "Device name must be 1 to 64 characters.")
	}

	claims, ok := h.verifyGoogle(c, in.IDToken)
	if !ok {
		return h.problem(c, http.StatusUnauthorized, "invalid-id-token", "Invalid ID token")
	}

	email, _ := claims.Extra["email"].(string)
	if email == "" || !truthy(claims.Extra["email_verified"]) {
		return h.problem(c, http.StatusConflict, "email-unverified-at-provider", "Email not verified at Google")
	}

	name, _ := claims.Extra["name"].(string)

	u, err := h.a.FindOrCreateFromSocial(c.Context(), social.User{
		Provider: "google", ID: claims.Subject, Email: email, EmailVerified: true, Name: name,
	})
	if err != nil {
		return err
	}

	return h.issue(c, u, device, in.Code, in.RecoveryCode)
}

func truthy(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return t == "true"
	}

	return false
}

func (h *Tokens) verifyGoogle(c *web.Ctx, idToken string) (jwt.Claims, bool) {
	for _, v := range h.google {
		claims, err := v.Verify(c.Context(), idToken)
		if err == nil && claims.Subject != "" {
			return claims, true
		}
	}

	web.SecurityEvent(c.Request(), "auth.id_token_rejected", "provider", "google")

	return jwt.Claims{}, false
}

// issue applies the checks that follow a proven identity (ban, verified email, second factor) and mints the token.
func (h *Tokens) issue(c *web.Ctx, u models.User, device, code, recovery string) error {
	if u.IsSeed {
		return h.problem(c, http.StatusUnprocessableEntity, "invalid-credentials", "Invalid credentials")
	}

	if u.IsBanned(h.a.Clock()) {
		return middleware.BannedProblem(c, h.a.ProblemBase, u)
	}

	if !u.EmailVerified() {
		return h.problem(c, http.StatusForbidden, "email-not-verified", "Email address not verified")
	}

	tf, err := h.a.TwoFactor(c.Context(), u.ID)
	if err != nil {
		return err
	}

	if tf.Enabled() {
		if code == "" && recovery == "" {
			return h.problem(c, http.StatusUnprocessableEntity, "two-factor-required", "Two-factor code required")
		}

		res, err := h.twoFactor.Allow(c.Context(), "app.2fa:"+strconv.FormatInt(u.ID, 10))
		if err != nil {
			return err
		}

		if !res.Allowed {
			return h.problem(c, http.StatusTooManyRequests, "too-many-attempts", "Too many attempts")
		}

		valid, err := h.a.VerifyTwoFactor(c.Context(), u.ID, code, recovery)
		if err != nil {
			return err
		}

		if !valid {
			return h.problem(c, http.StatusUnprocessableEntity, "invalid-credentials", "Invalid credentials")
		}
	}

	plain, tok, err := h.a.IssueAppToken(c.Context(), u, device)
	if err != nil {
		return err
	}

	return created(c, plain, tok)
}

func created(c *web.Ctx, plain string, tok auth.Token) error {
	return c.JSON(http.StatusCreated, map[string]any{
		"token": plain, "tokenId": tok.ID, "expiresAt": tok.ExpiresAt, "abilities": tok.Abilities,
	})
}

// Rotate replaces the current token with a new one: POST /app/v1/tokens/rotate.
func (h *Tokens) Rotate(c *web.Ctx) error {
	u, err := auth.User[models.User](c, Guard)
	if err != nil {
		return err
	}

	cur, ok := auth.CurrentToken(c, Guard)
	if !ok {
		return auth.ErrUnauthenticated
	}

	plain, tok, err := h.a.RotateAppToken(c.Context(), u, cur)
	if errors.Is(err, accounts.ErrTokenAlreadyRotated) {
		return h.problem(c, http.StatusUnauthorized, "token-revoked", "Token revoked")
	}

	if err != nil {
		return err
	}

	return created(c, plain, tok)
}

type tokenView struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	LastUsedAt *time.Time `json:"lastUsedAt"`
	ExpiresAt  *time.Time `json:"expiresAt"`
	Current    bool       `json:"current"`
}

// List returns the caller's live tokens: GET /app/v1/tokens.
func (h *Tokens) List(c *web.Ctx) error {
	u, err := auth.User[models.User](c, Guard)
	if err != nil {
		return err
	}

	cur, _ := auth.CurrentToken(c, Guard)

	toks, err := h.a.LiveAppTokens(c.Context(), u)
	if err != nil {
		return err
	}

	out := make([]tokenView, 0, len(toks))
	for _, t := range toks {
		out = append(out, tokenView{
			ID: t.ID, Name: t.Name, LastUsedAt: t.LastUsedAt, ExpiresAt: t.ExpiresAt, Current: t.ID == cur.ID,
		})
	}

	return c.JSON(http.StatusOK, out)
}

// RevokeCurrent revokes the token of this request: DELETE /app/v1/tokens/current.
func (h *Tokens) RevokeCurrent(c *web.Ctx) error {
	cur, ok := auth.CurrentToken(c, Guard)
	if !ok {
		return auth.ErrUnauthenticated
	}

	if err := h.a.Tokens.Revoke(c.Context(), cur.ID); err != nil {
		return err
	}

	return c.NoContent()
}

// Revoke revokes one of the caller's tokens: DELETE /app/v1/tokens/{id}. Other users' and unknown ids are 404.
func (h *Tokens) Revoke(c *web.Ctx) error {
	u, err := auth.User[models.User](c, Guard)
	if err != nil {
		return err
	}

	ok, err := h.a.Tokens.RevokeFor(c.Context(), u.AuthID(), c.Param("id"))
	if err != nil {
		return err
	}

	if !ok {
		return errs.New(http.StatusNotFound, "Token not found.")
	}

	return c.NoContent()
}
