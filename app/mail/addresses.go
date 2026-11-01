package mail

import (
	"sync/atomic"

	"github.com/AlexGalhardo/vukem/mail"

	"github.com/AlexGalhardo/hexascore-app/config"
)

type addresses struct{ from, support string }

var addrs atomic.Pointer[addresses]

// Configure sets the sender and support addresses (empty values use the
// config defaults). Call it once at boot with the MAIL_FROM and
// MAIL_SUPPORT_ADDRESS values.
func Configure(from, support string) {
	addrs.Store(&addresses{
		from:    config.MailFrom(func(string) string { return from }),
		support: config.SupportAddress(func(string) string { return support }),
	})
}

// Auto stamps the sender and, unless already set, the support Reply-To on an
// automatic message. vukem has no global default Reply-To (see vukem-feedback.md).
func Auto(e mail.Envelope) mail.Envelope {
	a := addrs.Load()
	if a == nil {
		Configure("", "")
		a = addrs.Load()
	}
	if e.From == "" {
		e.From = a.from
	}
	if len(e.ReplyTo) == 0 {
		e.ReplyTo = []string{a.support}
	}
	return e
}
