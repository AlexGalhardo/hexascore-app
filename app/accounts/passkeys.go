package accounts

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/auth/webauthn"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/orm"
	"github.com/AlexGalhardo/vukem/social"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

const (
	passkeyMax         = 10
	passkeyNameLimit   = 64
	passkeyCeremonyTTL = 5 * time.Minute
	passkeyDefaultName = "Passkey"

	sessRegChallenge   = "passkey.reg.challenge"
	sessRegSince       = "passkey.reg.since"
	sessRegHandle      = "passkey.reg.handle"
	sessRegName        = "passkey.reg.name"
	sessLoginChallenge = "passkey.login.challenge"
	sessLoginSince     = "passkey.login.since"
	sessGoogleConfirm  = "passkey.google_confirm"
	sessGoogleAt       = "passkey.google_at"
)

var (
	errPasskeyLimit    = errors.New("passkey limit")
	errPasskeyExists   = errors.New("passkey exists")
	errPasskeyNotFound = errors.New("passkey not found")
)

// PasskeyPublic is one passkey as the profile page may show it.
// Credential ids and public keys stay in the database.
type PasskeyPublic struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	CreatedAt  time.Time  `json:"createdAt"`
	LastUsedAt *time.Time `json:"lastUsedAt"`
}

// PasskeyBannedError is a verified assertion for an account that is banned.
// The controller turns it into the account-banned problem.
type PasskeyBannedError struct {
	User models.User
}

// Error implements error.
func (e *PasskeyBannedError) Error() string { return "passkey login: account banned" }

// MarkGoogleConfirm starts a signed-in Google round trip used to add a passkey.
func (*Accounts) MarkGoogleConfirm(s web.Session) {
	s.Put(sessGoogleConfirm, "1")
}

// NoteFreshGoogle records a Google sign-in as a passkey confirmation for five minutes.
func (a *Accounts) NoteFreshGoogle(s web.Session) {
	s.Put(sessGoogleAt, strconv.FormatInt(a.now().Unix(), 10))
}

// FinishGoogleConfirm completes a passkey confirmation when the session asked for one.
// handled is false when this callback is an ordinary sign-in.
func (a *Accounts) FinishGoogleConfirm(c *web.Ctx, su social.User) (bool, error) {
	flag, _ := c.Session().Get(sessGoogleConfirm).(string)
	if flag != "1" {
		return false, nil
	}

	c.Session().Forget(sessGoogleConfirm)

	u, err := auth.User[models.User](c, "web")
	if err != nil {
		return true, err
	}

	if su.Provider != "google" {
		return true, a.passkeyNeedGoogle(c.Context())
	}

	owns, err := a.ownsGoogle(c.Context(), u.ID, su.ID)
	if err != nil {
		return true, err
	}

	if !owns {
		return true, a.passkeyNeedGoogle(c.Context())
	}

	a.NoteFreshGoogle(c.Session())

	return true, c.Redirect("/profile")
}

