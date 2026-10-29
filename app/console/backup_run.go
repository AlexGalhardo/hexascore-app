package console

import (
	"context"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"slices"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/crypt"

	"github.com/AlexGalhardo/hexascore-app/app/backup"
)

// BackupRun is `backup:run --kind daily|monthly`: one encrypted pg_dump in
// the bucket, recorded in backup_runs, verified, copied to monthly on day 1
// and followed by the retention prune (SDD 11 §7.3). BACKUP_KEY (32 bytes,
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

			r, err := backupRunner(app, app.Source)
			if err != nil {
				return err
			}

			res, err := r.Run(ctx, *kind)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(app.Stdout(), "Uploaded %s (%d bytes, sha256 %s)\n", res.Key, res.Bytes, res.SHA256)

			if err := r.Verify(ctx, *kind, res); err != nil {
				return err
			}

			_, _ = fmt.Fprintln(app.Stdout(), "Verified")

			// Day 1 (UTC; the 06:00 America/Sao_Paulo run is 09:00 UTC, same date): keep a monthly copy.
			if *kind == "daily" && r.Now().UTC().Day() == 1 {
				m, err := r.CopyMonthly(ctx, res)
				if err != nil {
					return err
				}

				_, _ = fmt.Fprintf(app.Stdout(), "Copied to %s\n", m.Key)
			}

			n, err := r.Prune(ctx)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(app.Stdout(), "Pruned %d old dumps\n", n)

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
