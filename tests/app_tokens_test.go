package tests

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/auth/jwt"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// signedUp creates a verified user with testPassword through the sign-up flow and returns a fresh client.
func signedUp(t *testing.T, app *vukemtest.App, email string) *vukemtest.Client {
	t.Helper()

	register(t, app, app.Client(t), email)

	return app.Client(t)
}

type issued struct {
	Token     string   `json:"token"`
	TokenID   string   `json:"tokenId"`
	Abilities []string `json:"abilities"`
}

func issue(t *testing.T, c *vukemtest.Client, email, device string) issued {
	t.Helper()

	res := c.PostJSON("/app/v1/tokens", map[string]any{"identifier": email, "password": testPassword, "deviceName": device}).
		AssertCreated()

	var out issued
	res.JSON(&out)

	return out
}

func codeOf(res *vukemtest.Response) string {
	var p struct {
		Code string `json:"code"`
	}

	res.JSON(&p)

	return p.Code
}

func TestAppTokens(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)
	c := signedUp(t, app, "mobile@example.com")

	tok := issue(t, c, "mobile@example.com", "  Pixel 8  ")
	if !strings.HasPrefix(tok.Token, "hx_app_") {
		t.Fatalf("token prefix: %.12s", tok.Token)
	}

	want := accounts.AppAbilities()
	if len(tok.Abilities) != 6 || strings.Join(tok.Abilities, ",") != strings.Join(want, ",") {
		t.Fatalf("abilities = %v", tok.Abilities)
	}

	for _, a := range tok.Abilities {
		if a == "*" {
			t.Fatal("wildcard ability issued")
		}
	}

	list := app.Client(t).WithToken(tok.Token).GetJSON("/app/v1/tokens").AssertOK()
	list.AssertJSONCount("", 1).AssertJSONPath("0.name", "Pixel 8").AssertJSONPath("0.current", true)

	// Rotation: the new token works, the old one is revoked.
	var rotated issued

	app.Client(t).WithToken(tok.Token).PostJSON("/app/v1/tokens/rotate", nil).AssertCreated().JSON(&rotated)
	app.Client(t).WithToken(tok.Token).GetJSON("/app/v1/tokens").AssertUnauthorized()
	app.Client(t).WithToken(rotated.Token).GetJSON("/app/v1/tokens").AssertOK().AssertJSONCount("", 1)

	// Revoking another user's token id is 404; own id and "current" are 204.
	other := signedUp(t, app, "other@example.com")
	otherTok := issue(t, other, "other@example.com", "Galaxy")
	app.Client(t).WithToken(rotated.Token).Delete("/app/v1/tokens/" + otherTok.TokenID).AssertNotFound()
	app.Client(t).WithToken(otherTok.Token).GetJSON("/app/v1/tokens").AssertOK()

	second := issue(t, c, "mobile@example.com", "Tablet")
	app.Client(t).WithToken(rotated.Token).Delete("/app/v1/tokens/" + second.TokenID).AssertNoContent()
	app.Client(t).WithToken(second.Token).GetJSON("/app/v1/tokens").AssertUnauthorized()
	app.Client(t).WithToken(rotated.Token).Delete("/app/v1/tokens/current").AssertNoContent()
	app.Client(t).WithToken(rotated.Token).GetJSON("/app/v1/tokens").AssertUnauthorized()
}

func TestAppTokensCapAtTen(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)
	c := signedUp(t, app, "many@example.com")

	// The issue limit is per IP: issue through the accounts service instead of 11 HTTP calls.
	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	u, err := acc.Users.ByCredentials(t.Context(), map[string]string{"email": "many@example.com"})
	if err != nil {
		t.Fatal(err)
	}

	var first string

	for i := range accounts.MaxAppTokens + 1 {
		plain, _, err := acc.IssueAppToken(t.Context(), u, "device")
		if err != nil {
			t.Fatal(err)
		}

		// Use every token but the first so it is the least recently used (creation timestamps can tie).
		if i == 0 {
			first = plain
		} else if i < accounts.MaxAppTokens {
			c.WithToken(plain).GetJSON("/app/v1/tokens").AssertOK()
		}
	}

	live, err := acc.LiveAppTokens(t.Context(), u)
	if err != nil || len(live) != accounts.MaxAppTokens {
		t.Fatalf("live tokens = %d, %v", len(live), err)
	}

	c.WithToken(first).GetJSON("/app/v1/tokens").AssertUnauthorized()
}

