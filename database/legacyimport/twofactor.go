package legacyimport

import (
	"bufio"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/orm"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// TwoFactorSecret is one NDJSON line of scripts/legacy/twofactor-export.ts:
// the Better Auth TOTP secret and backup codes, already decrypted.
type TwoFactorSecret struct {
	UserID      string   `json:"userId"`
	Secret      string   `json:"secret"`
	BackupCodes []string `json:"backupCodes"`
}

// ErrNeedsStdin means the twofactor stage ran without the export piped in.
var ErrNeedsStdin = errors.New("legacyimport: the twofactor stage reads the export on stdin (--stdin)")

// TwoFactorBatch imports the decrypted 2FA secrets read from r (SDD 03 §7.6).
// The export is sorted by user id, so a resumed run skips lines up to the
// checkpoint key. The secret and the code hashes are only ever held in memory
// and sealed with APP_KEY; lockout counters are dropped.
func TwoFactorBatch(r io.Reader, now func() time.Time) BatchFunc {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)

	var prev string // last user id read from the stream, across batches

	return func(ctx context.Context, tx *database.Tx, after string, n int) (string, int, error) {
		last, rows := after, 0

		for rows < n && sc.Scan() {
			var s TwoFactorSecret
			if err := json.Unmarshal(sc.Bytes(), &s); err != nil || s.UserID == "" || s.Secret == "" {
				// The line holds a secret: never echo it.
				return "", 0, fmt.Errorf("twofactor: malformed line after user %q", last)
			}

			if s.UserID <= prev {
				return "", 0, fmt.Errorf("twofactor: export not sorted by user id at %q", s.UserID)
			}

			prev = s.UserID

			if s.UserID <= after {
				continue // imported before a resume
			}

			if err := importTwoFactor(ctx, tx, s, now().UTC()); err != nil {
				return "", 0, fmt.Errorf("twofactor %s: %w", s.UserID, err)
			}

			last = s.UserID
			rows++
		}

		if err := sc.Err(); err != nil {
			return "", 0, fmt.Errorf("twofactor: read stdin: %w", err)
		}

		return last, rows, nil
	}
}

func importTwoFactor(ctx context.Context, tx *database.Tx, s TwoFactorSecret, at time.Time) error {
	var (
		id       sql.NullInt64
		verified bool
	)

	err := tx.QueryRowContext(ctx, `SELECT n.id, coalesce(t.verified, true) FROM legacy_import.two_factor t
LEFT JOIN users n ON n.legacy_id = t.user_id WHERE t.user_id = $1 ORDER BY t.id LIMIT 1`, s.UserID).Scan(&id, &verified)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("no legacy two_factor row")
	}

	if err != nil {
		return err
	}

	if !id.Valid {
		return errors.New("user not imported (run the users stage first)")
	}

	// Better Auth keys HMAC with the secret's UTF-8 bytes and shows them base32
	// encoded in the otpauth URI; vukem's totp stores that base32 form.
	set := orm.Set{
		"two_factor_secret": orm.Encrypted[string]{
			V: base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte(s.Secret)),
		},
		"two_factor_recovery_codes": orm.Encrypted[[]string]{V: recoveryHashes(s.BackupCodes)},
		"two_factor_confirmed_at":   nil,
		"two_factor_last_step":      0,
	}
	if verified {
		set["two_factor_confirmed_at"] = at
	}

	_, err = orm.QueryOf[models.TwoFactor](tx).Where("id", id.Int64).Update(ctx, set)

	return err
}

// recoveryHashes hashes Better Auth's unused backup codes the way vukem's
// totp.UseRecoveryCode checks them.
//
// workaround(vukem): VF-NEW-P2.F.2.3, totp has no exported hash for an
// existing code; TestTwoFactorUsersPassChallenge proves the format still matches.
func recoveryHashes(codes []string) []string {
	out := make([]string, 0, len(codes))

	for _, c := range codes {
		sum := sha256.Sum256([]byte("totp-recovery:" + strings.ToUpper(strings.TrimSpace(c))))
		out = append(out, hex.EncodeToString(sum[:]))
	}

	return out
}
