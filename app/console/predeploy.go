package console

import (
	"context"
	"errors"
	"fmt"
	"io"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/schema"
)

// invalidIndexesSQL lists the INVALID indexes of the current schema that no
// backend is building right now: leftovers of a CREATE INDEX CONCURRENTLY
// that failed. A re-run of `CREATE INDEX CONCURRENTLY IF NOT EXISTS` skips
// such a leftover and would record the migration over an index the planner
// never uses. Partitioned parents (relkind 'I') are invalid by design until
// every partition has its index, so they are left alone.
const invalidIndexesSQL = `SELECT c.relname, format('%I.%I', n.nspname, c.relname)
FROM pg_index i
JOIN pg_class c ON c.oid = i.indexrelid
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE NOT i.indisvalid AND c.relkind = 'i' AND n.nspname = current_schema()
  AND NOT EXISTS (SELECT 1 FROM pg_stat_progress_create_index p WHERE p.index_relid = i.indexrelid)
ORDER BY c.relname`

// DropInvalidIndexes drops the INVALID leftovers of failed concurrent index
// builds (see invalidIndexesSQL) and returns their names. Each DROP runs
// outside a transaction on a pooled connection that holds no other lock.
func DropInvalidIndexes(ctx context.Context, db *database.DB) ([]string, error) {
	rows, err := db.SQL().QueryContext(ctx, invalidIndexesSQL)
	if err != nil {
		return nil, fmt.Errorf("invalid indexes: %w", err)
	}

	defer func() { _ = rows.Close() }()

	type index struct{ name, ident string }

	var found []index

	for rows.Next() {
		var ix index
		if err := rows.Scan(&ix.name, &ix.ident); err != nil {
			return nil, fmt.Errorf("invalid indexes: %w", err)
		}

		found = append(found, ix)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("invalid indexes: %w", err)
	}

	dropped := make([]string, 0, len(found))

	for _, ix := range found {
		if _, err := db.SQL().ExecContext(ctx, "DROP INDEX CONCURRENTLY IF EXISTS "+ix.ident); err != nil {
			return dropped, fmt.Errorf("drop invalid index %s: %w", ix.name, err)
		}

		dropped = append(dropped, ix.name)
	}

	return dropped, nil
}

// RunPredeploy drops INVALID index leftovers, then runs the pending
// migrations. vukem's runner holds its own advisory lock for the whole run
// (VF-026) and runs `NoTransaction` migrations outside a transaction
// (VF-017), so a concurrent index is a one-statement `NoTransaction`
// migration with `CREATE INDEX CONCURRENTLY IF NOT EXISTS`.
func RunPredeploy(ctx context.Context, db *database.DB, m *schema.Migrator, out io.Writer) error {
	if db.Dialect().Name() != "postgres" {
		return errors.New("hexascore:predeploy needs PostgreSQL")
	}

	dropped, err := DropInvalidIndexes(ctx, db)
	for _, name := range dropped {
		_, _ = fmt.Fprintln(out, "Dropped INVALID index:", name)
	}

	if err != nil {
		return err
	}

	ran, err := m.Up(ctx)
	for _, id := range ran {
		_, _ = fmt.Fprintln(out, "Migrated:", id)
	}

	if err == nil && len(ran) == 0 {
		_, _ = fmt.Fprintln(out, "Nothing to migrate.")
	}

	return err
}

// Predeploy is the Railway pre-deploy command (SDD 11 §7.2); a failure stops
// the deployment and the previous one keeps serving.
func Predeploy() vukem.Command {
	return vukem.CommandFunc{
		CommandName: "hexascore:predeploy",
		Desc:        "Drop INVALID index leftovers, then run pending migrations (pre-deploy)",
		Fn: func(ctx context.Context, app *vukem.App, _ []string) error {
			db, err := app.DB()
			if err != nil {
				return err
			}

			m, err := app.Migrator()
			if err != nil {
				return err
			}

			return RunPredeploy(ctx, db, m, app.Stdout())
		},
	}
}