func TestAppTokensRefusals(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)
	signedUp(t, app, "ok@example.com")

	bad := app.Client(t).PostJSON("/app/v1/tokens", map[string]any{
		"identifier": "ok@example.com", "password": "wrong password 1!", "deviceName": "x",
	}).AssertUnprocessable()
	if got := codeOf(bad); got != "invalid-credentials" {
		t.Errorf("wrong password code = %q", got)
	}

	unverified := makeUser(t, app, "unverified@example.com", "user", false)
	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	hash, err := acc.Hasher.Make(testPassword)
	if err != nil {
		t.Fatal(err)
	}

	db := acc.DB
	if _, err := db.ExecContext(
		t.Context(),
		`UPDATE users SET password = $1 WHERE id = $2`,
		hash,
		unverified.ID,
	); err != nil {
		t.Fatal(err)
	}

	res := app.Client(t).PostJSON("/app/v1/tokens", map[string]any{
		"identifier": "unverified@example.com", "password": testPassword, "deviceName": "x",
	}).AssertForbidden()
	if got := codeOf(res); got != "email-not-verified" {
		t.Errorf("unverified code = %q", got)
	}

	// A paid API key shape, an unknown app token and no token are all refused on /app/v1.
	for _, tok := range []string{"hx_live_abc.def", "hx_app_abc.def", "vk_abc.def"} {
		app.Client(t).WithToken(tok).GetJSON("/app/v1/tokens").AssertUnauthorized()
	}
}

func TestAppTokensTwoFactor(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)
	browser := app.Client(t).Inertia("")
	enableTwoFactor(t, app, browser, "tfa-accept@example.com")

	res := app.Client(t).PostJSON("/app/v1/tokens", map[string]any{
		"identifier": "tfa-accept@example.com", "password": testPassword, "deviceName": "Pixel",
	}).AssertUnprocessable()
	if got := codeOf(res); got != "two-factor-required" {
		t.Fatalf("code = %q, want two-factor-required", got)
	}
}

func TestAppTokensChallenge(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)
	api := app.Client(t)
	body := map[string]any{"identifier": "nobody-accept@example.com", "password": "x", "deviceName": "t"}

	for range 5 {
		api.PostJSON("/app/v1/tokens", body).AssertStatus(422)
	}

	res := api.PostJSON("/app/v1/tokens", body).AssertStatus(428)
	if got := codeOf(res); got != "challenge-required" {
		t.Fatalf("code = %q, want challenge-required", got)
	}
}

func TestAppTokensBanned(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)
	signedUp(t, app, "banned-accept@example.com")

	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	u, err := acc.Users.ByCredentials(t.Context(), map[string]string{"email": "banned-accept@example.com"})
	if err != nil {
		t.Fatal(err)
	}

	admin := makeUser(t, app, "admin-accept@example.com", "admin", true)
	until := time.Now().Add(48 * time.Hour).UTC()
	if err := acc.Ban(t.Context(), admin.ID, u.ID, "spam", &until); err != nil {
		t.Fatal(err)
	}

	res := app.Client(t).PostJSON("/app/v1/tokens", map[string]any{
		"identifier": "banned-accept@example.com", "password": testPassword, "deviceName": "x",
	}).AssertForbidden()

	var body struct {
		Code      string `json:"code"`
		Reason    string `json:"reason"`
		ExpiresAt string `json:"expiresAt"`
	}
	res.JSON(&body)
	if body.Code != "account-banned" || body.Reason != "spam" {
		t.Fatalf("problem = %+v", body)
	}

	if _, err := time.Parse(time.RFC3339, body.ExpiresAt); err != nil {
		t.Fatalf("expiresAt = %q: %v", body.ExpiresAt, err)
	}
}

func TestAppTokensRotateLimit(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)
	c := signedUp(t, app, "spin@example.com")
	plain := issue(t, c, "spin@example.com", "Phone").Token

	for range 10 {
		var next issued

		app.Client(t).WithToken(plain).PostJSON("/app/v1/tokens/rotate", nil).AssertCreated().JSON(&next)
		plain = next.Token
	}

	status := app.Client(t).WithToken(plain).PostJSON("/app/v1/tokens/rotate", nil).Status()
	if status != http.StatusTooManyRequests {
		t.Fatalf("11th rotate status = %d, want 429", status)
	}
}

