package library

import (
	"archive/zip"
	"bytes"
	"cmp"
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io/fs"
	"strconv"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/export"
	"github.com/AlexGalhardo/vukem/queue"
	"github.com/AlexGalhardo/vukem/signed"
	"github.com/AlexGalhardo/vukem/storage"
	"github.com/AlexGalhardo/vukem/uuid"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
)

// Export limits (SDD 14 FR-14-12, §7.4).
const (
	ExportCooldown     = 7 * 24 * time.Hour
	ExportLinkTTL      = 24 * time.Hour
	ExportsPrefix      = "/exports"
	CodeExportCooldown = "export-cooldown"
)

// ExportCooldownError refuses an export inside the 7-day cooldown; RetryAfter is in whole seconds. Handlers
// render it as the 429 export-cooldown problem with Retry-After.
type ExportCooldownError struct{ RetryAfter int }

func (ExportCooldownError) Error() string { return CodeExportCooldown }

// ExportJob builds the ZIP of one data_exports row ("library.export", SDD 14 §6.5).
type ExportJob struct {
	ExportID string `json:"exportId"`
}

// Kind names the job on the queue.
func (ExportJob) Kind() string { return "library.export" }

// Exports runs the LGPD data export (SDD 14 §7.4, SDD 03 §6.4).
type Exports struct {
	DB     *database.DB
	Queue  *queue.Queue
	Disk   storage.Disk
	Signer *signed.Signer
	// URL is the public base URL the mailed link starts with (APP_URL).
	URL string
	Now func() time.Time
	// Ready queues the mail with the link; expires is already formatted for the text.
	Ready func(ctx context.Context, userID int64, link, expires string) error
}

// NewExports wires the service: a private local disk under EXPORTS_DIR (default storage/exports) served only
// through links signed with a key derived from APP_KEY.
func NewExports(app *vukem.App, acc *accounts.Accounts) (*Exports, error) {
	q, err := app.Queue()
	if err != nil {
		return nil, err
	}

	disk, err := storage.NewLocal(cmp.Or(app.Source.Get("EXPORTS_DIR"), "storage/exports"), "")
	if err != nil {
		return nil, err
	}

	app.OnShutdown(func(context.Context) error { return disk.Close() })

	key, err := app.DeriveKey("exports.link", 32)
	if err != nil {
		return nil, err
	}

	return &Exports{
		DB: acc.DB, Queue: q, Disk: disk, Signer: signed.New(key).WithClock(app.Now),
		URL: app.Config.URL, Now: app.Now, Ready: acc.SendExportReady,
	}, nil
}

// Request queues an export for userID. A queued, running, ready or expired export younger than 7 days answers
// 429 export-cooldown with Retry-After; a failed one does not count.
func (s *Exports) Request(ctx context.Context, userID int64) (string, error) {
	now := s.Now().UTC()
	id := uuid.NewV7().String()

	err := s.DB.Tx(ctx, func(tx *database.Tx) error {
		// The user row lock serialises two concurrent requests, so only one passes the cooldown.
		if _, err := tx.ExecContext(ctx, `SELECT 1 FROM users WHERE id = $1 FOR UPDATE`, userID); err != nil {
			return err
		}

		var last sql.NullTime
		err := tx.QueryRowContext(ctx,
			`SELECT max(requested_at) FROM data_exports WHERE user_id = $1 AND status <> 'failed'`, userID).Scan(&last)
		if err != nil {
			return err
		}

		if last.Valid && now.Sub(last.Time) < ExportCooldown {
			return ExportCooldownError{RetryAfter: int(last.Time.Add(ExportCooldown).Sub(now).Seconds()) + 1}
		}

		if _, err := tx.ExecContext(ctx,
			`INSERT INTO data_exports (id, user_id, status, requested_at) VALUES ($1, $2, 'queued', $3)`,
			id, userID, now); err != nil {
			return err
		}

		return s.Queue.DispatchTx(ctx, tx, ExportJob{ExportID: id})
	})

	return id, err
}