// PasskeyList returns the caller's passkeys without key material.
func (a *Accounts) PasskeyList(ctx context.Context, userID int64) ([]PasskeyPublic, error) {
	p := a.DB.Dialect().Placeholder
	rows, err := a.DB.QueryContext(ctx,
		"SELECT id, name, created_at, last_used_at FROM passkeys WHERE user_id = "+p(1)+" ORDER BY id", userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := []PasskeyPublic{}
	for rows.Next() {
		var (
			row  PasskeyPublic
			last sql.NullTime
		)
		if err := rows.Scan(&row.ID, &row.Name, &row.CreatedAt, &last); err != nil {
			return nil, err
		}

		if last.Valid {
			tm := last.Time
			row.LastUsedAt = &tm
		}

		out = append(out, row)
	}

	return out, rows.Err()
}

// BeginPasskeyRegistration stores a single-use challenge and returns creation options.
func (a *Accounts) BeginPasskeyRegistration(
	ctx context.Context, s web.Session, u models.User, name, password string,
) (webauthn.CreationOptions, error) {
	if err := a.confirmPasskey(ctx, s, u.ID, password); err != nil {
		return webauthn.CreationOptions{}, err
	}

	name = passkeyName(name)

	n, err := a.passkeyCount(ctx, a.DB, u.ID)
	if err != nil {
		return webauthn.CreationOptions{}, err
	}

	if n >= passkeyMax {
		return webauthn.CreationOptions{}, a.passkeyLimit(ctx)
	}

	rp, err := a.relyingParty()
	if err != nil {
		return webauthn.CreationOptions{}, err
	}

	handle, err := a.passkeyHandle(ctx, u.ID)
	if err != nil {
		return webauthn.CreationOptions{}, err
	}

	exclude, err := a.passkeyExclude(ctx, u.ID)
	if err != nil {
		return webauthn.CreationOptions{}, err
	}

	challenge := webauthn.NewChallenge()
	stamp := strconv.FormatInt(a.now().Unix(), 10)
	s.Put(sessRegChallenge, challenge)
	s.Put(sessRegSince, stamp)
	s.Put(sessRegHandle, base64.RawURLEncoding.EncodeToString(handle))
	s.Put(sessRegName, name)

	display := u.Name
	if display == "" {
		display = u.Email
	}

	return rp.RegistrationOptions(
		webauthn.User{Handle: handle, Name: u.Email, DisplayName: display}, challenge, exclude), nil
}

// FinishPasskeyRegistration verifies the browser credential and stores it.
func (a *Accounts) FinishPasskeyRegistration(ctx context.Context, s web.Session, u models.User, body []byte) error {
	challenge, handle, name, ok := a.takeRegistration(s)
	if !ok {
		return a.passkeyNotSaved(ctx)
	}

	rp, err := a.relyingParty()
	if err != nil {
		return err
	}

	cred, err := rp.VerifyRegistration(
		challenge, webauthn.User{Handle: handle, Name: u.Email, DisplayName: u.Name}, body)
	if err != nil {
		a.log.InfoContext(ctx, "passkey ceremony rejected", "op", "register")

		return a.passkeyNotSaved(ctx)
	}

	err = a.DB.Tx(ctx, func(tx *database.Tx) error {
		p := tx.Dialect().Placeholder
		if _, err := tx.ExecContext(ctx,
			"UPDATE users SET updated_at = "+p(1)+" WHERE id = "+p(2), a.now().UTC(), u.ID); err != nil {
			return err
		}

		n, err := a.passkeyCount(ctx, tx, u.ID)
		if err != nil {
			return err
		}

		if n >= passkeyMax {
			return errPasskeyLimit
		}

		return a.insertPasskey(ctx, tx, u.ID, name, cred)
	})
	if errors.Is(err, errPasskeyLimit) {
		return a.passkeyLimit(ctx)
	}

	if errors.Is(err, errPasskeyExists) {
		return a.passkeyExists(ctx)
	}

	if err != nil {
		return err
	}

	a.sendPasskeyChanged(ctx, u.ID, "added")

	return nil
}

// BeginPasskeyLogin stores a single-use discoverable challenge.
func (a *Accounts) BeginPasskeyLogin(_ context.Context, s web.Session) (webauthn.RequestOptions, error) {
	rp, err := a.relyingParty()
	if err != nil {
		return webauthn.RequestOptions{}, err
	}

	challenge := webauthn.NewChallenge()
	s.Put(sessLoginChallenge, challenge)
	s.Put(sessLoginSince, strconv.FormatInt(a.now().Unix(), 10))

	return rp.AssertionOptions(challenge, nil), nil
}

// FinishPasskeyLogin verifies an assertion, updates the sign count, and returns the user.
// A banned account comes back as *PasskeyBannedError. Anonymised accounts get the same 401 as a bad passkey.
func (a *Accounts) FinishPasskeyLogin(ctx context.Context, s web.Session, body []byte) (models.User, error) {
	assertion, err := webauthn.ParseAssertion(body)
	if err != nil {
		return models.User{}, a.passkeyFailed(ctx)
	}

	challenge, ok := a.takeLogin(s)
	if !ok {
		return models.User{}, a.passkeyFailed(ctx)
	}

	row, err := a.passkeyByID(ctx, assertion.CredentialID)
	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, a.passkeyFailed(ctx)
	}

	if err != nil {
		return models.User{}, err
	}

	cred := row.credential()
	storedHandle := cred.UserHandle
	if len(storedHandle) == 0 {
		cred.UserHandle = assertion.UserHandle
	}

	rp, err := a.relyingParty()
	if err != nil {
		return models.User{}, err
	}

	updated, err := rp.VerifyAssertion(challenge, cred, assertion)
	if err != nil {
		a.log.InfoContext(ctx, "passkey ceremony rejected", "op", "login")

		return models.User{}, a.passkeyFailed(ctx)
	}

	if err := a.saveAssertion(ctx, row.id, storedHandle, assertion.UserHandle, updated); err != nil {
		return models.User{}, err
	}

	u, err := orm.QueryOf[models.User](a.DB).Where("id", row.userID).First(ctx)
	if errors.Is(err, orm.ErrNotFound) || u.AnonymizedAt.Valid {
		return models.User{}, a.passkeyFailed(ctx)
	}

	if err != nil {
		return models.User{}, err
	}

	if u.IsBanned(a.Clock()) {
		return models.User{}, &PasskeyBannedError{User: u}
	}

	return u, nil
}

