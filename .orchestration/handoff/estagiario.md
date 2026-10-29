# estagiario handoff (2026-10-29, after P2.B.1.1-3)

## Done this relaunch
- P2.B.1.1-3 implemented on agent/estagiario/P2.B.1.1-3 (pushed, 4 commits, head in reports/P2.B.1.1-3.md). Status parcial: setup-windows job blocked (see report risk 1). REV queued: queue/especialista/1-REV-P2.B.1.1-3.md (rodada 0).
- Earlier: E-016 branch agent/estagiario/E-016 pushed, REV queued; E-012 spike branch local.

## Open items for the orchestrator
- setup-windows job: vukem CLI not installable in CI from docs we have; windows-2025 has no Linux container engine for `docker compose`. Decide: install method + compose strategy, or accept a red job.
- Acceptance grep `uses: [^@]+@v?[0-9]` is a false positive for digit-leading SHAs; amend the acceptance to `@v?[0-9]+\.[0-9]`.
- Dockerfile uses ARG APP_VERSION, roadmap says VERSION; CI passes APP_VERSION. Fix roadmap wording or Dockerfile.
- Port 9000 conflict between local RustFS and the spike stack (from earlier session, unresolved).
- Legal i18n copy (E-015): 8 rewordings await owner decision.

## Pitfalls
- golangci-lint can print "parallel golangci-lint is running" when another worker holds the lock; retry after a wait.
- Bash tool: a bare `python3 -` heredoc hung on this Windows box and pushed the command to background; avoid it.
- MSYS: `git show origin/main:path` once failed with a mangled ref; prefer `git show "origin/main:path"`.
- markdownlint-cli2 lints all 57+ .md files regardless of argument.
- docker compose up works here now (postgres and rustfs healthy on 5433/9000); `docker compose down` run at end of task.

## Branch state
- wt/estagiario: agent/estagiario/P2.B.1.1-3 (pushed, clean).
- queue/estagiario, active/estagiario: empty.