// Build writes the ZIP of export id, marks it ready, records the account event and mails the link. A row that
// is gone or already settled is skipped, so a redelivered job does nothing.
func (s *Exports) Build(ctx context.Context, id string) error {
	var (
		userID int64
		status string
	)

	err := s.DB.QueryRowContext(ctx, `SELECT user_id, status FROM data_exports WHERE id = $1`, id).
		Scan(&userID, &status)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && status != "queued" && status != "running") {
		return nil
	}

	if err != nil {
		return err
	}

	res, err := s.DB.ExecContext(ctx, `UPDATE data_exports SET status = 'running' WHERE id = $1`, id)
	if err != nil {
		return err
	}

	if n, _ := res.RowsAffected(); n == 0 {
		return nil
	}

	data, err := s.zip(ctx, userID)
	if err != nil {
		return err
	}

	path := strconv.FormatInt(userID, 10) + "/" + id + ".zip"
	if err := storage.PutBytes(ctx, s.Disk, path, data); err != nil {
		return err
	}

	link, err := storage.TemporaryURL(s.Signer, ExportsPrefix, path, ExportLinkTTL)
	if err != nil {
		return err
	}

	now := s.Now().UTC()
	expires := now.Add(ExportLinkTTL)

	settled := true

	err = s.DB.Tx(ctx, func(tx *database.Tx) error {
		// A purge that ran meanwhile deleted the row: no row, no ZIP and no mail.
		res, err := tx.ExecContext(ctx, `UPDATE data_exports
SET status = 'ready', path = $2, size_bytes = $3, ready_at = $4, expires_at = $5
WHERE id = $1 AND status = 'running'`,
			id, path, len(data), now, expires)
		if err != nil {
			return err
		}

		if n, _ := res.RowsAffected(); n == 0 {
			settled = false

			return nil
		}

		_, err = tx.ExecContext(ctx, `INSERT INTO account_events (user_id, kind, data, created_at)
VALUES ($1, 'exported', jsonb_build_object('exportId', $2::text), $3)`, userID, id, now)

		return err
	})
	if err != nil {
		return err
	}

	if !settled {
		return s.deleteFile(ctx, path)
	}

	if s.Ready == nil {
		return nil
	}

	return s.Ready(ctx, userID, s.URL+link, expires.Format("2006-01-02 15:04 UTC"))
}

// Fail marks an export failed after its last attempt, so the user may ask again at once.
func (s *Exports) Fail(ctx context.Context, id string) error {
	_, err := s.DB.ExecContext(ctx,
		`UPDATE data_exports SET status = 'failed' WHERE id = $1 AND status IN ('queued', 'running')`, id)

	return err
}

type dueExport struct{ id, path string }

