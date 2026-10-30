package tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/challenge"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

type powChallenge struct {
	Required bool   `json:"required"`
	C        string `json:"c"`
	Bits     int    `json:"bits"`
	Exp      int64  `json:"exp"`
}

func solved(t *testing.T, c *vukemtest.Client, route string) string {
	t.Helper()

	var ch powChallenge
	c.GetJSON("/challenge?route=" + route).AssertOK().JSON(&ch)

	return ch.C + ":" + challenge.Solve(challenge.Challenge{C: ch.C, Bits: ch.Bits, Exp: ch.Exp})
}

func signupAs(app *vukemtest.App, c *vukemtest.Client, n int) *vukemtest.Response {
	return c.PostJSON("/signup", signUpForm(app, "Bot Net", fmt.Sprintf("bot%d@example.com", n), testPassword))
}

func TestSecurity_A07_PowSignup(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)
	c := app.Client(t).Inertia("")

	var first powChallenge
	c.GetJSON("/challenge?route=signup").AssertOK().JSON(&first)
	if first.Required {
		t.Fatal("a quiet client must not be asked for a challenge")
	}

	for i := range 3 {
		signupAs(app, c, i).AssertRedirect("/signup/check-email")
	}

	// The probe announces the attempt that will cross the limit, so the client solves before posting.
	var announced powChallenge
	c.GetJSON("/challenge?route=signup").AssertOK().JSON(&announced)
	if !announced.Required {
		t.Fatal("/challenge must require a solution for the attempt past the limit")
	}

	// Past the threshold: no solution, 428 with a fresh challenge.
	var problem struct {
		Code      string `json:"code"`
		Challenge struct {
			C    string `json:"c"`
			Bits int    `json:"bits"`
		} `json:"challenge"`
	}
	signupAs(app, c, 3).AssertStatus(428).JSON(&problem)
	if problem.Code != "challenge-required" || problem.Challenge.C == "" ||
		problem.Challenge.Bits != challenge.DefaultBits {
		t.Fatalf("problem = %+v", problem)
	}

	var again powChallenge
	c.GetJSON("/challenge?route=signup").AssertOK().JSON(&again)
	if !again.Required {
		t.Fatal("/challenge must report the requirement")
	}

	proof := solved(t, c, "signup")
	signupAs(app, c.WithHeader(middlewareHeader, proof), 4).AssertRedirect("/signup/check-email")

	// The same solution a second time is refused.
	signupAs(app, c.WithHeader(middlewareHeader, proof), 5).AssertStatus(428)

	// A solution minted for another route does not count.
	signupAs(app, c.WithHeader(middlewareHeader, solved(t, c, "login")), 6).AssertStatus(428)
	// Garbage does not either.
	signupAs(app, c.WithHeader(middlewareHeader, "x:1"), 7).AssertStatus(428)

	app.Client(t).GetJSON("/challenge?route=nope").AssertStatus(422)
}

func TestSecurity_A07_PowTokens(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)
	api := app.Client(t)
	body := map[string]any{"identifier": "nobody@example.com", "password": "x", "deviceName": "t"}
	for range 5 {
		api.PostJSON("/app/v1/tokens", body).AssertStatus(422)
	}

	api.PostJSON("/app/v1/tokens", body).AssertStatus(428)
	api.WithHeader(middlewareHeader, solved(t, api, "token")).PostJSON("/app/v1/tokens", body).AssertStatus(422)
}

const middlewareHeader = "X-Challenge"

// CHALLENGE_GLOBAL_LIMIT replaces the signup global threshold. The e2e stack lifts it;
// a low value proves the override is what the rule reads.
func TestSecurity_A07_PowGlobalLimitOverride(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{"CHALLENGE_GLOBAL_LIMIT": "2"}})
	c := app.Client(t).Inertia("")

	for i := range 2 {
		signupAs(app, c, i).AssertRedirect("/signup/check-email")
	}

	var ch powChallenge
	c.GetJSON("/challenge?route=signup").AssertOK().JSON(&ch)
	if !ch.Required || ch.Bits != challenge.AttackBits {
		t.Fatalf("challenge = %+v, want attack difficulty", ch)
	}
}

func TestSecurity_A07_PowGlobalLimitRefusedAtBoot(t *testing.T) {
	t.Parallel()

	_, err := vukemtest.Build(t, bootstrap.New, vukemtest.Options{
		Env: map[string]string{"CHALLENGE_GLOBAL_LIMIT": "nope"},
	})
	if err == nil || !strings.Contains(err.Error(), "CHALLENGE_GLOBAL_LIMIT") {
		t.Fatalf("boot err = %v, want a refusal", err)
	}
}
