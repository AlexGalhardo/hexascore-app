package backup

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/signed"
	"github.com/AlexGalhardo/vukem/storage"
	"github.com/AlexGalhardo/vukem/uuid"

	"github.com/AlexGalhardo/hexascore-app/app/logging"
)

// Off-site copy (SDD 11 §7.3 step 6, D-32): the link lives 7 days, is served by the app under
// OffsitePrefix and is signed with a key derived from APP_KEY (NewOffsiteSigner).
const (
	OffsiteTTL    = 7 * 24 * time.Hour
	OffsitePrefix = "/backups/offsite"
)

// monthlyPrefix is the only part of the bucket the off-site link may serve.
const monthlyPrefix = "/backups/postgres/monthly/"

// ErrNoMonthlyDump means this month has no uploaded monthly dump to link to.
var ErrNoMonthlyDump = errors.New("backup: no monthly dump this month")

// OffsiteMail is what the owner receives. The link is the only secret and expires; BACKUP_KEY never travels.
type OffsiteMail struct {
	To      []string
	Link    string
	Key     string
	SHA256  string
	Bytes   int64
	Expires time.Time
}

// Offsite mails ADMIN_EMAILS a signed link to this month's encrypted monthly dump.
type Offsite struct {
	DB     *database.DB
	Disk   storage.Disk
	Signer *signed.Signer
	// URL is APP_URL, the base of the mailed link.
	URL string
	// To is ADMIN_EMAILS.
	To []string
	// Send delivers the mail synchronously, so the link is never stored in the jobs table.
	Send func(ctx context.Context, m OffsiteMail) error
	Now  func() time.Time
}

// Mail sends the link once per monthly dump and records backup_runs(kind='offsite'): uploaded when mailed,
// failed with the scrubbed reason otherwise.
func (o *Offsite) Mail(ctx context.Context) error {
	if len(o.To) == 0 {
		return errors.New("backup: ADMIN_EMAILS is empty, no off-site recipient")
	}

	now := o.Now().UTC()
	month := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	var (
		m      = OffsiteMail{To: o.To, Expires: now.Add(OffsiteTTL)}
		sha    sql.NullString
		size   sql.NullInt64
		mailed bool
	)

	err := o.DB.QueryRowContext(ctx, `SELECT r.object_key, r.sha256, r.bytes,
  EXISTS (SELECT 1 FROM backup_runs s WHERE s.kind = 'offsite' AND s.status = 'uploaded' AND s.object_key = r.object_key)
FROM backup_runs r WHERE r.kind = 'monthly' AND r.status IN ('uploaded', 'verified') AND r.started_at >= $1
ORDER BY r.started_at DESC LIMIT 1`, month).Scan(&m.Key, &sha, &size, &mailed)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w (%s)", ErrNoMonthlyDump, month.Format("2006-01"))
	}

	if err != nil {
		return fmt.Errorf("backup: find monthly dump: %w", err)
	}

	if mailed {
		return nil
	}

	m.SHA256, m.Bytes = sha.String, size.Int64

	err = o.send(ctx, &m)

	status, msg := "uploaded", sql.NullString{}
	if err != nil {
		s := logging.Scrub(err.Error())
		status, msg = "failed", sql.NullString{String: s[:min(len(s), 2000)], Valid: true}
	}

	if _, dbErr := o.DB.ExecContext(context.WithoutCancel(ctx), `INSERT INTO backup_runs
  (id, kind, object_key, bytes, sha256, status, error, started_at, finished_at)
VALUES ($1, 'offsite', $2, $3, $4, $5, $6, $7, $8)`,
		uuid.NewV7().String(), m.Key, size, sha, status, msg, now, o.Now().UTC()); dbErr != nil {
		return errors.Join(err, fmt.Errorf("backup: record offsite: %w", dbErr))
	}

	return err
}

func (o *Offsite) send(ctx context.Context, m *OffsiteMail) error {
	// A pruned object would mail a dead link.
	if ok, err := o.Disk.Exists(ctx, m.Key); err != nil || !ok {
		return errors.Join(fmt.Errorf("backup: monthly dump %s is not in the bucket", m.Key), err)
	}

	link, err := storage.TemporaryURL(o.Signer, OffsitePrefix, m.Key, OffsiteTTL)
	if err != nil {
		return fmt.Errorf("backup: sign offsite link: %w", err)
	}

	m.Link = strings.TrimRight(o.URL, "/") + link

	if err := o.Send(ctx, *m); err != nil {
		return fmt.Errorf("backup: mail offsite link: %w", err)
	}

	return nil
}

// NewOffsiteSigner is the signer of off-site links, keyed by HKDF from APP_KEY: the command that mails a
// link and the web process that serves it derive the same key.
func NewOffsiteSigner(app *vukem.App) (*signed.Signer, error) {
	key, err := app.DeriveKey("backup.offsite", 32)
	if err != nil {
		return nil, fmt.Errorf("backup: offsite key: %w", err)
	}

	return signed.New(key).WithClock(app.Now), nil
}

// OffsiteHandler serves signed links from disk (mount it with http.StripPrefix(OffsitePrefix, ...)). Every
// path outside backups/postgres/monthly/ is a 404 even with a valid signature, so the link can never
// reach covers, daily dumps or anything else in the shared bucket.
func OffsiteHandler(disk storage.Disk, signer *signed.Signer) http.Handler {
	// The dump is a secret behind a bearer link: no browser or proxy (Cloudflare) may keep a copy.
	serve := storage.HandlerWith(disk, signer, storage.Options{CacheControl: "private, no-store"})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, monthlyPrefix) || strings.Contains(r.URL.Path, "..") {
			http.NotFound(w, r)
			return
		}

		serve.ServeHTTP(w, r)
	})
}
