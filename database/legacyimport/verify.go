package legacyimport

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/storage"

	"github.com/AlexGalhardo/hexascore-app/app/scoring"
)

// ErrVerifyFailed means at least one cutover check failed; the output names each table.
var ErrVerifyFailed = errors.New("cutover-verify: checks failed")

// VerifyOptions configure Verify.
type VerifyOptions struct {
	// CountsOnly runs only the per-table counts (the fast pre-check).
	CountsOnly bool
	// SamplePct is the share of ids (1..100) whose rows are hashed on both sides.
	SamplePct int
	// Disk holds the cover objects; required unless CountsOnly.
	Disk storage.Disk
	Out  io.Writer
}

// countCheck pairs a legacy count with the imported one. The legacy query
// applies the documented exclusions (SDD 11 §7.6 step 6), the same filters
// the stages apply. Not compared: session, verification, rate_limit,
// external_scores, follows, processed_events (never imported), credential
// accounts (folded into users.password), and passkey (imported by the users
// stage; incomplete rows are skipped, so it is not a count pair).
type countCheck struct{ name, legacy, imported string }

var countChecks = []countCheck{
	{"users", `SELECT count(*) FROM legacy_import."user"`, `SELECT count(*) FROM users WHERE legacy_id IS NOT NULL`},
	{
		"social_accounts",
		`SELECT count(*) FROM legacy_import.account WHERE provider_id = 'google'`,
		`SELECT count(*) FROM social_accounts WHERE provider = 'google'`,
	},
	{
		"two_factor",
		`SELECT count(*) FROM legacy_import.two_factor`,
		`SELECT count(*) FROM users WHERE legacy_id IS NOT NULL AND two_factor_secret IS NOT NULL`,
	},
	{"items", `SELECT count(*) FROM legacy_import.items`, `SELECT count(*) FROM items`},
	{
		"item_translations",
		`SELECT count(*) FROM legacy_import.item_translations`,
		`SELECT count(*) FROM item_translations`,
	},
	{"affiliate_links", `SELECT count(*) FROM legacy_import.affiliate_links`, `SELECT count(*) FROM affiliate_links`},
	{"source_records", `SELECT count(*) FROM legacy_import.source_records`, `SELECT count(*) FROM source_records`},
	{"ratings", `SELECT count(*) FROM legacy_import.ratings`, `SELECT count(*) FROM ratings`},
	{"item_status", `SELECT count(*) FROM legacy_import.item_status`, `SELECT count(*) FROM item_status`},
	{"time_to_beat", `SELECT count(*) FROM legacy_import.time_to_beat`, `SELECT count(*) FROM time_to_beat`},
	{"lists", `SELECT count(*) FROM legacy_import.lists`, `SELECT count(*) FROM lists`},
	{"list_items", `SELECT count(*) FROM legacy_import.list_items`, `SELECT count(*) FROM list_items`},
	{
		// Reports whose target user or list was not imported are dropped by the social stage.
		"reports",
		`SELECT count(*) FROM legacy_import.reports r WHERE CASE r.target_type::text
  WHEN 'user' THEN EXISTS (SELECT 1 FROM legacy_import."user" u WHERE u.id = r.target_id)
  ELSE EXISTS (SELECT 1 FROM legacy_import.lists l WHERE l.id::text = r.target_id) END`,
		`SELECT count(*) FROM reports`,
	},
	{
		"contact_messages",
		`SELECT count(*) FROM legacy_import.contact_messages`,
		`SELECT count(*) FROM contact_messages`,
	},
	{"api_keys", `SELECT count(*) FROM legacy_import.api_keys`, `SELECT count(*) FROM api_keys`},
	{"api_usage_daily", `SELECT count(*) FROM legacy_import.api_usage_daily`, `SELECT count(*) FROM api_usage_daily`},
	{
		// Free plans are the default state, not a subscription row (SDD 05 §5).
		"subscriptions",
		`SELECT count(*) FROM legacy_import.subscriptions WHERE plan::text <> 'free'`,
		`SELECT count(*) FROM subscriptions`,
	},
	{
		"stripe_customers",
		`SELECT count(*) FROM legacy_import.stripe_customers`,
		`SELECT count(*) FROM users WHERE legacy_id IS NOT NULL AND stripe_customer_id IS NOT NULL`,
	},
	{"credit_ledger", `SELECT count(*) FROM legacy_import.credit_ledger`, `SELECT count(*) FROM credit_ledger`},
	{"pix_charges", `SELECT count(*) FROM legacy_import.pix_charges`, `SELECT count(*) FROM pix_charges`},
}

