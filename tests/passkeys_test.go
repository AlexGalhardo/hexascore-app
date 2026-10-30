package tests

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"maps"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/auth/webauthn/webauthntest"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func passkeyApp(t *testing.T, extra map[string]string) *vukemtest.App {
	t.Helper()

	env := map[string]string{"APP_URL": "http://localhost", "APP_NAME": "HexaScore"}
	maps.Copy(env, extra)

	return vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: env})
}

func addPasskey(t *testing.T, c *vukemtest.Client, authn *webauthntest.Authenticator, name string) {
	t.Helper()

	res := c.PostJSON("/user/passkeys/options", map[string]any{"name": name, "password": testPassword}).AssertOK()
	cred, err := authn.Create(res.Recorder().Body.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	c.Do(http.MethodPost, "/user/passkeys", bytes.NewReader(cred), "application/json").AssertRedirect("/profile")
}

func passkeyAssertion(t *testing.T, c *vukemtest.Client, authn *webauthntest.Authenticator) []byte {
	t.Helper()

	res := c.PostJSON("/login/passkey/options", map[string]any{}).AssertOK()
	if !bytes.Contains(res.Recorder().Body.Bytes(), []byte(`"allowCredentials":[]`)) {
		t.Fatal("sign-in options must advertise an empty allow list")
	}

	body, err := authn.Get(res.Recorder().Body.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	return body
}

func postAssertion(c *vukemtest.Client, body []byte) *vukemtest.Response {
	return c.Do(http.MethodPost, "/login/passkey", bytes.NewReader(body), "application/json")
}

func stillGuest(t *testing.T, c *vukemtest.Client) {
	t.Helper()

	c.Get("/dashboard").AssertRedirect("/login")
}

func passkeyID(t *testing.T, app *vukemtest.App) int64 {
	t.Helper()

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	var id int64
	err = db.QueryRowContext(t.Context(), `SELECT id FROM passkeys ORDER BY id DESC LIMIT 1`).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}

	return id
}

func passkeyRows(t *testing.T, app *vukemtest.App) int {
	t.Helper()

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	var n int
	if err := db.QueryRowContext(t.Context(), `SELECT count(*) FROM passkeys`).Scan(&n); err != nil {
		t.Fatal(err)
	}

	return n
}

func credentialMaterial(t *testing.T, app *vukemtest.App) (id, public []byte) {
	t.Helper()

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	if err := db.QueryRowContext(t.Context(),
		`SELECT credential_id, public_key FROM passkeys ORDER BY id DESC LIMIT 1`).Scan(&id, &public); err != nil {
		t.Fatal(err)
	}

	return id, public
}

func passkeyMails(app *vukemtest.App, email string) int {
	n := 0
	for _, m := range app.Mails().SentTo(email) {
		if m.Subject == "Your passkey was changed" {
			n++
		}
	}

	return n
}

func eventCredential(t *testing.T, app *vukemtest.App, kind string) string {
	t.Helper()

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	var data string
	if err := db.QueryRowContext(t.Context(),
		`SELECT data FROM account_events WHERE kind = ? ORDER BY id DESC LIMIT 1`, kind).Scan(&data); err != nil {
		t.Fatal(err)
	}

	var payload map[string]string
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		t.Fatal(err)
	}

	if len(payload) != 1 || len(payload["credential"]) != 64 {
		t.Fatalf("event %s data keys = %d", kind, len(payload))
	}

	return payload["credential"]
}

func ageLoginChallenge(t *testing.T, app *vukemtest.App) {
	t.Helper()

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	pending, err := agedLoginSessions(t, db)
	if err != nil {
		t.Fatal(err)
	}

	if len(pending) == 0 {
		t.Fatal("no session stored passkey.login.since")
	}

	for _, row := range pending {
		if _, err := db.ExecContext(
			t.Context(), `UPDATE sessions SET data = ? WHERE id = ?`, row.data, row.id); err != nil {
			t.Fatal(err)
		}
	}
}

type agedSession struct {
	id, data string
}