// due lists the ready exports whose link has expired.
func (s *Exports) due(ctx context.Context) ([]dueExport, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, path FROM data_exports WHERE status = 'ready' AND expires_at <= $1`, s.Now().UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []dueExport

	for rows.Next() {
		var d dueExport
		if err := rows.Scan(&d.id, &d.path); err != nil {
			return nil, err
		}

		all = append(all, d)
	}

	return all, rows.Err()
}

// Expire deletes the files of ready exports past expires_at and marks them expired (NFR-14-6).
func (s *Exports) Expire(ctx context.Context) (int, error) {
	all, err := s.due(ctx)
	if err != nil {
		return 0, err
	}

	for i, d := range all {
		if err := s.Disk.Delete(ctx, d.path); err != nil {
			return i, err
		}

		if _, err := s.DB.ExecContext(
			ctx,
			`UPDATE data_exports SET status = 'expired' WHERE id = $1`,
			d.id,
		); err != nil {
			return i, err
		}
	}

	return len(all), nil
}

// Item columns shared by the export queries: $2 is the user's locale, titles fall back to English, then the
// original title. External ids include the Wikidata QID.
const (
	exportTitle   = `coalesce(t.title, te.title, i.original_title, i.slug)`
	exportYear    = `extract(year FROM i.release_date)::int`
	exportExtIDs  = `jsonb_strip_nulls(coalesce(i.external_ids, '{}'::jsonb) || jsonb_build_object('wikidata', i.wikidata_qid))`
	exportTitleJn = `LEFT JOIN item_translations t ON t.item_id = i.id AND t.locale = $2
LEFT JOIN item_translations te ON te.item_id = i.id AND te.locale = 'en'`
)

// exportJSON are the JSON files; each query takes ($1 user id, $2 locale) and returns one JSON value. Stars
// are 0.5–10 (stored 5–100), time-to-beat in minutes. Secrets (password, 2FA, token hashes), actor ids and other
// users' data are never selected (SDD 03 §6.4).
var exportJSON = []struct{ name, query string }{
	{"profile", `SELECT json_build_object('id', id, 'email', email, 'name', name, 'username', username,
  'locale', locale, 'profilePrivate', profile_private, 'personalizationOptOut', personalization_opt_out,
  'createdAt', created_at)
FROM users WHERE id = $1::bigint AND $2::text IS NOT NULL`},
	{
		"ratings",
		`SELECT coalesce(json_agg(json_build_object('itemId', i.id, 'type', i.type, 'title', ` + exportTitle + `,
  'year', ` + exportYear + `, 'externalIds', ` + exportExtIDs + `, 'overall', round(r.overall / 10, 2),
  'c1', round(r.c1 / 10.0, 1), 'c2', round(r.c2 / 10.0, 1), 'c3', round(r.c3 / 10.0, 1),
  'c4', round(r.c4 / 10.0, 1), 'c5', round(r.c5 / 10.0, 1), 'c6', round(r.c6 / 10.0, 1),
  'ratedAt', r.created_at, 'updatedAt', r.updated_at) ORDER BY r.created_at), '[]')
FROM ratings r JOIN items i ON i.id = r.item_id ` + exportTitleJn + `
WHERE r.user_id = $1::bigint`,
	},
	{
		"statuses",
		`SELECT coalesce(json_agg(json_build_object('itemId', i.id, 'type', i.type, 'title', ` + exportTitle + `,
  'status', s.status, 'updatedAt', s.updated_at) ORDER BY s.updated_at), '[]')
FROM item_status s JOIN items i ON i.id = s.item_id ` + exportTitleJn + `
WHERE s.user_id = $1::bigint`,
	},
	{
		"time_to_beat",
		`SELECT coalesce(json_agg(json_build_object('itemId', i.id, 'type', i.type, 'title', ` + exportTitle + `,
  'mainMinutes', b.main_minutes, 'extrasMinutes', b.extras_minutes, 'completionistMinutes', b.completionist_minutes,
  'createdAt', b.created_at, 'updatedAt', b.updated_at) ORDER BY b.created_at), '[]')
FROM time_to_beat b JOIN items i ON i.id = b.item_id ` + exportTitleJn + `
WHERE b.user_id = $1::bigint`,
	},
	{
		"editions",
		`SELECT coalesce(json_agg(json_build_object('itemId', i.id, 'editionId', e.id, 'isbn13', e.isbn13,
  'title', e.title, 'updatedAt', ue.updated_at) ORDER BY ue.updated_at), '[]')
FROM user_item_editions ue JOIN items i ON i.id = ue.item_id JOIN editions e ON e.id = ue.edition_id
WHERE ue.user_id = $1::bigint AND $2::text IS NOT NULL`,
	},
	// shortcut: edit suggestions (SDD 04) have no table yet; fill this when it lands so the ZIP keeps its 12 files.
	{"suggestions", `SELECT '[]'::json WHERE $1::bigint IS NOT NULL AND $2::text IS NOT NULL`},
	{"lists", `SELECT coalesce(json_agg(json_build_object('id', l.id, 'type', l.type, 'title', l.title,
  'description', l.description, 'isPublic', l.is_public, 'hidden', l.hidden,
  'createdAt', l.created_at, 'updatedAt', l.updated_at,
  'items', (SELECT coalesce(json_agg(json_build_object('itemId', i.id, 'type', i.type, 'title', ` + exportTitle + `,
      'year', ` + exportYear + `, 'position', li.position, 'note', li.note, 'addedAt', li.added_at) ORDER BY li.position), '[]')
    FROM list_items li JOIN items i ON i.id = li.item_id ` + exportTitleJn + `
    WHERE li.list_id = l.id)) ORDER BY l.created_at), '[]')
FROM lists l WHERE l.user_id = $1::bigint`},
	{"notifications", `SELECT coalesce(json_agg(json_build_object('kind', kind, 'payload', payload, 'readAt', read_at,
  'createdAt', created_at) ORDER BY created_at), '[]')
FROM user_notifications WHERE user_id = $1::bigint AND $2::text IS NOT NULL`},
	// apiKeys and subscriptions wait for SDD 05 keys and SDD 11 billing tables.
	{"account", `SELECT json_build_object(
  'socialAccounts', (SELECT coalesce(json_agg(json_build_object('provider', provider, 'createdAt', created_at)
    ORDER BY created_at), '[]') FROM social_accounts WHERE user_id = $1::bigint),
  'apiKeys', '[]'::json,
  'subscriptions', '[]'::json,
  'sessions', (SELECT coalesce(json_agg(json_build_object('expiresAt', expires_at) ORDER BY expires_at), '[]')
    FROM remember_tokens WHERE user_id = $1::bigint::text),
  'tokens', (SELECT coalesce(json_agg(json_build_object('name', name, 'abilities', abilities,
      'createdAt', created_at, 'lastUsedAt', last_used_at, 'expiresAt', expires_at, 'revokedAt', revoked_at)
    ORDER BY created_at), '[]') FROM access_tokens WHERE user_id = $1::bigint::text),
  'events', (SELECT coalesce(json_agg(json_build_object('kind', kind, 'data', data, 'createdAt', created_at)
    ORDER BY created_at), '[]') FROM account_events WHERE user_id = $1::bigint))
WHERE $2::text IS NOT NULL`},
}

// exportCSV are the import-friendly CSV files (Goodreads/Letterboxd style); every column is text.
var exportCSV = []struct {
	name   string
	header []string
	query  string
}{
	{
		"ratings",
		[]string{"type", "title", "year", "externalIds", "overall", "c1", "c2", "c3", "c4", "c5", "c6", "ratedAt"},
		`SELECT i.type, ` + exportTitle + `, coalesce(` + exportYear + `::text, ''), ` + exportExtIDs + `::text,
  round(r.overall / 10, 2)::text, round(r.c1 / 10.0, 1)::text, round(r.c2 / 10.0, 1)::text,
  round(r.c3 / 10.0, 1)::text, round(r.c4 / 10.0, 1)::text, round(r.c5 / 10.0, 1)::text,
  round(r.c6 / 10.0, 1)::text, to_char(r.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
FROM ratings r JOIN items i ON i.id = r.item_id ` + exportTitleJn + `
WHERE r.user_id = $1::bigint ORDER BY r.created_at`,
	},
	{
		"lists",
		[]string{"list", "position", "type", "title", "year", "note", "addedAt"},
		`SELECT l.title, li.position::text, i.type, ` + exportTitle + `, coalesce(` + exportYear + `::text, ''),
  coalesce(li.note, ''), to_char(li.added_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
FROM lists l JOIN list_items li ON li.list_id = l.id JOIN items i ON i.id = li.item_id ` + exportTitleJn + `
WHERE l.user_id = $1::bigint ORDER BY l.created_at, li.position`,
	},
}

// zip reads everything in one repeatable-read snapshot (SDD 03 §6.4) and returns the archive.
func (s *Exports) zip(ctx context.Context, userID int64) ([]byte, error) {
	var buf bytes.Buffer

	zw := zip.NewWriter(&buf)

	err := s.DB.Tx(ctx, func(tx *database.Tx) error {
		if _, err := tx.ExecContext(ctx, `SET TRANSACTION ISOLATION LEVEL REPEATABLE READ, READ ONLY`); err != nil {
			return err
		}

		var locale string
		if err := tx.QueryRowContext(ctx, `SELECT coalesce(locale, 'en') FROM users WHERE id = $1`, userID).
			Scan(&locale); err != nil {
			return fmt.Errorf("export user %d: %w", userID, err)
		}

		for _, f := range exportJSON {
			var raw []byte
			if err := tx.QueryRowContext(ctx, f.query, userID, locale).Scan(&raw); err != nil {
				return fmt.Errorf("export %s: %w", f.name, err)
			}

			if err := zipFile(zw, f.name+".json", raw); err != nil {
				return err
			}
		}

		for _, f := range exportCSV {
			if err := zipCSV(ctx, tx, zw, f.name+".csv", f.header, f.query, userID, locale); err != nil {
				return fmt.Errorf("export %s.csv: %w", f.name, err)
			}
		}

		return zipFile(zw, "README.md", []byte(exportReadme))
	})
	if err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func zipFile(zw *zip.Writer, name string, data []byte) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}

	_, err = w.Write(data)

	return err
}

// zipCSV writes a CSV file. workaround(vukem): VF-029: export.CSV only writes to an http.ResponseWriter, so
// encoding/csv plus export.SafeCell applies the same formula guard per cell (P0.F.9.1).
func zipCSV(ctx context.Context, tx *database.Tx, zw *zip.Writer, name string, header []string, query string,
	args ...any,
) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}

	cw := csv.NewWriter(w)
	if err := cw.Write(header); err != nil {
		return err
	}

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	cells := make([]string, len(header))
	ptrs := make([]any, len(header))

	for i := range cells {
		ptrs[i] = &cells[i]
	}

	for rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			return err
		}

		safe := make([]string, len(cells))
		for i, c := range cells {
			safe[i] = export.SafeCell(c)
		}

		if err := cw.Write(safe); err != nil {
			return err
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	cw.Flush()

	return cw.Error()
}

const exportReadme = `# HexaScore data export

Your personal records on HexaScore, read at one instant. Dates are ISO 8601. Scores are stars from 0.5 to 10
(c1..c6 are the six categories of the item type, overall is their mean). Times are minutes.

- profile.json: account profile (id, email, name, username, locale, privacy settings, createdAt).
- ratings.json: your ratings with item id, type, title, year, external ids, overall, c1..c6, ratedAt, updatedAt.
- statuses.json: your library statuses per item.
- time_to_beat.json: the play times you reported (mainMinutes, extrasMinutes, completionistMinutes).
- editions.json: the edition you chose per book (empty until editions can be chosen).
- lists.json: your lists with their items, positions and notes.
- suggestions.json: your catalog edit suggestions (empty until suggestions open).
- notifications.json: notifications sent to you.
- account.json: linked sign-in providers, API keys, subscriptions, remembered sessions (expiry only), app tokens
  (name, abilities, dates; never the secret) and account events.
- ratings.csv: type,title,year,externalIds,overall,c1..c6,ratedAt — for import into other services.
- lists.csv: list,position,type,title,year,note,addedAt.

CSV cells that start with =, +, -, @, tab or carriage return are prefixed with ' so spreadsheets do not run them.
Passwords, two-factor secrets, token hashes and other people's data are never included.
`

// deleteFile removes a file; one that is already gone counts as deleted.
func (s *Exports) deleteFile(ctx context.Context, p string) error {
	if err := s.Disk.Delete(ctx, p); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	return nil
}

// Purge deletes every export file of userID (rows' paths and the user's whole export directory) and their
// rows, so no ZIP survives the account purge (SDD 14 §8). Missing files are fine; a failing file keeps its
// row for the next run while the others still go, and all errors are joined. A Build still running finds its
// row gone at its final step and deletes its own file.
func (s *Exports) Purge(ctx context.Context, userID int64) error {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, coalesce(path, '') FROM data_exports WHERE user_id = $1 AND status <> 'expired'`, userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var all []dueExport

	for rows.Next() {
		var d dueExport
		if err := rows.Scan(&d.id, &d.path); err != nil {
			return err
		}

		all = append(all, d)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	var errs error

	files, err := s.Disk.List(ctx, strconv.FormatInt(userID, 10))
	if err != nil {
		errs = err
	}

	failed := map[string]bool{}

	for _, f := range files {
		if err := s.deleteFile(ctx, f); err != nil {
			failed[f] = true
			errs = errors.Join(errs, err)
		}
	}

	for _, d := range all {
		if d.path != "" && failed[d.path] {
			continue
		}

		if d.path != "" {
			if err := s.deleteFile(ctx, d.path); err != nil {
				errs = errors.Join(errs, err)

				continue
			}
		}

		if _, err := s.DB.ExecContext(ctx, `DELETE FROM data_exports WHERE id = $1`, d.id); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	if _, err := s.DB.ExecContext(ctx, `DELETE FROM data_exports WHERE user_id = $1 AND status = 'expired'`,
		userID); err != nil {
		errs = errors.Join(errs, err)
	}

	return errs
}

// PurgeAnonymised purges the exports of every anonymised user that still has rows, so a purge that failed on
// an earlier run is retried. Returns the joined per-user errors.
func (s *Exports) PurgeAnonymised(ctx context.Context) error {
	rows, err := s.DB.QueryContext(ctx, `SELECT DISTINCT e.user_id FROM data_exports e
JOIN users u ON u.id = e.user_id WHERE u.anonymized_at IS NOT NULL`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var ids []int64

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return err
		}

		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	var errs error
	for _, id := range ids {
		errs = errors.Join(errs, s.Purge(ctx, id))
	}

	return errs
}
