package console

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/app/scoring"
	"github.com/AlexGalhardo/hexascore-app/internal/search"
)

// SeedTable is a column that references a seed user. Rows go by the users FK (cascade), are nulled by it
// (setnull), or are deleted explicitly because the column has no FK (loose). TestSeedPurgeCoversEveryUserFK keeps
// the FK entries in step with the schema.
type SeedTable struct {
	Table, Column string
	Kind          string
	Text          bool // the column holds the id as text
}

// SeedTables lists every column that references a user id.
var SeedTables = []SeedTable{
	{"account_events", "actor_id", "setnull", false},
	{"account_events", "user_id", "cascade", false},
	{"activity_events", "actor_id", "cascade", false},
	{"api_credit_balances", "user_id", "cascade", false},
	{"api_keys", "user_id", "cascade", false},
	{"api_quota_counters", "user_id", "cascade", false},
	{"api_usage_daily", "user_id", "cascade", false},
	{"catalog_audit", "actor_id", "setnull", false},
	{"catalog_locks", "locked_by", "setnull", false},
	{"comments", "user_id", "cascade", false},
	{"community_honeypot_hits", "user_id", "cascade", false},
	{"contact_messages", "user_id", "setnull", false},
	{"contributor_stats", "user_id", "cascade", false},
	{"credit_ledger", "user_id", "cascade", false},
	{"data_exports", "user_id", "cascade", false},
	{"entity_match_reviews", "resolved_by", "setnull", false},
	{"follows", "followee_id", "cascade", false},
	{"follows", "follower_id", "cascade", false},
	{"item_status", "user_id", "cascade", false},
	{"lists", "user_id", "cascade", false},
	{"passkeys", "user_id", "cascade", false},
	{"pix_charges", "user_id", "cascade", false},
	{"ratings", "user_id", "cascade", false},
	{"reactions", "user_id", "cascade", false},
	{"reports", "reporter_id", "setnull", false},
	{"reports", "resolved_by", "setnull", false},
	{"reviews", "user_id", "cascade", false},
	{"score_bursts", "resolved_by", "setnull", false},
	{"social_accounts", "user_id", "cascade", false},
	{"subscriptions", "user_id", "cascade", false},
	{"suggestion_events", "actor_id", "setnull", false},
	{"suggestion_reviews", "reviewer_id", "cascade", false},
	{"suggestions", "assigned_to", "setnull", false},
	{"suggestions", "author_id", "setnull", false},
	{"suggestions", "decided_by", "setnull", false},
	{"time_to_beat", "user_id", "cascade", false},
	{"user_item_editions", "user_id", "cascade", false},
	{"user_notifications", "user_id", "cascade", false},
	{"user_recommendations", "user_id", "cascade", false},
	{"webhook_endpoints", "user_id", "cascade", false},
	{"access_tokens", "user_id", "loose", true},
	{"remember_tokens", "user_id", "loose", true},
	{"rating_events", "user_id", "loose", false},
	{"rating_flags", "user_id", "loose", false},
	{"rating_migration_audit", "user_id", "loose", false},
	{"score_actions", "user_id", "loose", false},
	{"search_events", "user_id", "loose", false},
}

// seedItemTables are the tables whose rows move an item's score or counters.
var seedItemTables = []string{"ratings", "item_status", "time_to_beat"}

// SeedMark is `seed:mark`: flags the fake legacy seed users (users.is_seed) so they can never sign in and
// `seed:purge` can remove them. A user must match every filter given; admins are never marked.
func SeedMark() vukem.Command {
	var suffix, prefix *string

	var dryRun *bool

	return vukem.CommandFunc{
		CommandName: "seed:mark",
		Desc:        "Flag fake seed users (email suffix and legacy id prefix) as is_seed",
		DefineFlags: func(fs *flag.FlagSet) {
			suffix = fs.String("email-suffix", "", "email suffix of seed users, for example @example.test")
			prefix = fs.String("legacy-prefix", "", "legacy_id prefix of seed users, for example seed-user-")
			dryRun = fs.Bool("dry-run", false, "print the counts, change nothing")
		},
		Fn: func(ctx context.Context, app *vukem.App, _ []string) error {
			if *suffix == "" && *prefix == "" {
				return errors.New("seed:mark: give --email-suffix, --legacy-prefix or both")
			}

			db, err := app.DB()
			if err != nil {
				return err
			}

			where, args := seedMatch(db, *suffix, *prefix)

			var total, already int64

			err = db.QueryRowContext(ctx, "SELECT count(*), count(*) FILTER (WHERE is_seed) FROM users WHERE "+where, args...).
				Scan(&total, &already)
			if err != nil {
				return err
			}

			var marked int64

			if !*dryRun {
				res, err := db.ExecContext(ctx, "UPDATE users SET is_seed = "+db.Dialect().Placeholder(len(args)+1)+
					" WHERE "+where+" AND NOT is_seed", append(args, true)...)
				if err != nil {
					return err
				}

				if marked, err = res.RowsAffected(); err != nil {
					return err
				}
			}

			_, _ = fmt.Fprintf(app.Stdout(), "Matching users: %d\nAlready seed: %d\nMarked: %d\nDry run: %t\n",
				total, already, marked, *dryRun)

			return nil
		},
	}
}

func seedMatch(db *database.DB, suffix, prefix string) (string, []any) {
	p := db.Dialect().Placeholder
	where, args := []string{"role <> 'admin'"}, []any{}

	if suffix != "" {
		args = append(args, "%"+likeEscape(strings.ToLower(suffix)))
		where = append(where, `lower(email) LIKE `+p(len(args))+` ESCAPE '\'`)
	}

	if prefix != "" {
		args = append(args, likeEscape(prefix)+"%")
		where = append(where, `legacy_id LIKE `+p(len(args))+` ESCAPE '\'`)
	}

	return strings.Join(where, " AND "), args
}