// sampleCheck hashes the columns copied verbatim, ordered by the legacy id,
// for the ids whose hash falls in the sample. Both sides select the same
// columns as text in the same order.
type sampleCheck struct{ table, legacy, imported string }

var sampleChecks = []sampleCheck{
	{
		"users", `SELECT u.id AS k, concat_ws('|', u.id, lower(u.email)) AS v FROM legacy_import."user" u`,
		`SELECT legacy_id AS k, concat_ws('|', legacy_id, email) AS v FROM users WHERE legacy_id IS NOT NULL`,
	},
	{
		"items",
		`SELECT id::text AS k, concat_ws('|', id, type::text, slug, wikidata_qid) AS v FROM legacy_import.items`,
		`SELECT id::text AS k, concat_ws('|', id, type, slug, wikidata_qid) AS v FROM items`,
	},
	{
		"lists", `SELECT id::text AS k, concat_ws('|', id, title, is_public, hidden) AS v FROM legacy_import.lists`,
		`SELECT id::text AS k, concat_ws('|', id, title, is_public, hidden) AS v FROM lists`,
	},
	{
		"api_keys",
		`SELECT id::text AS k, concat_ws('|', id, prefix, hash, revoked_at) AS v FROM legacy_import.api_keys`,
		`SELECT id::text AS k, concat_ws('|', id, prefix, hash, revoked_at) AS v FROM api_keys`,
	},
	{
		"credit_ledger",
		`SELECT id::text AS k, concat_ws('|', id, delta, reason, ref) AS v FROM legacy_import.credit_ledger`,
		`SELECT id::text AS k, concat_ws('|', id, delta, reason, ref) AS v FROM credit_ledger`,
	},
	{
		"pix_charges",
		`SELECT id::text AS k, concat_ws('|', id, provider_ref, amount_cents) AS v FROM legacy_import.pix_charges`,
		`SELECT id::text AS k, concat_ws('|', id, provider_ref, amount_cents) AS v FROM pix_charges`,
	},
	{
		"contact_messages",
		`SELECT id::text AS k, concat_ws('|', id, email, subject, message) AS v FROM legacy_import.contact_messages`,
		`SELECT id::text AS k, concat_ws('|', id, email, subject, message) AS v FROM contact_messages`,
	},
}

const sampleSQL = `SELECT coalesce(md5(string_agg(v, '|' ORDER BY k)), '') FROM (%s) s WHERE abs(hashtext(k)) %% 100 < $1`

// scoreTolerance is half the last digit of items.score, numeric(7,4).
const scoreTolerance = 1e-4

// Verify runs the cutover checks of SDD 11 §7.6 step 6 against the imported
// database and the restored legacy_import schema. It writes one line per
// check and returns ErrVerifyFailed naming the failed tables. It only reads.
func Verify(ctx context.Context, db *database.DB, o VerifyOptions) error {
	v := &verifier{db: db, out: o.Out}

	for _, c := range countChecks {
		var legacy, imported int64
		if err := db.QueryRowContext(ctx, c.legacy).Scan(&legacy); err != nil {
			return fmt.Errorf("count legacy %s: %w", c.name, err)
		}

		if err := db.QueryRowContext(ctx, c.imported).Scan(&imported); err != nil {
			return fmt.Errorf("count %s: %w", c.name, err)
		}

		v.check(legacy == imported, "count", c.name, "legacy=%d imported=%d", legacy, imported)
	}

	if o.CountsOnly {
		return v.result()
	}

	pct := min(max(o.SamplePct, 1), 100)
	for _, c := range sampleChecks {
		var legacy, imported string
		if err := db.QueryRowContext(ctx, fmt.Sprintf(sampleSQL, c.legacy), pct).Scan(&legacy); err != nil {
			return fmt.Errorf("sample legacy %s: %w", c.table, err)
		}

		if err := db.QueryRowContext(ctx, fmt.Sprintf(sampleSQL, c.imported), pct).Scan(&imported); err != nil {
			return fmt.Errorf("sample %s: %w", c.table, err)
		}

		v.check(legacy == imported, "sample", c.table, "%d%% md5 legacy=%s imported=%s", pct, legacy, imported)
	}

	steps := []func(context.Context) error{
		func(ctx context.Context) error { return v.covers(ctx, o.Disk) },
		v.orphans,
		v.scores,
		v.legacyScores,
	}
	for _, step := range steps {
		if err := step(ctx); err != nil {
			return err
		}
	}

	return v.result()
}

