package console

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/config"
	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/app/scoring"
)

// BackupDrill is `backup:drill`: restore the latest daily dump of the source
// bucket into a scratch database next to DATABASE_URL, compare row counts
// with the dump manifest, run the data quality asserts, drop it and record
// backup_runs(kind='drill') (SDD 11 §7.4, FR-11-6). The source bucket and
// key come from BACKUP_SOURCE_<name> (BACKUP_SOURCE_S3_*, BACKUP_SOURCE_BACKUP_KEY)
// and fall back to the plain names, so staging drills production's dumps
// once the owner grants it read access. --first-sunday makes the weekly
// schedule entry a monthly one.
func BackupDrill() vukem.Command {
	var firstSunday *bool

	return vukem.CommandFunc{
		CommandName: "backup:drill",
		Desc:        "Restore the latest daily dump into a scratch database and check it",
		DefineFlags: func(fs *flag.FlagSet) {
			firstSunday = fs.Bool("first-sunday", false, "run only on the first Sunday of the month (scheduler)")
		},
		Fn: func(ctx context.Context, app *vukem.App, _ []string) error {
			if *firstSunday && app.Now().UTC().Day() > 7 {
				return nil
			}

			r, err := backupRunner(app, sourcePrefixed{app.Source})
			if err != nil {
				return err
			}

			res, err := r.Drill(ctx, app.Source.Get("DATABASE_URL"), drillAsserts)
			if res.Key != "" {
				_, _ = fmt.Fprintf(app.Stdout(), "Drill of %s: %s, %d assert errors\n",
					res.Key, res.Duration.Round(time.Second), res.AssertErrors)
			}

			for _, d := range res.Diff {
				_, _ = fmt.Fprintln(app.Stdout(), "  count mismatch "+d)
			}

			return err
		},
	}
}

// drillAsserts runs the asserts that exist (SDD 01 §7.6 plus scoring); the
// SDD 06 asserts join in P6.
func drillAsserts(ctx context.Context, db *database.DB) (int, error) {
	report, err := catalog.RunDQ(ctx, db, scoring.DQAsserts()...)

	return report.Errors, err
}

// sourcePrefixed reads BACKUP_SOURCE_<key> first, then <key>.
type sourcePrefixed struct{ src config.Lookuper }

func (s sourcePrefixed) Lookup(key string) (string, bool) {
	if v, ok := s.src.Lookup("BACKUP_SOURCE_" + strings.TrimPrefix(key, "BACKUP_SOURCE_")); ok && v != "" {
		return v, true
	}

	return s.src.Lookup(key)
}
