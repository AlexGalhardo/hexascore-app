# CONFIG-email-addresses (senior-b)

Branch `agent/senior/CONFIG-email-addresses` @ 9477d2f1, worktree `C:\dev\hexascore-app\wt\senior-b-CONFIG-email-addresses`. Owner decision Q-A17 (2026-11-01), confirmed "Sim, aplicar".

## Changes

- `config/mail.go`: `MailFrom` (`MAIL_FROM`, default `HexaScore <noreply@hexascore.app>`), `SupportAddress` (`MAIL_SUPPORT_ADDRESS`, default `support@hexascore.app`), `ContactInbox` (`CONTACT_INBOX`, falls back to the support address).
- `app/mail/addresses.go`: `Configure` (called in `bootstrap/app.go`) and `Auto`, which stamps From and a support Reply-To on every automatic envelope (all `app/mail` mailables plus the webhook-disabled notice). The contact notification keeps the visitor as Reply-To (support answers the visitor).
- `routes/reports.go`: contact recipient via `config.ContactInbox`; unset `CONTACT_INBOX` no longer disables mailing.
- `resources/lang/{en,pt-BR,es}.json`: contact page body shows `support@hexascore.app`.
- `.env.example` (worktree only): `MAIL_FROM`, `MAIL_SUPPORT_ADDRESS`, `CONTACT_INBOX` comment.
- `vukem-feedback.md`: VF-NEW-MAIL-REPLYTO (vukem has `MAIL_FROM` but no global Reply-To; workaround is `Auto`).
- `privacy@hexascore.app` (D-39) untouched. Owner to confirm whether privacy requests also go to support.

## Evidence

- Red: `go test ./config ./app/mail` failed to build (undefined `config.MailFrom`, `Configure`, ...).
- Green: `go test ./config ./app/mail` ok; `go test ./tests -run TestAutomaticMailFromAndReplyTo` PASS (real verification mail: From noreply@, Reply-To support@); `TestMailLocaleFromPayload`, `TestContactHoneypotHTTP` ok.
- `TestOnlyDocumentedAddresses` scans lang, js, mobile, app, routes: only privacy@, support@, noreply@.
- `golangci-lint ... ./config ./app/... ./routes ./bootstrap ./tests`: 0 issues.

## DNS (read-only, 2026-11-01)

Nameservers: Vercel DNS. Apex MX: none. `send` MX: feedback-smtp.us-east-1.amazonses.com. Apex SPF: none (only google-site-verification TXT). `send` SPF: `v=spf1 include:amazonses.com ~all`. DKIM `resend._domainkey`: present. DMARC `_dmarc`: none.
Owner to-do: mailbox or forwarder for support@ and privacy@ plus apex MX; apex SPF for that provider; DMARC TXT; confirm the Resend domain shows Verified; set `MAIL_FROM`/`MAIL_SUPPORT_ADDRESS` in production (defaults already match).

## Not done

Footer/help/About/legal pages show no address: they link to the contact form, which now states support@. No AGENTS.md or CHANGELOG edit (CHANGELOG line for the specialist: "Added: default From noreply@, Reply-To and contact inbox support@hexascore.app via MAIL_FROM / MAIL_SUPPORT_ADDRESS / CONTACT_INBOX").

## Review (especialista, 2026-11-01)

verdict: APROVADO (merge of main resolved a `vukem-feedback.md` conflict, both sides kept; `routes/reports.go` merged cleanly with the P7 social service)
re-run: `go test ./config ./app/mail ./tests/i18n` ok; `go test -tags integration ./tests -run 'TestAutomaticMailFromAndReplyTo|TestOnlyDocumentedAddresses|TestMail|TestContact|TestReports|TestWebhook'` ok; golangci-lint (`-j 2`, config, app/mail, app/webhooks, routes, bootstrap, tests) 0 issues; `golangci-lint fmt --diff ./...` clean
checked: Q-A17 (noreply@ sender, support@ Reply-To and contact inbox); every `mail.Mailable` envelope in `app/mail` goes through `Auto`, the contact notice keeps the visitor as Reply-To; an unset `CONTACT_INBOX` now mails support@ instead of disabling mail (owner-approved); no secret or address outside the three documented ones (`TestOnlyDocumentedAddresses`); i18n parity
open for the owner: DNS to-do in the delivery report (support@ and privacy@ mailboxes, apex MX and SPF, DMARC); `VF-NEW-MAIL-REPLYTO` keeps its provisional id for the orchestrator
PRONTO_PARA_MERGE agent/especialista/CONFIG-email-addresses (merged 37b8770b)
