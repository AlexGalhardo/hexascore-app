package console

import (
	"context"
	"errors"
	"flag"
	"fmt"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/config"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres"
	"github.com/AlexGalhardo/vukem/drivers/s3"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/backup"
)

// BackupRestore is `backup:restore <object key> --target-url <url>`: decrypt
// one dump into an empty database, then re-apply the account deletions the
// running database recorded after the dump (SDD 11 §7.5 step 3, SDD 13 §8).
// The replay source defaults to DATABASE_URL; --no-replay is the explicit
// choice when that database is gone, and the command says what it skipped.
func BackupRestore() vukem.Command {
	var target, replayFrom *string

	var noReplay *bool

	return vukem.CommandFunc{
		CommandName: "backup:restore",
		Desc:        "Restore an encrypted dump into --target-url and replay newer account deletions",
		DefineFlags: func(fs *flag.FlagSet) {
			target = fs.String("target-url", "", "postgres:// URL of the empty database to restore into (required)")
			replayFrom = fs.String("replay-from", "", "database to replay deletions from (default DATABASE_URL)")
			noReplay = fs.Bool("no-replay", false, "skip the deletion replay (the source database is gone)")
		},
		Fn: func(ctx context.Context, app *vukem.App, args []string) error {
			if len(args) != 1 || *target == "" {
				return errors.New("backup:restore: usage: backup:restore <object key> --target-url <url>")
			}

			r, err := backupRunner(app, app.Source)
			if err != nil {
				return err
			}

			if err := r.Restore(ctx, args[0], *target); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(app.Stdout(), "Restored %s\n", args[0])

			if *noReplay {
				_, _ = fmt.Fprintln(
					app.Stdout(),
					"Deletion replay skipped (--no-replay): re-apply deletions newer than the dump by hand",
				)

				return nil
			}

			stats, err := replayDeletions(ctx, app, *replayFrom, *target)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(app.Stdout(), "Replayed deletions: %d anonymised, %d scheduled, %d cancelled\n",
				stats.Anonymised, stats.Scheduled, stats.Cancelled)

			return nil
		},
	}
}

func replayDeletions(ctx context.Context, app *vukem.App, sourceURL, targetURL string) (accounts.ReplayStats, error) {
	acc, err := accounts.New(app)
	if err != nil {
		return accounts.ReplayStats{}, err
	}

	source := acc.DB
	if sourceURL != "" {
		if source, err = database.Open(ctx, sourceURL, database.WithDriver(postgres.Driver())); err != nil {
			return accounts.ReplayStats{}, fmt.Errorf("backup:restore: replay source: %w", err)
		}
		defer func() { _ = source.Close() }()
	}

	target, err := database.Open(ctx, targetURL, database.WithDriver(postgres.Driver()))
	if err != nil {
		return accounts.ReplayStats{}, fmt.Errorf("backup:restore: target: %w", err)
	}
	defer func() { _ = target.Close() }()

	return acc.ReplayDeletions(ctx, source, target)
}

// backupRunner builds the runner from src: S3_* for the disk, BACKUP_KEY for
// the frames, DATABASE_URL for the run records.
func backupRunner(app *vukem.App, src config.Lookuper) (*backup.Runner, error) {
	key, _ := src.Lookup("BACKUP_KEY")

	enc, err := backupEncrypter(key)
	if err != nil {
		return nil, err
	}

	cfg, err := config.Load[s3.Config](src)
	if err != nil {
		return nil, fmt.Errorf("backup: S3 config: %w", err)
	}

	disk, err := s3.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("backup: S3 disk: %w", err)
	}

	db, err := app.DB()
	if err != nil {
		return nil, err
	}

	return &backup.Runner{
		DB: db, Disk: disk, Enc: enc, DatabaseURL: app.Source.Get("DATABASE_URL"),
		PgDump: "pg_dump", PgRestore: "pg_restore", Env: string(app.Config.Env), Now: app.Now,
	}, nil
}