func likeEscape(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

// SeedPurge is `seed:purge`: deletes every is_seed user in batches, one transaction per batch, then queues the
// score recompute and the search refresh of the items those users touched.
func SeedPurge() vukem.Command {
	var dryRun, confirm *bool

	var batch *int

	return vukem.CommandFunc{
		CommandName: "seed:purge",
		Desc:        "Delete users flagged is_seed and recompute the items they touched",
		DefineFlags: func(fs *flag.FlagSet) {
			dryRun = fs.Bool("dry-run", false, "print the counts per table, delete nothing")
			confirm = fs.Bool("confirm", false, "required to delete")
			batch = fs.Int("batch", 500, "users deleted per transaction")
		},
		Fn: func(ctx context.Context, app *vukem.App, _ []string) error {
			if !*dryRun && !*confirm {
				return errors.New("seed:purge: refusing to delete without --confirm (use --dry-run to preview)")
			}

			if *batch < 1 || *batch > 5000 {
				return errors.New("seed:purge: --batch must be between 1 and 5000")
			}

			db, err := app.DB()
			if err != nil {
				return err
			}

			return seedPurge(ctx, app, db, *batch, *dryRun)
		},
	}
}

func seedPurge(ctx context.Context, app *vukem.App, db *database.DB, batch int, dryRun bool) error {
	out := app.Stdout()

	items, err := seedItems(ctx, db)
	if err != nil {
		return err
	}

	counts := map[string]int64{}
	users, last := int64(0), int64(0)

	for {
		ids, err := seedUserPage(ctx, db, last, batch)
		if err != nil {
			return err
		}

		if len(ids) == 0 {
			break
		}

		last = ids[len(ids)-1]
		users += int64(len(ids))

		err = db.Tx(ctx, func(tx *database.Tx) error {
			return purgeBatch(ctx, tx, ids, counts, dryRun)
		})
		if err != nil {
			return fmt.Errorf("seed:purge: batch ending at user %d: %w", last, err)
		}
	}

	_, _ = fmt.Fprintf(out, "Seed users: %d\nAffected items: %d\nDry run: %t\n", users, len(items), dryRun)

	for _, t := range SeedTables {
		key := t.Table + "." + t.Column
		_, _ = fmt.Fprintf(out, "%-36s %-8s %d\n", key, t.Kind, counts[key])
	}

	if dryRun || users == 0 {
		return nil
	}

	return seedRefresh(ctx, app, db, items)
}

// seedUserPage returns up to n seed user ids above after. Paging by id, not by "still flagged", keeps a dry run
// (which deletes nothing) from looping.
func seedUserPage(ctx context.Context, db *database.DB, after int64, n int) ([]int64, error) {
	p := db.Dialect().Placeholder

	rows, err := db.QueryContext(ctx, "SELECT id FROM users WHERE is_seed AND id > "+p(1)+" ORDER BY id LIMIT "+p(2),
		after, n)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var ids []int64

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, rows.Err()
}

// purgeBatch counts the rows that reference ids, then (unless dryRun) deletes the loose rows and the users; the
// FKs remove or null the rest. The users DELETE repeats is_seed so a flag cleared meanwhile is never deleted.
func purgeBatch(ctx context.Context, tx *database.Tx, ids []int64, counts map[string]int64, dryRun bool) error {
	p := tx.Dialect().Placeholder
	numeric := make([]any, len(ids))
	text := make([]any, len(ids))
	marks := make([]string, len(ids))

	for i, id := range ids {
		numeric[i], text[i], marks[i] = id, strconv.FormatInt(id, 10), p(i+1)
	}

	in := " IN (" + strings.Join(marks, ",") + ")"

	for _, t := range SeedTables {
		args := numeric
		if t.Text {
			args = text
		}

		var n int64
		if err := tx.QueryRowContext(ctx, "SELECT count(*) FROM "+t.Table+" WHERE "+t.Column+in, args...).
			Scan(&n); err != nil {
			return err
		}

		counts[t.Table+"."+t.Column] += n

		if t.Kind == "loose" && !dryRun {
			if _, err := tx.ExecContext(ctx, "DELETE FROM "+t.Table+" WHERE "+t.Column+in, args...); err != nil {
				return err
			}
		}
	}

	if dryRun {
		return nil
	}

	_, err := tx.ExecContext(ctx, "DELETE FROM users WHERE is_seed AND id"+in, numeric...)

	return err
}

// seedItems returns the ids of every item a seed user rated, set a status on or reported a time to beat for.
func seedItems(ctx context.Context, db *database.DB) ([]string, error) {
	parts := make([]string, len(seedItemTables))
	for i, t := range seedItemTables {
		parts[i] = "SELECT item_id::text AS id FROM " + t + " WHERE user_id IN (SELECT id FROM users WHERE is_seed)"
	}

	rows, err := db.QueryContext(ctx, strings.Join(parts, " UNION ")+" ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var ids []string

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, rows.Err()
}

// seedRefresh queues the score recompute of items and marks them dirty for the search suggest sync (the queue the
// catalog triggers write; the scheduled sync job drains it).
func seedRefresh(ctx context.Context, app *vukem.App, db *database.DB, items []string) error {
	q, err := app.Queue()
	if err != nil {
		return err
	}

	if err := scoring.EnqueueRecomputes(ctx, db, q, items); err != nil {
		return err
	}

	for _, id := range items {
		if err := search.MarkDirty(ctx, db, "item", id); err != nil {
			return err
		}
	}

	_, _ = fmt.Fprintln(app.Stdout(), "Queued score recompute and search refresh for the affected items")

	return nil
}
