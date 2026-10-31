//go:build integration

package legacyimport_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/auth/webauthn"
	"github.com/AlexGalhardo/vukem/auth/webauthn/webauthntest"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
	"github.com/AlexGalhardo/hexascore-app/database/legacyimport"
)

func TestPasskeysStage(t *testing.T) {
	t.Parallel()

	const origin = "http://localhost:3000"

	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{
		"DATABASE_URL": pgtest.Database(t, ""),
		"APP_URL":      origin,
		"APP_NAME":     "HexaScore",
	}})
	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	sql, err := os.ReadFile("testdata/fixture.sql")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(t.Context(), string(sql)); err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	rp := webauthn.RelyingParty{ID: "localhost", Name: "HexaScore", Origins: []string{origin}}
	if err := rp.Validate(); err != nil {
		t.Fatal(err)
	}

	user := webauthn.User{Handle: webauthn.NewUserHandle(), Name: "ana@example.test", DisplayName: "Ana"}
	challenge := webauthn.NewChallenge()
	raw, err := json.Marshal(rp.RegistrationOptions(user, challenge, nil))
	if err != nil {
		t.Fatal(err)
	}

	authn := webauthntest.New(origin)
	body, err := authn.Create(raw)
	if err != nil {
		t.Fatal(err)
	}

	cred, err := rp.VerifyRegistration(challenge, user, body)
	if err != nil {
		t.Fatal(err)
	}

	const aaguid = "6a3b1f8e-9c24-4d0a-8e77-1b2c3d4e5f60"

	created := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	if _, err := db.ExecContext(t.Context(), `INSERT INTO legacy_import.passkey
  (id, user_id, name, public_key, credential_id, counter, device_type, backed_up, transports, created_at, aaguid)
VALUES ('p01k', 'u01', 'Yubikey', $1, $2, 0, 'singleDevice', false, 'internal', $3, $4)`,
		base64.StdEncoding.EncodeToString(cred.PublicKey),
		base64.RawURLEncoding.EncodeToString(cred.ID),
		created, aaguid); err != nil {
		t.Fatal(err)
	}

	stages, err := legacyimport.Select(legacyimport.Stages(), []string{"users"})
	if err != nil {
		t.Fatal(err)
	}

	r := &legacyimport.Runner{DB: db, Batch: 3, Out: io.Discard, Now: time.Now}
	if err := r.Run(t.Context(), stages); err != nil {
		t.Fatalf("run: %v", err)
	}

	c := app.Client(t)
	optRes := c.PostJSON("/login/passkey/options", map[string]any{}).AssertOK()
	assertion, err := authn.Get(optRes.Recorder().Body.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	c.Do(http.MethodPost, "/login/passkey", bytes.NewReader(assertion), "application/json").AssertRedirect("/dashboard")

	var (
		gotAAGUID  string
		gotCreated time.Time
		signCount  int64
		handle     []byte
		alg        int64
	)
	if err := db.QueryRowContext(t.Context(), `SELECT aaguid, created_at, sign_count, user_handle, algorithm
FROM passkeys`).Scan(&gotAAGUID, &gotCreated, &signCount, &handle, &alg); err != nil {
		t.Fatal(err)
	}

	if gotAAGUID != aaguid || !gotCreated.UTC().Equal(created) || signCount != 1 || len(handle) == 0 || alg != -7 {
		t.Fatalf("imported passkey aaguid=%s created=%s sign=%d handleLen=%d alg=%d",
			gotAAGUID, gotCreated.UTC().Format(time.RFC3339), signCount, len(handle), alg)
	}
}
