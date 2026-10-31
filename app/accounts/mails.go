package accounts

import (
	"context"

	vmail "github.com/AlexGalhardo/vukem/mail"
	"github.com/AlexGalhardo/vukem/orm"

	"github.com/AlexGalhardo/hexascore-app/app/mail"
	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// userLocale is the locale saved on the account; mail is queued and rendered outside the request, so the
// request locale is not used (SDD 03 §6.5, VF-013).
func (a *Accounts) userLocale(ctx context.Context, userID int64) string {
	var locale string
	err := a.DB.QueryRowContext(ctx, "SELECT locale FROM users WHERE id = "+a.DB.Dialect().Placeholder(1), userID).
		Scan(&locale)
	if err != nil || locale == "" {
		return "en"
	}

	return locale
}

// mailStrings translates the subject and the listed fields of auth.mail.<name> for the user's saved locale.
func (a *Accounts) mailStrings(
	ctx context.Context, u models.User, name string, fields []string, extra ...any,
) mail.Strings {
	locale := a.userLocale(ctx, u.ID)
	text := make(map[string]string, len(fields))
	for _, f := range fields {
		text[f] = a.tIn(ctx, locale, "auth.mail."+name+"."+f, append([]any{"name", u.Name}, extra...)...)
	}

	return mail.Strings{Locale: locale, Subject: a.tIn(ctx, locale, "auth.mail."+name+"_subject"), Text: text}
}

// notify queues a security notice; the action it reports is already committed, so a failure is logged.
func (a *Accounts) notify(ctx context.Context, m vmail.Mailable, userID int64, what string) {
	if err := a.Mailer.SendLater(ctx, m); err != nil {
		a.log.ErrorContext(ctx, "queue account mail", "mail", what, "user", userID, "err", err)
	}
}

// noticeFields are the template strings of the password and two-factor notices.
var (
	noticeFields = []string{"greeting", "body", "action", "footer"}
	bodyFields   = []string{"greeting", "body"}
)

func (a *Accounts) sendPasswordChanged(ctx context.Context, userID int64) {
	u, err := orm.QueryOf[models.User](a.DB).Where("id", userID).First(ctx)
	if err != nil {
		a.log.ErrorContext(ctx, "load user for mail", "user", userID, "err", err)
		return
	}
	a.notify(ctx, mail.PasswordChanged{
		User: u, ResetLink: a.URL + "/recover-password",
		Strings: a.mailStrings(ctx, u, "password_changed", noticeFields),
	}, userID, "password_changed")
}

// sendPasskeyChanged reports a committed add or removal. event is "added" or "removed".
func (a *Accounts) sendPasskeyChanged(ctx context.Context, userID int64, event string) {
	u, err := orm.QueryOf[models.User](a.DB).Where("id", userID).First(ctx)
	if err != nil {
		a.log.ErrorContext(ctx, "load user for mail", "user", userID, "err", err)
		return
	}
	a.notify(ctx, mail.PasskeyChanged{
		User: u, ResetLink: a.URL + "/recover-password",
		Strings: a.mailStrings(ctx, u, "passkey_"+event, noticeFields),
	}, userID, "passkey_"+event)
}

// sendTwoFactorChanged reports event: enabled, disabled or recovery_used.
func (a *Accounts) sendTwoFactorChanged(ctx context.Context, userID int64, event string) {
	u, err := orm.QueryOf[models.User](a.DB).Where("id", userID).First(ctx)
	if err != nil {
		a.log.ErrorContext(ctx, "load user for mail", "user", userID, "err", err)
		return
	}
	a.notify(ctx, mail.TwoFactorChanged{
		User: u, ResetLink: a.URL + "/recover-password",
		Strings: a.mailStrings(ctx, u, "two_factor_"+event, noticeFields),
	}, userID, "two_factor_"+event)
}

// SendExportReady queues the mail with the signed export link, in the user's saved locale (SDD 14 §7.4).
func (a *Accounts) SendExportReady(ctx context.Context, userID int64, link, expires string) error {
	u, err := orm.QueryOf[models.User](a.DB).Where("id", userID).First(ctx)
	if err != nil {
		return err
	}

	return a.Mailer.SendLater(ctx, mail.ExportReady{
		User: u, Link: link, Strings: a.mailStrings(ctx, u, "export_ready", noticeFields, "expires", expires),
	})
}