func TestSecurity_A01_TokenAbilities(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)
	u := makeUser(t, app, "abil@example.com", "user", true)
	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	plain, _, err := acc.Tokens.Create(t.Context(), u.AuthID(), "ratings only", []string{"ratings:write"}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	app.Client(t).WithToken(plain).GetJSON("/app/v1/tokens").AssertForbidden()
	app.Client(t).GetJSON("/app/v1/tokens").AssertUnauthorized()
}

func TestSecurity_A07_AppTokenIssueThrottle(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)
	c := app.Client(t)

	var last int

	for range 12 {
		last = c.PostJSON("/app/v1/tokens", map[string]any{
			"identifier": "nobody@example.com", "password": "wrong password 1!", "deviceName": "x",
		}).Status()
	}

	if last != http.StatusTooManyRequests {
		t.Errorf("status after repeated failures = %d, want 429", last)
	}
}

func TestSecurity_A07_AppTokenBannedAndTwoFactor(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)
	browser := app.Client(t).Inertia("")
	secret, codes := enableTwoFactor(t, app, browser, "tfa@example.com")

	body := map[string]any{"identifier": "tfa@example.com", "password": testPassword, "deviceName": "Pixel"}

	res := app.Client(t).PostJSON("/app/v1/tokens", body).AssertUnprocessable()
	if got := codeOf(res); got != "two-factor-required" {
		t.Errorf("code = %q, want two-factor-required", got)
	}

	body["code"] = "000000"
	if got := codeOf(
		app.Client(t).PostJSON("/app/v1/tokens", body).AssertUnprocessable(),
	); got != "invalid-credentials" {
		t.Errorf("wrong code = %q", got)
	}

	body["code"] = totpCode(t, secret, 1) // step 0 was spent confirming the setup
	app.Client(t).PostJSON("/app/v1/tokens", body).AssertCreated()

	delete(body, "code")
	body["recoveryCode"] = codes[0]
	app.Client(t).PostJSON("/app/v1/tokens", body).AssertCreated()
	app.Client(t).PostJSON("/app/v1/tokens", body).AssertUnprocessable() // recovery codes work once
}

func TestSecurity_A07_AppTokenBanned(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)
	signedUp(t, app, "banned@example.com")

	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	u, err := acc.Users.ByCredentials(t.Context(), map[string]string{"email": "banned@example.com"})
	if err != nil {
		t.Fatal(err)
	}

	admin := makeUser(t, app, "admin@example.com", "admin", true)
	if err := acc.Ban(t.Context(), admin.ID, u.ID, "spam", nil); err != nil {
		t.Fatal(err)
	}

	res := app.Client(t).PostJSON("/app/v1/tokens", map[string]any{
		"identifier": "banned@example.com", "password": testPassword, "deviceName": "x",
	}).AssertForbidden()
	if got := codeOf(res); got != "account-banned" {
		t.Errorf("code = %q", got)
	}
}

// googleHarness serves a JWKS for a test key and returns a function that signs ID tokens with it.
func googleHarness(t *testing.T) (env map[string]string, sign func(aud string, extra map[string]any) string) {
	t.Helper()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	key := jwt.RS256("test-key", priv)

	set, err := jwt.PublicJWKS(key)
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(set)
	}))
	t.Cleanup(srv.Close)

	env = map[string]string{
		"GOOGLE_JWKS_URL":          srv.URL,
		"SOCIAL_GOOGLE_CLIENT_ID":  "web-client",
		"GOOGLE_ANDROID_CLIENT_ID": "android-client",
	}

	return env, func(aud string, extra map[string]any) string {
		tok, err := jwt.Sign(key, jwt.Claims{
			Issuer: "https://accounts.google.com", Subject: "g-123", Audience: []string{aud},
			IssuedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour), Extra: extra,
		})
		if err != nil {
			t.Fatal(err)
		}

		return tok
	}
}

