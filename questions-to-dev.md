# Questions to the developer

Doubts and decisions the orchestrator could not take safely while the owner was away (from 2026-10-29). Work continues around each item; nothing here blocks the rest of the roadmap. Answer inline under each item.

## Decisions taken on your behalf (blanket approval, 2026-10-29)

- Q-001: `golang.org/x/crypto v0.58.0` approved as a direct dependency (scrypt verifier for Better Auth hashes, ADR 0008).
- FQ-001 (VF-001 passkeys): option (a) chosen, passkeys stay post-1.0; no new framework dependency. Change it here if you prefer (b) stdlib or (c) `contrib/webauthn`.
- vukem releases: the framework engineer may tag releases after merged VFs (never BREAKING CHANGE without you).

## Held back: irreversible or outside what an agent can do alone

These need you even with the blanket approval, because they cannot be undone, touch real users, spend money, or need your accounts:

- Production cutover (P2.G): moving `hexascore.app` DNS from Vercel to Railway, read-only window on the live site, decommissioning Vercel/Neon. Staging deploys on Railway go ahead; production cutover waits for you.
- Real production data (P2.F): export from Neon, purge of synthetic users, import into the Railway Postgres. Rehearsals use synthetic or copied data only.
- Stripe live mode, Google Play publishing, developer accounts, new recurring third-party spend.
- Legal reads still owed (D-36 Twitch/IGDB, A-05 Open Library, Steam/GOG terms) and `privacy@hexascore.app` (D-39); connectors gated by them stay off.
- `v1.0.0` release of the app (equals cutover).

## Open questions

- Q-D-001 (2026-10-29): the auto-mode classifier refused twice to launch the Opus "especialista" worker, even after your approval in chat. The orchestrator covers reviews itself. To restore the fourth worker, allow subagent launches in your permission settings or run the next session outside auto mode.
- Q-D-002 (2026-10-29): vukem PR #6 commits `.claude/settings.json` with `"disableAllHooks": true`, which disables hooks for every clone of the framework. Kept because you asked; consider moving it to an uncommitted `settings.local.json`.
- Q-D-003 (2026-10-29): `docs/web-research.md` IMDb source list has no item 6 (renumbering hid the gap), and 34 numeric citation markers have no reference definition (markdownlint MD052). Content only you can restore.
- Q-D-004 (2026-10-29): Vercel "development" env pulled to `C:\dev\hexascore-app\vercel-env\.env.development.local` (outside repos). Stripe keys are test mode and AbacatePay is sandbox, so agents use them. `DATABASE_URL` points at a Neon endpoint that may be the production database: agents are forbidden to connect to it. Confirm whether the development env has its own Neon branch; if yes, it can serve the P2.F migration rehearsal.