// RemovePasskey deletes one of the caller's passkeys. Another user's id is not found.
func (a *Accounts) RemovePasskey(ctx context.Context, userID, id int64) error {
	err := a.DB.Tx(ctx, func(tx *database.Tx) error {
		p := tx.Dialect().Placeholder
		var credID []byte
		err := tx.QueryRowContext(ctx,
			"DELETE FROM passkeys WHERE id = "+p(1)+" AND user_id = "+p(2)+" RETURNING credential_id",
			id, userID).Scan(&credID)
		if errors.Is(err, sql.ErrNoRows) {
			return errPasskeyNotFound
		}

		if err != nil {
			return err
		}

		return recordEvent(ctx, tx, userID, nil, "passkey.removed", passkeyEventData(credID), a.now().UTC())
	})
	if errors.Is(err, errPasskeyNotFound) {
		return a.passkeyNotFound(ctx)
	}

	if err != nil {
		return err
	}

	a.sendPasskeyChanged(ctx, userID, "removed")

	return nil
}

func (a *Accounts) relyingParty() (webauthn.RelyingParty, error) {
	u, err := url.Parse(a.URL)
	if err != nil {
		return webauthn.RelyingParty{}, err
	}

	// The RP ID is the public hostname, so production stays hexascore.app and imported credentials still verify.
	name := a.Name
	if name == "" {
		name = "HexaScore"
	}

	rp := webauthn.RelyingParty{ID: u.Hostname(), Name: name, Origins: []string{u.Scheme + "://" + u.Host}}

	return rp, rp.Validate()
}

func (a *Accounts) confirmPasskey(ctx context.Context, s web.Session, userID int64, password string) error {
	var (
		stored string
		set    sql.NullTime
	)

	err := a.DB.QueryRowContext(ctx,
		"SELECT password, password_set_at FROM users WHERE id = "+a.DB.Dialect().Placeholder(1), userID).
		Scan(&stored, &set)
	if err != nil {
		return err
	}

	if set.Valid {
		ok, err := a.Hasher.Verify(stored, password)
		if err != nil {
			return err
		}

		if !ok {
			return errs.NewValidation().Add("password", a.T(ctx, "auth.two_factor.wrong_password"))
		}

		return nil
	}

	if a.passkeyGoogleFresh(s) {
		return nil
	}

	return a.passkeyNeedGoogle(ctx)
}

func (a *Accounts) passkeyGoogleFresh(s web.Session) bool {
	raw, _ := s.Get(sessGoogleAt).(string)

	return a.ceremonyFresh(raw)
}

func (a *Accounts) ownsGoogle(ctx context.Context, userID int64, providerID string) (bool, error) {
	p := a.DB.Dialect().Placeholder
	var one int
	err := a.DB.QueryRowContext(ctx,
		"SELECT 1 FROM social_accounts WHERE user_id = "+p(1)+" AND provider = 'google' AND provider_id = "+p(2),
		userID, providerID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	return err == nil, err
}

func (a *Accounts) passkeyCount(ctx context.Context, ex database.Executor, userID int64) (int, error) {
	var n int
	err := ex.QueryRowContext(ctx,
		"SELECT count(*) FROM passkeys WHERE user_id = "+ex.Dialect().Placeholder(1), userID).Scan(&n)

	return n, err
}

func (a *Accounts) passkeyHandle(ctx context.Context, userID int64) ([]byte, error) {
	var handle []byte
	err := a.DB.QueryRowContext(ctx,
		"SELECT user_handle FROM passkeys WHERE user_id = "+a.DB.Dialect().Placeholder(1)+
			" AND user_handle IS NOT NULL LIMIT 1", userID).Scan(&handle)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && len(handle) == 0) {
		return webauthn.NewUserHandle(), nil
	}

	return handle, err
}

