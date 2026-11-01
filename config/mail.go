package config

import "cmp"

// Sender and support addresses (owner decision Q-A17, 2026-11-01): noreply@
// sends every automatic message; support@ is the only address that receives
// and answers user email. privacy@hexascore.app (D-39) is a legal-page contact,
// not part of this config.
const (
	DefaultMailFrom       = "HexaScore <noreply@hexascore.app>"
	DefaultSupportAddress = "support@hexascore.app"
)

// MailFrom reads MAIL_FROM (the vukem default sender setting).
func MailFrom(get func(string) string) string {
	return cmp.Or(get("MAIL_FROM"), DefaultMailFrom)
}

// SupportAddress reads MAIL_SUPPORT_ADDRESS: the Reply-To of automatic mail.
func SupportAddress(get func(string) string) string {
	return cmp.Or(get("MAIL_SUPPORT_ADDRESS"), DefaultSupportAddress)
}

// ContactInbox reads CONTACT_INBOX, defaulting to the support address.
func ContactInbox(get func(string) string) string {
	return cmp.Or(get("CONTACT_INBOX"), SupportAddress(get))
}
