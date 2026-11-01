package mail

import (
	"github.com/AlexGalhardo/vukem/mail"
)

// BackupOffsite gives the owner the 7-day signed link to the encrypted monthly dump (SDD 11 §7.3 step 6).
// Owner-only and English: it carries no BACKUP_KEY and no other secret than the expiring link.
type BackupOffsite struct {
	To                         []string
	Link, Key, SHA256, Expires string
	Bytes                      int64
}

// Envelope implements mail.Mailable.
func (m BackupOffsite) Envelope() mail.Envelope {
	return Auto(mail.Envelope{To: m.To, Subject: "HexaScore monthly backup: off-site copy"})
}

// Content implements mail.Mailable.
func (m BackupOffsite) Content() mail.Content { return mail.View("mail/backup_offsite", m) }
