package mail

import (
	"cmp"

	"github.com/AlexGalhardo/vukem/mail"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// SuggestionNotice tells an author that their suggestion was decided, reverted or expired (SDD 04 FR-04-14).
type SuggestionNotice struct {
	User models.User
	Link string
	Strings
}

// Envelope implements mail.Mailable.
func (m SuggestionNotice) Envelope() mail.Envelope {
	return mail.Envelope{
		To:      []string{m.User.Email},
		Subject: cmp.Or(m.Subject, "Your HexaScore suggestion was updated"),
	}
}

// Content implements mail.Mailable.
func (m SuggestionNotice) Content() mail.Content { return mail.View("mail/suggestion_notice", m) }