// PasskeyMissing is the 404 for an id that is not one of the caller's passkeys.
func (a *Accounts) PasskeyMissing(ctx context.Context) error {
	return a.passkeyNotFound(ctx)
}

func (a *Accounts) passkeyExclude(ctx context.Context, userID int64) ([]webauthn.Credential, error) {
	rows, err := a.DB.QueryContext(ctx,
		"SELECT credential_id, transports FROM passkeys WHERE user_id = "+a.DB.Dialect().Placeholder(1), userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []webauthn.Credential
	for rows.Next() {
		var (
			id   []byte
			raw  string
			list []string
		)
		if err := rows.Scan(&id, &raw); err != nil {
			return nil, err
		}

		if raw != "" {
			if err := json.Unmarshal([]byte(raw), &list); err != nil {
				return nil, err
			}
		}

		out = append(out, webauthn.Credential{ID: id, Transports: list})
	}

	return out, rows.Err()
}

func (a *Accounts) insertPasskey(
	ctx context.Context, tx *database.Tx, userID int64, name string, cred webauthn.Credential,
) error {
	raw, err := transportsJSON(cred.Transports)
	if err != nil {
		return err
	}

	p := tx.Dialect().Placeholder
	_, err = tx.ExecContext(ctx, `INSERT INTO passkeys (
  user_id, credential_id, public_key, algorithm, sign_count, transports, aaguid, user_handle,
  backup_eligible, backup_state, name, created_at
) VALUES (`+p(1)+","+p(2)+","+p(3)+","+p(4)+","+p(5)+","+p(6)+","+
		p(7)+","+p(8)+","+p(9)+","+p(10)+","+p(11)+","+p(12)+")",
		userID, cred.ID, cred.PublicKey, cred.Algorithm, int64(cred.SignCount), raw, nil, cred.UserHandle,
		cred.BackupEligible, cred.BackupState, name, a.now().UTC())
	if isUnique(err) {
		return errPasskeyExists
	}

	if err != nil {
		return err
	}

	return recordEvent(ctx, tx, userID, nil, "passkey.added", passkeyEventData(cred.ID), a.now().UTC())
}

type storedPasskey struct {
	id                          int64
	userID                      int64
	credentialID, publicKey     []byte
	userHandle                  []byte
	algorithm                   int64
	signCount                   int64
	backupEligible, backupState bool
	transports                  []string
}

func (s storedPasskey) credential() webauthn.Credential {
	count := uint32(0)
	switch {
	case s.signCount > int64(math.MaxUint32):
		count = math.MaxUint32
	case s.signCount > 0:
		count = uint32(s.signCount)
	}

	return webauthn.Credential{
		ID: s.credentialID, PublicKey: s.publicKey, Algorithm: s.algorithm, SignCount: count,
		UserHandle: s.userHandle, BackupEligible: s.backupEligible, BackupState: s.backupState,
		Transports: s.transports,
	}
}

func (a *Accounts) passkeyByID(ctx context.Context, credentialID []byte) (storedPasskey, error) {
	p := a.DB.Dialect().Placeholder
	var (
		row storedPasskey
		raw string
	)
	err := a.DB.QueryRowContext(ctx, `SELECT id, user_id, credential_id, public_key, algorithm, sign_count,
  user_handle, backup_eligible, backup_state, transports FROM passkeys WHERE credential_id = `+p(1), credentialID).
		Scan(&row.id, &row.userID, &row.credentialID, &row.publicKey, &row.algorithm, &row.signCount,
			&row.userHandle, &row.backupEligible, &row.backupState, &raw)
	if err != nil {
		return storedPasskey{}, err
	}

	if raw != "" {
		if err := json.Unmarshal([]byte(raw), &row.transports); err != nil {
			return storedPasskey{}, err
		}
	}

	return row, nil
}

func (a *Accounts) saveAssertion(
	ctx context.Context, id int64, stored, asserted []byte, cred webauthn.Credential,
) error {
	var pin any
	if len(stored) == 0 && len(asserted) > 0 {
		pin = asserted
	}

	p := a.DB.Dialect().Placeholder
	_, err := a.DB.ExecContext(ctx, `UPDATE passkeys SET sign_count = `+p(1)+`, backup_state = `+p(2)+
		`, last_used_at = `+p(3)+`, user_handle = CASE WHEN user_handle IS NULL THEN `+p(4)+
		` ELSE user_handle END WHERE id = `+p(5),
		int64(cred.SignCount), cred.BackupState, a.now().UTC(), pin, id)

	return err
}

func (a *Accounts) takeRegistration(s web.Session) (string, []byte, string, bool) {
	challenge, _ := s.Get(sessRegChallenge).(string)
	since, _ := s.Get(sessRegSince).(string)
	handleB64, _ := s.Get(sessRegHandle).(string)
	name, _ := s.Get(sessRegName).(string)
	s.Forget(sessRegChallenge, sessRegSince, sessRegHandle, sessRegName)

	handle, err := base64.RawURLEncoding.DecodeString(handleB64)
	if err != nil || challenge == "" || name == "" || len(handle) == 0 || !a.ceremonyFresh(since) {
		return "", nil, "", false
	}

	return challenge, handle, name, true
}

func (a *Accounts) takeLogin(s web.Session) (string, bool) {
	challenge, _ := s.Get(sessLoginChallenge).(string)
	since, _ := s.Get(sessLoginSince).(string)
	s.Forget(sessLoginChallenge, sessLoginSince)

	return challenge, challenge != "" && a.ceremonyFresh(since)
}

func (a *Accounts) ceremonyFresh(since string) bool {
	sec, err := strconv.ParseInt(since, 10, 64)
	if err != nil {
		return false
	}

	at := time.Unix(sec, 0)
	now := a.now()

	return !at.After(now.Add(time.Minute)) && now.Sub(at) <= passkeyCeremonyTTL
}

func passkeyName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return passkeyDefaultName
	}

	if utf8.RuneCountInString(name) > passkeyNameLimit {
		return string([]rune(name)[:passkeyNameLimit])
	}

	return name
}

