package routes

import (
	"cmp"
	"strconv"

	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/router"

	"github.com/AlexGalhardo/hexascore-app/app/http/limits"
	appmiddleware "github.com/AlexGalhardo/hexascore-app/app/http/middleware"
	"github.com/AlexGalhardo/hexascore-app/app/security"
)

// newChallenger builds the proof-of-work gate shared by the auth routes
// (ADR 0016, SDD 03 section 7.3).
func newChallenger(app *vukem.App, problemBase string) (*appmiddleware.Challenger, error) {
	key, err := app.DeriveKey("pow", 32)
	if err != nil {
		return nil, err
	}

	globalLimit, _ := strconv.Atoi(app.Source.Get("CHALLENGE_GLOBAL_LIMIT"))

	return &appmiddleware.Challenger{
		PoW:         security.New(key, app.Cache, app.Now),
		Cache:       app.Cache,
		ProblemBase: problemBase,
		GlobalLimit: globalLimit,
	}, nil
}

// Challenge registers GET /challenge, which hands out a challenge for a protected route.
func Challenge(r *router.Router, app *vukem.App) error {
	limiters, err := limits.Factory(app)
	if err != nil {
		return err
	}

	ch, err := newChallenger(app, cmp.Or(app.Source.Get("APP_PROBLEM_BASE_URL"), "https://hexascore.app/problems/"))
	if err != nil {
		return err
	}

	r.Get("/challenge", ch.Issue).Name("challenge").
		Use(ratelimit.Middleware(limits.Challenge.From(limiters), ratelimit.ByIP))

	return nil
}
