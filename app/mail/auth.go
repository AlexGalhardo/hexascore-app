package mail

import (
	"cmp"
	"time"

	"github.com/AlexGalhardo/vukem/mail"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// VerifyEmail asks a new user to confirm their address.
type VerifyEmail struct {
	User models.User
	Link string
	// Subject is translated by the sender (accounts.T), since mailables
	// have no request context; empty means English.
	Subject string
}

// Envelope implements mail.Mailable.
func (m VerifyEmail) Envelope() mail.Envelope {
	return mail.Envelope{To: []string{m.User.Email}, Subject: cmp.Or(m.Subject, "Verify your email address")}
}

// Content implements mail.Mailable.
func (m VerifyEmail) Content() mail.Content { return mail.View("mail/verify_email", m) }

// ResetPassword carries a password reset link.
type ResetPassword struct {
	User    models.User
	Link    string
	Subject string
}

// Envelope implements mail.Mailable.
func (m ResetPassword) Envelope() mail.Envelope {
	return mail.Envelope{To: []string{m.User.Email}, Subject: cmp.Or(m.Subject, "Reset your password")}
}

// Content implements mail.Mailable.
func (m ResetPassword) Content() mail.Content { return mail.View("mail/reset_password", m) }

// AccountExists tells the owner of an address that someone tried to sign
// up with it: the sign-up form answers the same for every address, so this
// mail is the only place that says the account exists.
type AccountExists struct {
	User      models.User
	LoginLink string
	ResetLink string
	Subject   string
}

// Envelope implements mail.Mailable.
func (m AccountExists) Envelope() mail.Envelope {
	return mail.Envelope{
		To:      []string{m.User.Email},
		Subject: cmp.Or(m.Subject, "Someone tried to sign up with your email"),
	}
}

// Content implements mail.Mailable.
func (m AccountExists) Content() mail.Content { return mail.View("mail/account_exists", m) }

// TwoFactorReset tells a user that an administrator cleared their
// two-factor authentication after a manual identity check.
type TwoFactorReset struct {
	User      models.User
	ResetLink string
	Subject   string
}

// Envelope implements mail.Mailable.
func (m TwoFactorReset) Envelope() mail.Envelope {
	return mail.Envelope{
		To:      []string{m.User.Email},
		Subject: cmp.Or(m.Subject, "Two-factor authentication was reset"),
	}
}

// Content implements mail.Mailable.
func (m TwoFactorReset) Content() mail.Content { return mail.View("mail/two_factor_reset", m) }

// DeletionScheduled confirms a scheduled deletion and says how to cancel it.
type DeletionScheduled struct {
	User       models.User
	Date       time.Time
	CancelLink string
	Subject    string
}

// Envelope implements mail.Mailable.
func (m DeletionScheduled) Envelope() mail.Envelope {
	return mail.Envelope{
		To:      []string{m.User.Email},
		Subject: cmp.Or(m.Subject, "Your account is scheduled for deletion"),
	}
}

// Content implements mail.Mailable.
func (m DeletionScheduled) Content() mail.Content { return mail.View("mail/deletion_scheduled", m) }

// AccountAnonymized tells the former owner that the deletion took effect;
// User still holds the address and name from before the purge.
type AccountAnonymized struct {
	User    models.User
	Subject string
}

// Envelope implements mail.Mailable.
func (m AccountAnonymized) Envelope() mail.Envelope {
	return mail.Envelope{To: []string{m.User.Email}, Subject: cmp.Or(m.Subject, "Your account was deleted")}
}

// Content implements mail.Mailable.
func (m AccountAnonymized) Content() mail.Content { return mail.View("mail/account_anonymized", m) }
