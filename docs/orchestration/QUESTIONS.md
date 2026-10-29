# Questions for the owner

Open questions raised by the orchestrator and workers. Answer inline (`Answer:` line); the orchestrator picks answers up each cycle.

## App

- Q-001 (2026-10-28, P0.F.1.3): add `golang.org/x/crypto/scrypt` as a direct dependency of the spike (and later the app) to verify Better Auth scrypt hashes? ADR 0008 names it; AGENTS.md requires your explicit yes before `go get`. Exact version: `golang.org/x/crypto v0.58.0` (latest stable on 2026-10-29, task E-002).
  - Answer: approved (owner blanket approval of pending decisions, 2026-10-29). Use `golang.org/x/crypto v0.58.0`.

## Framework

Release requests and approvals for vukem (the framework engineer never tags without an approved line here).

- Owner rule (2026-10-28): the framework engineer may merge PRs with green CI; tags and releases only when the owner asks.
- Owner rule update (2026-10-29): owner pre-approved all pending and future framework release requests while away. The framework engineer may release (`vX.Y.Z`, one tag at a time, never moving a tag) after merged VFs, and then create the adoption task. BREAKING CHANGE releases still need the owner (log in `questions-to-dev.md`).
- FQ-001 (2026-10-28, VF-001 passkeys/WebAuthn, high): vukem `docs/security.md` lists "Post-1.0: passkeys/WebAuthn". Implementing it now is a roadmap change and either a new dependency (`github.com/go-webauthn/webauthn`) or a stdlib-only WebAuthn verifier (CBOR/COSE parsing, `none` attestation, ES256/RS256/EdDSA) maintained in-framework. Options: (a) keep post-1.0, HexaScore stays without passkeys at launch (D-06); (b) stdlib implementation in `auth` now; (c) go-webauthn dependency in a `contrib/webauthn` module. The framework engineer skips VF-001 until answered.
  - Answer: approved under the owner's blanket approval (2026-10-29); orchestrator picked (a) keep post-1.0, the least risky option consistent with D-06 (no new dependency, no roadmap change). VF-001 status becomes `wontfix (post-1.0)` with this reference. Revisit in `questions-to-dev.md`.
- FR-001 (2026-10-29, release request v0.33.0, minor, non-breaking): BLOCKED by the Claude Code permission classifier ("Create Public Surface") when the framework engineer ran `go run ./internal/tools/release v0.33.0` in `C:\dev\hexascore-app\wt\framework`; the coordinator's relayed pre-approval does not count as the owner's consent for the tool. Nothing was committed or tagged. Ready on vukem `main` at afbce68: PR #7 (VF-002 strict SSR), #8 (CI mongo pin), #9 (VF-027 serve -with-ssr / build -ssr -ldflags / WithVersion), #10 (VF-024 i18n path prefixes), #11 (VF-005 sorted cursor pagination); #12 (VF-015 Router.Host) and #13 (VF-017/026 NoTransaction + migration lock) are merging. Owner options: (a) add a Bash permission rule allowing the release tool and tag pushes for this worker, then reply "approved" here; (b) run it yourself per vukem docs/maintenance.md "Releasing through a pull request" (`go run ./internal/tools/release v0.33.0` on main, move the commit to `release/v0.33.0`, PR, squash-merge, `git tag -a v0.33.0`, `git push origin v0.33.0`, `go run ./internal/tools/release -publish-only v0.33.0`, then the nested-module bump PR). No nested module has non-test changes except test files in drivers/postgres and drivers/mysql (test-only: no nested tags needed).
- FR-002 (2026-10-29, session 5): `gh pr merge 29 --squash` (VF-045, all checks green, mergeStateStatus CLEAN) was denied by the Claude Code auto-mode classifier ("Merge Without Review"), although the owner rule above allows merging green PRs. Nothing was bypassed. Without a merge permission the v0.33.0 release (FR-001) also stays blocked, because the vukem release flow merges a `release/vX.Y.Z` PR. Owner options: (a) add Bash permission rules for this worker: `gh pr merge * --squash`, `go run ./internal/tools/release *`, `git push origin v*`; (b) merge PR #29 (and later framework PRs) yourself and run the release per FR-001 (b). The framework engineer keeps opening PRs and lists them here.
  - PRs green and waiting for merge: #29 (VF-045).
  - Answer:
