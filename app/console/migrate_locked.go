// Package console holds the application's console commands.
package console

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/database"
)

// migrateLockSQL is the advisory lock id shared by every migration runner of
// the app; hashtext keeps the name readable in SQL.
const migrateLockSQL = `hashtext('hexascore:migrate')`

// WithMigrationLock runs fn while holding a session-level Postgres advisory
// lock, so a second runner blocks until the first finishes. vukem's runner
// takes no lock itself (VF-026). The lock lives on a dedicated connection
// that is kept for the whole run and closed afterwards, which also releases
// the lock if unlock fails or the process dies.
func WithMigrationLock(ctx context.Context, db *database.DB, fn func(context.Context) error) (err error) {
	if db.Dialect().Name() != "postgres" {
		return errors.New("migrate:locked needs PostgreSQL")
	}

	conn, err := db.SQL().Conn(ctx)
	if err != nil {
		return fmt.Errorf("migration lock: connection: %w", err)
	}

	defer func() { err = errors.Join(err, closeLocked(conn)) }()

	if _, err := conn.ExecContext(ctx, "SELECT pg_advisory_lock("+migrateLockSQL+")"); err != nil {
		return fmt.Errorf("migration lock: acquire: %w", err)
	}

	return fn(ctx)
}

// closeLocked unlocks and returns the connection to the pool. A lock still
// held after a failed unlock would stay with the pooled session, so callers
// treat the joined error as fatal for the run.
func closeLocked(conn *sql.Conn) error {
	_, unlockErr := conn.ExecContext(context.Background(), "SELECT pg_advisory_unlock("+migrateLockSQL+")")

	return errors.Join(unlockErr, conn.Close())
}

// MigrateLocked is `migrate` under the advisory lock: the pre-deploy command
// (SDD 11 §7.2). Output matches `migrate`.
func MigrateLocked() vukem.Command {
	return vukem.CommandFunc{
		CommandName: "migrate:locked",
		Desc:        "Run pending migrations under a Postgres advisory lock",
		Fn: func(ctx context.Context, app *vukem.App, _ []string) error {
			db, err := app.DB()
			if err != nil {
				return err
			}

			return WithMigrationLock(ctx, db, func(ctx context.Context) error {
				m, err := app.Migrator()
				if err != nil {
					return err
				}

				ran, err := m.Up(ctx)
				for _, id := range ran {
					_, _ = fmt.Fprintln(app.Stdout(), "Migrated:", id)
				}

				if err == nil && len(ran) == 0 {
					_, _ = fmt.Fprintln(app.Stdout(), "Nothing to migrate.")
				}

				return err
			})
		},
	}
}