type verifier struct {
	db     *database.DB
	out    io.Writer
	failed []string
}

func (v *verifier) check(ok bool, kind, table, format string, args ...any) {
	status := "ok"
	if !ok {
		status = "FAIL"
		v.failed = append(v.failed, kind+" "+table)
	}

	_, _ = fmt.Fprintf(v.out, "%-4s %-6s %-20s %s\n", status, kind, table, fmt.Sprintf(format, args...))
}

func (v *verifier) result() error {
	if len(v.failed) == 0 {
		_, _ = fmt.Fprintln(v.out, "cutover-verify: all checks passed")
		return nil
	}

	return fmt.Errorf("%w: %s", ErrVerifyFailed, strings.Join(v.failed, ", "))
}

// covers checks every items.cover_key and cover_thumb_key against the bucket.
func (v *verifier) covers(ctx context.Context, disk storage.Disk) error {
	if disk == nil {
		return errors.New("cutover-verify: no cover disk")
	}

	keys, err := collect(ctx, v.db, func(r *sql.Rows, k *string) error { return r.Scan(k) },
		`SELECT k FROM (SELECT cover_key AS k FROM items
UNION ALL SELECT cover_thumb_key FROM items) c WHERE k IS NOT NULL ORDER BY k`)
	if err != nil {
		return fmt.Errorf("list cover keys: %w", err)
	}

	var missing []string
	for _, k := range keys {
		ok, err := disk.Exists(ctx, k)
		if err != nil {
			return fmt.Errorf("stat cover %s: %w", k, err)
		}

		if !ok {
			missing = append(missing, k)
		}
	}

	for _, k := range missing[:min(len(missing), 20)] {
		_, _ = fmt.Fprintf(v.out, "     missing cover %s\n", k)
	}

	v.check(len(missing) == 0, "covers", "items", "%d keys, %d missing", len(keys), len(missing))

	return nil
}

// orphanFKs builds one count per foreign key of the public schema: child rows
// whose non-null key has no parent. Constraints normally make it zero; the
// check catches rows loaded with triggers off or NOT VALID constraints.
const orphanFKs = `SELECT c.conname, c.conrelid::regclass::text,
  format('SELECT count(*) FROM %s ch WHERE %s AND NOT EXISTS (SELECT 1 FROM %s p WHERE %s)',
    c.conrelid::regclass, string_agg(format('ch.%I IS NOT NULL', ca.attname), ' AND '),
    c.confrelid::regclass, string_agg(format('p.%I = ch.%I', pa.attname, ca.attname), ' AND '))
FROM pg_constraint c
CROSS JOIN LATERAL unnest(c.conkey, c.confkey) AS k(child, parent)
JOIN pg_attribute ca ON ca.attrelid = c.conrelid AND ca.attnum = k.child
JOIN pg_attribute pa ON pa.attrelid = c.confrelid AND pa.attnum = k.parent
WHERE c.contype = 'f' AND c.connamespace = 'public'::regnamespace
GROUP BY c.oid, c.conname, c.conrelid, c.confrelid ORDER BY 2, 1`

// reportOrphans covers reports.target_id, a polymorphic reference without a constraint.
const reportOrphans = `SELECT count(*) FROM reports r WHERE CASE r.target_type
  WHEN 'user' THEN NOT EXISTS (SELECT 1 FROM users u WHERE u.id::text = r.target_id)
  ELSE NOT EXISTS (SELECT 1 FROM lists l WHERE l.id::text = r.target_id) END`

