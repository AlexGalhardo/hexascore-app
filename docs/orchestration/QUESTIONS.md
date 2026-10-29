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