func agedLoginSessions(t *testing.T, db *database.DB) ([]agedSession, error) {
	t.Helper()

	rows, err := db.QueryContext(t.Context(), `SELECT id, data FROM sessions`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	old := strconv.FormatInt(time.Now().Add(-10*time.Minute).Unix(), 10)
	var pending []agedSession

	for rows.Next() {
		var id, raw string
		if err := rows.Scan(&id, &raw); err != nil {
			return nil, err
		}

		if !strings.Contains(raw, "passkey.login.since") {
			continue
		}

		dec := json.NewDecoder(strings.NewReader(raw))
		dec.UseNumber()

		var data map[string]any
		if err := dec.Decode(&data); err != nil {
			return nil, err
		}

		data["passkey.login.since"] = old

		next, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}

		pending = append(pending, agedSession{id: id, data: string(next)})
	}

	return pending, rows.Err()
}

func assertNoMaterial(t *testing.T, body string, blobs ...[]byte) {
	t.Helper()

	for _, b := range blobs {
		if len(b) == 0 {
			continue
		}

		for _, v := range []string{
			base64.RawURLEncoding.EncodeToString(b),
			base64.StdEncoding.EncodeToString(b),
			hex.EncodeToString(b),
		} {
			if v != "" && strings.Contains(body, v) {
				t.Fatal("profile response contains key material")
			}
		}
	}
}

func TestSecurity_A07_PasskeyForeignCredential(t *testing.T) {
	t.Parallel()

	app := passkeyApp(t, nil)
	c := app.Client(t)
	register(t, app, c, "foreign@example.com")
	authn := webauthntest.New("http://localhost")

	res := c.PostJSON("/user/passkeys/options", map[string]any{"password": testPassword}).AssertOK()
	if _, err := authn.Create(res.Recorder().Body.Bytes()); err != nil {
		t.Fatal(err)
	}

	c.PostJSON("/logout", nil).AssertRedirect("/")
	postAssertion(c, passkeyAssertion(t, c, authn)).AssertUnauthorized()
	stillGuest(t, c)
}

func TestSecurity_A07_PasskeyReplayedChallenge(t *testing.T) {
	t.Parallel()

	app := passkeyApp(t, nil)
	c := app.Client(t)
	register(t, app, c, "replay@example.com")
	authn := webauthntest.New("http://localhost")
	addPasskey(t, c, authn, "Desk")
	c.PostJSON("/logout", nil).AssertRedirect("/")

	body := passkeyAssertion(t, c, authn)
	postAssertion(c, body).AssertRedirect("/onboarding")
	c.PostJSON("/logout", nil).AssertRedirect("/")
	postAssertion(c, body).AssertUnauthorized()
	stillGuest(t, c)
}

