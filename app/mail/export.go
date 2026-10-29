package mail

import (
	"cmp"

	"github.com/AlexGalhardo/vukem/mail"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// ExportReady sends the signed, 24-hour download link of a data export (SDD 14 §7.4).
type ExportReady struct {
	User models.User
	Link string
	Strings
}

// Envelope implements mail.Mailable.
func (m ExportReady) Envelope() mail.Envelope {
	return mail.Envelope{To: []string{m.User.Email}, Subject: cmp.Or(m.Subject, "Your HexaScore data export is ready")}
}

// Content implements mail.Mailable.
func (m ExportReady) Content() mail.Content { return mail.View("mail/export_ready", m) }
