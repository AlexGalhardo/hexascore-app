package routes

import (
	"cmp"

	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/auth/jwt"
	"github.com/AlexGalhardo/vukem/middleware"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/router"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/controllers/appapi"
	"github.com/AlexGalhardo/hexascore-app/app/http/limits"
)

const (
	googleIssuer  = "https://accounts.google.com"
	googleJWKSURL = "https://www.googleapis.com/oauth2/v3/certs"
)

// googleVerifiers builds one verifier per accepted client id (SDD 03 §7.4): the Android client and the web
// client. Without ids no ID token is accepted. GOOGLE_JWKS_URL exists for tests and defaults to Google's keys.
func googleVerifiers(get func(string) string) []*jwt.Verifier {
	keys := &jwt.RemoteJWKS{URL: cmp.Or(get("GOOGLE_JWKS_URL"), googleJWKSURL)}

	var out []*jwt.Verifier

	for _, name := range []string{"GOOGLE_ANDROID_CLIENT_ID", "SOCIAL_GOOGLE_CLIENT_ID"} {
		if id := get(name); id != "" {
			out = append(out, jwt.NewVerifier(keys, jwt.VerifyOptions{
				Issuer: googleIssuer, Audience: id, Algorithms: []string{jwt.AlgRS256},
			}))
		}
	}

	return out
}

// AppAPI registers /app/v1/tokens (SDD 03 §6.2). Rating routes join the same prefix in Ratings.
func AppAPI(r *router.Router, app *vukem.App) error {
	acc, err := accounts.New(app)
	if err != nil {
		return err
	}

	limiters, err := limits.Factory(app)
	if err != nil {
		return err
	}

	pow, err := newChallenger(app, acc.ProblemBase)
	if err != nil {
		return err
	}

	h := appapi.NewTokens(acc, googleVerifiers(app.Source.Get), limits.App2FA.From(limiters))
	issueLimit := ratelimit.Middleware(limits.AppTokens.From(limiters), ratelimit.ByIP)
	// Per identifier across IPs: someone holding a password cannot rotate
	// IPs to brute-force the 6-digit TOTP (security review, P1.C.4.1).
	_, byIdentifier := auth.ThrottleKeys("identifier")
	accountLimit := ratelimit.Middleware(limits.AppTokensAcc.From(limiters), byIdentifier)
	userLimit := ratelimit.Middleware(limits.AppTokensUsr.From(limiters), ratelimit.ByUser(acc.API.Name()))

	bearer, err := AppWriteStack(app, acc, limiters, userLimit)
	if err != nil {
		return err
	}

	r.Group("/app/v1", func(g *router.Group) {
		g.Use(middleware.JSONOnly())

		g.Post("/tokens", h.Create).Name("tokens.create").Use(issueLimit, accountLimit, pow.Require("token"))
		g.Post("/tokens/google", h.Google).Name("tokens.google").Use(issueLimit, pow.Require("token"))
		g.Post("/tokens/rotate", h.Rotate).Name("tokens.rotate").Use(bearer...)
		g.Get("/tokens", h.List).Name("tokens.index").Use(append(bearer, auth.Ability(acc.API, "catalog:read"))...)
		g.Delete("/tokens/current", h.RevokeCurrent).Name("tokens.revoke_current").Use(bearer...)
		g.Delete("/tokens/{id}", h.Revoke).Name("tokens.revoke").Use(bearer...)
	}).Name("app.v1.")

	return nil
}