func passkeyEventData(credentialID []byte) map[string]any {
	sum := sha256.Sum256(credentialID)

	return map[string]any{"credential": hex.EncodeToString(sum[:])}
}

func transportsJSON(list []string) (string, error) {
	if list == nil {
		list = []string{}
	}

	raw, err := json.Marshal(list)

	return string(raw), err
}

func isUnique(err error) bool {
	return err != nil && strings.Contains(strings.ToUpper(err.Error()), "UNIQUE")
}

func (a *Accounts) passkeyProblem(ctx context.Context, status int, code, messageKey string) error {
	return errs.NewKey(status, code, a.T(ctx, messageKey), nil).WithExtension("code", code)
}

func (a *Accounts) passkeyFailed(ctx context.Context) error {
	return a.passkeyProblem(ctx, http.StatusUnauthorized, "passkey-failed", "auth.passkey.failed")
}

func (a *Accounts) passkeyNotSaved(ctx context.Context) error {
	return a.passkeyProblem(ctx, http.StatusUnprocessableEntity, "passkey-failed", "auth.passkey.not_saved")
}

func (a *Accounts) passkeyNotFound(ctx context.Context) error {
	return a.passkeyProblem(ctx, http.StatusNotFound, "passkey-not-found", "auth.passkey.not_found")
}

func (a *Accounts) passkeyLimit(ctx context.Context) error {
	return a.passkeyProblem(ctx, http.StatusConflict, "passkey-limit", "auth.passkey.limit")
}

func (a *Accounts) passkeyExists(ctx context.Context) error {
	return a.passkeyProblem(ctx, http.StatusConflict, "passkey-exists", "auth.passkey.exists")
}

func (a *Accounts) passkeyNeedGoogle(ctx context.Context) error {
	return a.passkeyProblem(ctx, http.StatusForbidden, "passkey-confirm", "auth.passkey.need_google")
}