func TestSecurity_A07_GoogleIDTokenAudience(t *testing.T) {
	t.Parallel()

	env, sign := googleHarness(t)
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: env})
	c := app.Client(t)

	claims := map[string]any{"email": "g.user@example.com", "email_verified": true, "name": "G User"}
	post := func(idToken string) *vukemtest.Response {
		return c.PostJSON("/app/v1/tokens/google", map[string]any{"idToken": idToken, "deviceName": "Pixel"})
	}

	// A token Google issued to another app is refused even though the signature is valid.
	if got := codeOf(post(sign("someone-elses-client", claims)).AssertUnauthorized()); got != "invalid-id-token" {
		t.Errorf("foreign audience code = %q", got)
	}

	post("not-a-jwt").AssertUnauthorized()

	// Both configured clients are accepted.
	post(sign("android-client", claims)).AssertCreated()
	post(sign("web-client", claims)).AssertCreated()

	unverified := map[string]any{"email": "u@example.com", "email_verified": false}
	if got := codeOf(
		post(sign("android-client", unverified)).AssertStatus(http.StatusConflict),
	); got != "email-unverified-at-provider" {
		t.Errorf("unverified code = %q", got)
	}
}

// TestSecurity_A07_AppTokenRotationSingleUse: the token a rotation consumes
// cannot be rotated again, so two concurrent rotations with a stolen token
// never leave two live successors.
func TestSecurity_A07_AppTokenRotationSingleUse(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)
	signedUp(t, app, "rotate@example.com")

	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	u, err := acc.Users.ByCredentials(t.Context(), map[string]string{"email": "rotate@example.com"})
	if err != nil {
		t.Fatal(err)
	}

	_, cur, err := acc.IssueAppToken(t.Context(), u, "phone")
	if err != nil {
		t.Fatal(err)
	}

	if _, _, err := acc.RotateAppToken(t.Context(), u, cur); err != nil {
		t.Fatal(err)
	}

	if _, _, err := acc.RotateAppToken(t.Context(), u, cur); !errors.Is(err, accounts.ErrTokenAlreadyRotated) {
		t.Fatalf("second rotation err = %v, want ErrTokenAlreadyRotated", err)
	}

	live, err := acc.LiveAppTokens(t.Context(), u)
	if err != nil || len(live) != 1 {
		t.Fatalf("live tokens = %d, %v", len(live), err)
	}
}

// TestSecurity_A07_AppTokenTwoFactorAttemptsPerUser: wrong second factors
// are capped per user, so a correct code after the cap is refused too.
func TestSecurity_A07_AppTokenTwoFactorAttemptsPerUser(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)
	secret, _ := enableTwoFactor(t, app, app.Client(t).Inertia(""), "cap2fa@example.com")

	body := map[string]any{
		"identifier": "cap2fa@example.com", "password": testPassword, "deviceName": "Pixel", "code": "000000",
	}
	for range 5 {
		app.Client(t).PostJSON("/app/v1/tokens", body).AssertUnprocessable()
	}

	body["code"] = totpCode(t, secret, 1)
	// Past the sixth attempt from one IP the route also asks for proof of work.
	c := app.Client(t)
	if got := codeOf(
		c.WithHeader(middlewareHeader, solved(t, c, "token")).
			PostJSON("/app/v1/tokens", body).
			AssertStatus(http.StatusTooManyRequests),
	); got != "too-many-attempts" {
		t.Errorf("code after cap = %q, want too-many-attempts", got)
	}
}

func TestSecurity_A01_AppTokenRotationKeepsAbilities(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)
	signedUp(t, app, "narrow@example.com")

	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	u, err := acc.Users.ByCredentials(t.Context(), map[string]string{"email": "narrow@example.com"})
	if err != nil {
		t.Fatal(err)
	}

	uid := strconv.FormatInt(u.ID, 10)

	_, narrow, err := acc.Tokens.Create(t.Context(), uid, "narrow", []string{"catalog:read", "admin:all"}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	_, next, err := acc.RotateAppToken(t.Context(), u, narrow)
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(next.Abilities, []string{"catalog:read"}) {
		t.Errorf("rotated abilities = %v, want [catalog:read]", next.Abilities)
	}

	_, foreign, err := acc.Tokens.Create(t.Context(), uid, "foreign", []string{"admin:all"}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	if _, _, err := acc.RotateAppToken(t.Context(), u, foreign); !errors.Is(err, accounts.ErrTokenAlreadyRotated) {
		t.Errorf("rotating a non-app token err = %v", err)
	}
}
