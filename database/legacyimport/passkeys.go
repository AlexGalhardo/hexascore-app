package legacyimport

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/AlexGalhardo/vukem/database"
)

// importPasskeys copies Better Auth passkey rows for users in (after, last]
// onto passkeys. Incomplete rows (no credential id or public key) are skipped.
// user_handle stays NULL: Better Auth 1.7.7 does not store it.
func importPasskeys(ctx context.Context, tx *database.Tx, after, last string, at time.Time) error {
	batch, err := readLegacyPasskeys(ctx, tx, after, last)
	if err != nil {
		return err
	}

	for i := range batch {
		if err := batch[i].save(ctx, tx, at); err != nil {
			return err
		}
	}

	return nil
}

func readLegacyPasskeys(ctx context.Context, tx *database.Tx, after, last string) ([]legacyPasskey, error) {
	rows, err := tx.QueryContext(ctx, `SELECT id, user_id, name, public_key, credential_id, counter,
  device_type, backed_up, transports, created_at, aaguid
FROM legacy_import.passkey WHERE user_id > $1 AND user_id <= $2 ORDER BY id`, after, last)
	if err != nil {
		return nil, fmt.Errorf("read passkeys: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var batch []legacyPasskey
	for rows.Next() {
		var row legacyPasskey
		if err := rows.Scan(&row.id, &row.userID, &row.name, &row.publicKey, &row.credentialID, &row.counter,
			&row.deviceType, &row.backedUp, &row.transports, &row.createdAt, &row.aaguid); err != nil {
			return nil, err
		}

		batch = append(batch, row)
	}

	return batch, rows.Err()
}

type legacyPasskey struct {
	id, userID             string
	name                   sql.NullString
	publicKey              sql.NullString
	credentialID           sql.NullString
	counter                sql.NullInt64
	deviceType, transports sql.NullString
	aaguid                 sql.NullString
	backedUp               sql.NullBool
	createdAt              sql.NullTime
}

func (row *legacyPasskey) save(ctx context.Context, tx *database.Tx, at time.Time) error {
	if !row.publicKey.Valid || strings.TrimSpace(row.publicKey.String) == "" ||
		!row.credentialID.Valid || strings.TrimSpace(row.credentialID.String) == "" {
		return nil
	}

	credID, err := decodeLegacyKey(row.credentialID.String)
	if err != nil {
		return fmt.Errorf("passkey %s: credential id: %w", row.id, err)
	}

	public, err := decodeLegacyKey(row.publicKey.String)
	if err != nil {
		return fmt.Errorf("passkey %s: public key: %w", row.id, err)
	}

	alg, err := coseAlg(public)
	if err != nil {
		return fmt.Errorf("passkey %s: cose alg: %w", row.id, err)
	}

	transports, err := legacyTransports(row.transports)
	if err != nil {
		return fmt.Errorf("passkey %s: transports: %w", row.id, err)
	}

	created := at.UTC()
	if row.createdAt.Valid {
		created = row.createdAt.Time.UTC()
	}

	var aaguid any
	if row.aaguid.Valid && strings.TrimSpace(row.aaguid.String) != "" {
		aaguid = strings.TrimSpace(row.aaguid.String)
	}

	counter := int64(0)
	if row.counter.Valid {
		counter = row.counter.Int64
	}

	res, err := tx.ExecContext(ctx, `INSERT INTO passkeys (
  user_id, credential_id, public_key, algorithm, sign_count, transports, aaguid,
  backup_eligible, backup_state, name, created_at)
SELECT u.id, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
FROM users u WHERE u.legacy_id = $11
ON CONFLICT (credential_id) DO UPDATE SET
  public_key = EXCLUDED.public_key,
  algorithm = EXCLUDED.algorithm,
  sign_count = CASE WHEN passkeys.sign_count > EXCLUDED.sign_count
    THEN passkeys.sign_count ELSE EXCLUDED.sign_count END,
  transports = EXCLUDED.transports,
  aaguid = EXCLUDED.aaguid,
  backup_eligible = EXCLUDED.backup_eligible,
  backup_state = EXCLUDED.backup_state,
  name = EXCLUDED.name`,
		credID, public, alg, counter, transports, aaguid,
		row.deviceType.Valid && row.deviceType.String == "multiDevice",
		row.backedUp.Valid && row.backedUp.Bool,
		legacyPasskeyName(row.name), created, row.userID)
	if err != nil {
		return fmt.Errorf("passkey %s: %w", row.id, err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n == 0 {
		return fmt.Errorf("passkey %s: owner %s not imported", row.id, row.userID)
	}

	return nil
}

func legacyPasskeyName(name sql.NullString) string {
	if !name.Valid {
		return "Passkey"
	}

	n := strings.TrimSpace(name.String)
	if n == "" {
		return "Passkey"
	}

	if utf8.RuneCountInString(n) > 64 {
		return string([]rune(n)[:64])
	}

	return n
}

func legacyTransports(raw sql.NullString) (string, error) {
	if !raw.Valid || strings.TrimSpace(raw.String) == "" {
		return "[]", nil
	}

	text := strings.TrimSpace(raw.String)
	var list []string
	if strings.HasPrefix(text, "[") {
		if err := json.Unmarshal([]byte(text), &list); err != nil {
			return "", err
		}
	} else {
		for part := range strings.SplitSeq(text, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				list = append(list, part)
			}
		}
	}

	if list == nil {
		list = []string{}
	}

	b, err := json.Marshal(list)

	return string(b), err
}

// decodeLegacyKey accepts Better Auth's standard base64 public keys and base64url credential ids.
func decodeLegacyKey(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if strings.ContainsAny(s, "-_") {
		return base64.RawURLEncoding.DecodeString(strings.TrimRight(s, "="))
	}

	b, err := base64.StdEncoding.DecodeString(s)
	if err == nil {
		return b, nil
	}

	b, rawErr := base64.RawStdEncoding.DecodeString(s)
	if rawErr != nil {
		return nil, err
	}

	return b, nil
}

var errCOSE = errors.New("cose key")

// coseAlg reads COSE key map integer 3. Indefinite lengths and trailing bytes fail closed.
func coseAlg(raw []byte) (int64, error) {
	major, n, rest, err := cborHead(raw)
	if err != nil {
		return 0, err
	}

	if major != 5 {
		return 0, fmt.Errorf("cose alg: %w", errCOSE)
	}

	var (
		alg   int64
		found bool
	)

	for range n {
		key, next, err := cborInt(rest)
		if err != nil {
			return 0, err
		}

		if key == 3 {
			val, next, err := cborInt(next)
			if err != nil {
				return 0, err
			}

			alg = val
			found = true
			rest = next

			continue
		}

		next, err = cborSkip(next)
		if err != nil {
			return 0, err
		}

		rest = next
	}

	if len(rest) != 0 || !found {
		return 0, fmt.Errorf("cose alg: %w", errCOSE)
	}

	return alg, nil
}

func cborHead(raw []byte) (byte, uint64, []byte, error) {
	if len(raw) == 0 {
		return 0, 0, nil, errCOSE
	}

	b := raw[0]
	major := b >> 5
	ai := b & 0x1f
	rest := raw[1:]
	if ai < 24 {
		return major, uint64(ai), rest, nil
	}

	width := 0
	switch ai {
	case 24:
		width = 1
	case 25:
		width = 2
	case 26:
		width = 4
	case 27:
		width = 8
	default:
		return 0, 0, nil, errCOSE
	}

	if len(rest) < width {
		return 0, 0, nil, errCOSE
	}

	var n uint64
	for i := range width {
		n = n<<8 | uint64(rest[i])
	}

	return major, n, rest[width:], nil
}

func cborInt(raw []byte) (int64, []byte, error) {
	major, n, rest, err := cborHead(raw)
	if err != nil {
		return 0, nil, err
	}

	if n > ^uint64(0)>>1 {
		return 0, nil, errCOSE
	}

	switch major {
	case 0:
		return int64(n), rest, nil
	case 1:
		return -1 - int64(n), rest, nil
	default:
		return 0, nil, errCOSE
	}
}

func cborSkip(raw []byte) ([]byte, error) {
	major, n, rest, err := cborHead(raw)
	if err != nil {
		return nil, err
	}

	switch major {
	case 0, 1, 7:
		return rest, nil
	case 2, 3:
		if uint64(len(rest)) < n {
			return nil, errCOSE
		}

		return rest[n:], nil
	case 4:
		for range n {
			rest, err = cborSkip(rest)
			if err != nil {
				return nil, err
			}
		}

		return rest, nil
	case 5:
		for range n * 2 {
			rest, err = cborSkip(rest)
			if err != nil {
				return nil, err
			}
		}

		return rest, nil
	case 6:
		return cborSkip(rest)
	default:
		return nil, errCOSE
	}
}