func TestSecurity_A07_PasskeyExpiredChallenge(t *testing.T) {
	t.Parallel()

	app := passkeyApp(t, map[string]string{"SESSION_DRIVER": "database"})
	c := app.Client(t)
	register(t, app, c, "expired@example.com")
	authn := webauthntest.New("http://localhost")
	addPasskey(t, c, authn, "")
	c.PostJSON("/logout", nil).AssertRedirect("/")

	res := c.PostJSON("/login/passkey/options", map[string]any{}).AssertOK()
	body, err := authn.Get(res.Recorder().Body.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	ageLoginChallenge(t, app)
	postAssertion(c, body).AssertUnauthorized()
	stillGuest(t, c)
}

func TestSecurity_A07_PasskeyCounterRegression(t *testing.T) {
	t.Parallel()

	app := passkeyApp(t, nil)
	c := app.Client(t)
	register(t, app, c, "counter@example.com")
	authn := webauthntest.New("http://localhost")
	addPasskey(t, c, authn, "")

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(t.Context(), `UPDATE passkeys SET sign_count = 100`); err != nil {
		t.Fatal(err)
	}

	c.PostJSON("/logout", nil).AssertRedirect("/")
	postAssertion(c, passkeyAssertion(t, c, authn)).AssertUnauthorized()
	stillGuest(t, c)
}

func TestSecurity_A01_PasskeyRemoveOnlyOwn(t *testing.T) {
	t.Parallel()

	app := passkeyApp(t, nil)
	owner := app.Client(t)
	register(t, app, owner, "owner@example.com")
	addPasskey(t, owner, webauthntest.New("http://localhost"), "Desk")
	id := passkeyID(t, app)

	other := app.Client(t)
	register(t, app, other, "other@example.com")
	other.Delete("/profile/passkeys/" + strconv.FormatInt(id, 10)).AssertNotFound()
	if passkeyRows(t, app) != 1 {
		t.Fatal("another user removed the passkey")
	}

	owner.Delete("/profile/passkeys/nope").AssertNotFound()
	owner.Delete("/profile/passkeys/" + strconv.FormatInt(id, 10)).AssertRedirect("/profile")
	if passkeyRows(t, app) != 0 {
		t.Fatal("owner delete left the row")
	}
}

func TestPasskeyRegisterLimit(t *testing.T) {
	t.Parallel()

	app := passkeyApp(t, nil)
	c := app.Client(t)
	register(t, app, c, "limit@example.com")
	authn := webauthntest.New("http://localhost")

	for i := range 10 {
		addPasskey(t, c, authn, "k"+strconv.Itoa(i))
	}

	c.PostJSON("/user/passkeys/options", map[string]any{"password": testPassword}).
		AssertStatus(http.StatusConflict).
		AssertJSONPath("code", "passkey-limit")
}

func TestPasskeyLoginSkipsTOTP(t *testing.T) {
	t.Parallel()

	app := passkeyApp(t, nil)
	owner := app.Client(t).Inertia("")
	enableTwoFactor(t, app, owner, "totp@example.com")
	authn := webauthntest.New("http://localhost")
	addPasskey(t, owner, authn, "")
	owner.PostJSON("/logout", nil).AssertRedirect("/")

	owner.PostJSON("/login", map[string]any{"identifier": "totp@example.com", "password": testPassword}).
		AssertRedirect("/two-factor-challenge")

	fresh := app.Client(t)
	postAssertion(fresh, passkeyAssertion(t, fresh, authn)).AssertRedirect("/onboarding")
	fresh.Get("/onboarding").AssertOK()
}

func TestPasskeyChangeAuditedAndMailed(t *testing.T) {
	t.Parallel()

	const email = "audit@example.com"

	app := passkeyApp(t, nil)
	c := app.Client(t)
	register(t, app, c, email)
	addPasskey(t, c, webauthntest.New("http://localhost"), "Desk")

	if passkeyMails(app, email) != 1 {
		t.Fatalf("mails after add = %d, want 1", passkeyMails(app, email))
	}

	id, _ := credentialMaterial(t, app)
	sum := sha256.Sum256(id)
	want := hex.EncodeToString(sum[:])
	if got := eventCredential(t, app, "passkey.added"); got != want {
		t.Fatal("passkey.added hash does not match the credential id")
	}

	c.Delete("/profile/passkeys/" + strconv.FormatInt(passkeyID(t, app), 10)).AssertRedirect("/profile")
	if passkeyMails(app, email) != 2 {
		t.Fatalf("mails after remove = %d, want 2", passkeyMails(app, email))
	}

	if got := eventCredential(t, app, "passkey.removed"); got != want {
		t.Fatal("passkey.removed hash does not match the credential id")
	}
}

func TestPasskeyProfileProps(t *testing.T) {
	t.Parallel()

	app := passkeyApp(t, nil)
	c := app.Client(t).Inertia("")
	register(t, app, c, "props@example.com")
	addPasskey(t, c, webauthntest.New("http://localhost"), "Desk")

	res := c.Get("/profile").AssertOK()
	res.AssertInertia("Profile/Index").
		PropEquals("passkeys.0.name", "Desk").
		HasProp("passkeys.0.id").
		HasProp("passkeys.0.createdAt").
		PropEquals("passkeys.0.lastUsedAt", nil).
		MissingProp("publicKey").
		MissingProp("credentialId").
		MissingProp("public_key").
		MissingProp("passkeys.0.publicKey").
		MissingProp("passkeys.0.credentialId").
		MissingProp("passkeys.0.public_key")

	credID, public := credentialMaterial(t, app)
	assertNoMaterial(t, res.Recorder().Body.String(), credID, public)
}