func (v *verifier) orphans(ctx context.Context) error {
	type fk struct{ name, table, query string }

	fks, err := collect(
		ctx,
		v.db,
		func(r *sql.Rows, f *fk) error { return r.Scan(&f.name, &f.table, &f.query) },
		orphanFKs,
	)
	if err != nil {
		return fmt.Errorf("list foreign keys: %w", err)
	}

	fks = append(fks, fk{"reports_target", "reports", reportOrphans})

	total := 0
	for _, f := range fks {
		var n int64
		// The query is built by Postgres from the catalog (format with %I), never from input.
		if err := v.db.QueryRowContext(ctx, f.query).Scan(&n); err != nil {
			return fmt.Errorf("orphans %s: %w", f.name, err)
		}

		if n > 0 {
			total++
			v.check(false, "orphan", f.table, "%s: %d rows", f.name, n)
		}
	}

	if total == 0 {
		_, _ = fmt.Fprintf(v.out, "ok   orphan %-20s %d foreign keys, 0 orphans\n", "all", len(fks))
	}

	return nil
}

// scores compares the stored SQL recompute with the Go formula for every
// computed item, at the instant the SQL ran (weights depend on account age).
func (v *verifier) scores(ctx context.Context) error {
	type item struct {
		id    string
		at    time.Time
		n     int
		score sql.NullFloat64
	}

	var (
		after      = "00000000-0000-0000-0000-000000000000"
		checked    int
		mismatched []string
	)

	for {
		var page []item

		err := v.db.Tx(ctx, func(tx *database.Tx) error {
			var err error

			page, err = collect(
				ctx,
				tx,
				func(r *sql.Rows, it *item) error { return r.Scan(&it.id, &it.at, &it.n, &it.score) },
				`SELECT id::text, score_computed_at, rating_count, score FROM items
WHERE id > $1::uuid AND score_computed_at IS NOT NULL ORDER BY id LIMIT 500`,
				after,
			)
			if err != nil {
				return err
			}

			for _, it := range page {
				res, err := scoring.ComputeItem(ctx, tx, it.id, it.at)
				if err != nil {
					return fmt.Errorf("go score %s: %w", it.id, err)
				}

				same := res.N == it.n && (res.N == 0 && !it.score.Valid ||
					it.score.Valid && math.Abs(res.B-it.score.Float64) <= scoreTolerance)
				if !same {
					mismatched = append(mismatched, it.id)
					_, _ = fmt.Fprintf(v.out, "     score %s sql=%v go=%.4f n=%d/%d\n",
						it.id, it.score.Float64, res.B, it.n, res.N)
				}
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("compare scores: %w", err)
		}

		checked += len(page)
		if len(page) < 500 {
			break
		}

		after = page[len(page)-1].id
	}

	v.check(len(mismatched) == 0, "score", "items", "%d items, sql vs go: %d differ", checked, len(mismatched))

	return nil
}

// legacyScores is the sanity check on the copy: the legacy aggregate against
// the mean of the imported (rounded) ratings. The formulas differ by design,
// so an item off by more than 1 point is listed, not failed.
func (v *verifier) legacyScores(ctx context.Context) error {
	rows, err := v.db.QueryContext(ctx, `SELECT l.id::text, l.score, round(avg(r.overall), 2)
FROM legacy_import.items l JOIN ratings r ON r.item_id = l.id
WHERE l.score IS NOT NULL GROUP BY l.id, l.score HAVING abs(l.score - avg(r.overall)) > 1 ORDER BY l.id`)
	if err != nil {
		return fmt.Errorf("legacy scores: %w", err)
	}
	defer func() { _ = rows.Close() }()

	n := 0
	for rows.Next() {
		var (
			id     string
			legacy int
			mean   float64
		)
		if err := rows.Scan(&id, &legacy, &mean); err != nil {
			return err
		}

		n++
		_, _ = fmt.Fprintf(v.out, "     legacy score drift %s legacy=%d imported_mean=%.2f\n", id, legacy, mean)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(v.out, "info legacy %-20s %d items beyond ±1 (listed above; sanity check only)\n", "items", n)

	return nil
}

type querier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// collect runs query and scans every row into a T.
func collect[T any](
	ctx context.Context,
	q querier,
	scan func(*sql.Rows, *T) error,
	query string,
	args ...any,
) ([]T, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []T
	for rows.Next() {
		var v T
		if err := scan(rows, &v); err != nil {
			return nil, err
		}

		out = append(out, v)
	}

	return out, rows.Err()
}
