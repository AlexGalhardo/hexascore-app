package console

import (
	"context"
	"fmt"
	"strings"
	"time"

	vukem "github.com/AlexGalhardo/vukem"

	"github.com/AlexGalhardo/hexascore-app/app/backup"
	hmail "github.com/AlexGalhardo/hexascore-app/app/mail"
)

// BackupOffsite is `backup:offsite`: mail ADMIN_EMAILS the 7-day signed link to this month's encrypted
// monthly dump (SDD 11 §7.3 step 6, D-32). Scheduled on day 1 after backup.daily has made the monthly copy.
func BackupOffsite() vukem.Command {
	return vukem.CommandFunc{
		CommandName: "backup:offsite",
		Desc:        "Mail ADMIN_EMAILS a 7-day signed link to this month's encrypted monthly dump",
		Fn: func(ctx context.Context, app *vukem.App, _ []string) error {
			r, err := backupRunner(app, app.Source)
			if err != nil {
				return err
			}

			signer, err := backup.NewOffsiteSigner(app)
			if err != nil {
				return err
			}

			mailer, err := app.Mailer()
			if err != nil {
				return err
			}

			o := &backup.Offsite{
				DB: r.DB, Disk: r.Disk, Signer: signer, URL: app.Config.URL,
				To: adminEmails(app.Source.Get("ADMIN_EMAILS")), Now: app.Now,
				Send: func(ctx context.Context, m backup.OffsiteMail) error {
					return mailer.Send(ctx, hmail.BackupOffsite{
						To: m.To, Link: m.Link, Key: m.Key, SHA256: m.SHA256, Bytes: m.Bytes,
						Expires: m.Expires.Format(time.RFC3339),
					})
				},
			}

			if err := o.Mail(ctx); err != nil {
				return err
			}

			_, _ = fmt.Fprintln(app.Stdout(), "Off-site link mailed (or already mailed this month)")

			return nil
		},
	}
}

func adminEmails(raw string) []string {
	var out []string

	for e := range strings.SplitSeq(raw, ",") {
		if e = strings.TrimSpace(e); e != "" {
			out = append(out, e)
		}
	}

	return out
}
