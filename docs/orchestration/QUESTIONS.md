# Questions for the owner

Open questions raised by the orchestrator and workers. Answer inline (`Answer:` line); the orchestrator picks answers up each cycle.

## App

- Q-001 (2026-10-28, P0.F.1.3): add `golang.org/x/crypto/scrypt` as a direct dependency of the spike (and later the app) to verify Better Auth scrypt hashes? ADR 0008 names it; AGENTS.md requires your explicit yes before `go get`. Exact version: `golang.org/x/crypto v0.58.0` (latest stable on 2026-10-29, task E-002).
  - Answer:

## Framework

Release requests and approvals for vukem (the framework engineer never tags without an approved line here).

- Owner rule (2026-10-28): the framework engineer may merge PRs with green CI; tags and releases only when the owner asks.
