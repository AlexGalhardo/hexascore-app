package console

import (
	"context"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"slices"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/config"
	"github.com/AlexGalhardo/vukem/crypt"
	"github.com/AlexGalhardo/vukem/drivers/s3"

	"github.com/AlexGalhardo/hexascore-app/app/backup"
)

// BackupRun is `backup:run --kind daily|monthly`: one encrypted pg_dump in
// the bucket, recorded in backup_runs (SDD 11 §7.3). BACKUP_KEY (32 bytes,
// base64) is separate from APP_KEY so a leaked app key does not open dumps.
func BackupRun() vukem.Command {
	var kind *string

	return vukem.CommandFunc{
		CommandName: "backup:run",
		Desc:        "Dump, encrypt and upload the database to backups/postgres/<kind>/",
		DefineFlags: func(fs *flag.FlagSet) { kind = fs.String("kind", "daily", "daily or monthly") },
		Fn: func(ctx context.Context, app *vukem.App, _ []string) error {
			if !slices.Contains(backup.Kinds, *kind) {
				return fmt.Errorf("backup:run: --kind must be one of %v", backup.Kinds)
			}

			enc, err := backupEncrypter(app.Source.Get("BACKUP_KEY"))
			if err != nil {
				return err
			}

			cfg, err := config.Load[s3.Config](app.Source)
			if err != nil {
				return fmt.Errorf("backup:run: S3 config: %w", err)
			}

			disk, err := s3.New(cfg)
			if err != nil {
				return fmt.Errorf("backup:run: S3 disk: %w", err)
			}

			db, err := app.DB()
			if err != nil {
				return err
			}

			r := &backup.Runner{
				DB: db, Disk: disk, Enc: enc, DatabaseURL: app.Source.Get("DATABASE_URL"),
				PgDump: "pg_dump", Env: string(app.Config.Env), Now: app.Now,
			}

			res, err := r.Run(ctx, *kind)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(app.Stdout(), "Uploaded %s (%d bytes, sha256 %s)\n", res.Key, res.Bytes, res.SHA256)

			return nil
		},
	}
}

func backupEncrypter(b64 string) (*crypt.Encrypter, error) {
	key, err := base64.StdEncoding.DecodeString(b64)
	if err != nil || len(key) != 32 {
		return nil, errors.New("backup:run: BACKUP_KEY must be 32 bytes, base64")
	}

	return crypt.New(key)
}
