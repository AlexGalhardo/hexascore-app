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

// SupportAddress is the configured support address, the default Reply-To of
// automatic mail (bootstrap passes it to the vukem mailer's SetReplyTo).
func SupportAddress() string {
	a := addrs.Load()
	if a == nil {
		Configure("", "")
		a = addrs.Load()
	}
	return a.support
}

// Auto stamps the sender on an automatic message: the HexaScore default
// differs from vukem's MAIL_FROM default, so an unset MAIL_FROM still sends
// from noreply@. The Reply-To comes from the mailer (MAIL_REPLY_TO, else
// MAIL_SUPPORT_ADDRESS).
func Auto(e mail.Envelope) mail.Envelope {
	a := addrs.Load()
	if a == nil {
		Configure("", "")
		a = addrs.Load()
	}
	if e.From == "" {
		e.From = a.from
	}
	return e
}
