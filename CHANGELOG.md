# Changelog

All notable changes to this project are documented here.
Format: [Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/). Versioning: [SemVer](https://semver.org/).

## [Unreleased]

### Added

- `PLAN.md` with decisions and milestones; reference docs in `docs/` (architecture, data model, API, frontend, design system, payments, ingestion, security, testing, process).
- `AGENTS.md` and `CLAUDE.md` agent guides.
- Auth & account (0.2): `/login` (email or nickname, Google when configured, passkey, TOTP/backup-code 2FA step, resend verification), `/signup` (live nickname avatar, password checklist), `/verify-email`, `/recover-password`, `/reset-password`, `/profile` (name, set/change password, 2FA with QR + backup codes, passkeys, private profile, account deletion with 10-second confirm and 30-day grace) and a cancel-deletion banner after login. All in EN, pt-BR and ES.
- Server Actions in `apps/web/actions/` (`account.ts`, `contact.ts`) returning typed `{ ok } | { ok: false, error }` results.
- `/contact` (session identity overrides the form, DB rate limit per user/IP, optional email to `CONTACT_INBOX`), `/policy` and `/terms` in 3 languages.
- Daily Vercel cron `/api/cron/purge-accounts` (Bearer `CRON_SECRET`, constant-time) that anonymizes accounts 30 days after scheduled deletion, keeping their ratings.
- Vercel BotID on signup, login and contact (`botid` 1.5.11), bypassed outside Vercel. `qrcode` 1.5.4 renders the TOTP QR in the browser.
- Navbar user menu (Profile, API, Log out) and Login/Sign up links.

### Fixed

- `proxy.ts` matcher had an unescaped dot, so no page outside `/pt-br`/`/es` was localized (`/login` returned 404).
- PGlite is now a server external package and `next build` workers use an in-memory instance, so dev/build no longer crash or corrupt `.pglite`.
- Dev email log printed `link=(none)`: templates are now passed with `createElement`.

### Removed

- Files inherited from the previous project (RespondeAê): old plan, docs, setup scripts, tests, REST client samples, graphify output, `vercel.json`.
