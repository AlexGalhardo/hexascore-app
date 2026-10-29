# FW-CI-MONGO

- Status: done, merged in vukem main (PR #8, squash 4e9acbd).
- Branch: fix/ci-mongo-service.
- Root cause: `mongo:8` now resolves to 8.3.13 (digest e16e606f), which exits at startup on the runner kernel (ubuntu24/20261025, kernel >= 6.19): service log of run 40795887049 `MongoDB cannot start: Linux kernel versions 6.19 and newer has a known incompatibility with this version of MongoDB` (SERVER-121912). Not an IPv6 issue; port mapping was fine.
- Fix: `.github/workflows/ci.yml` service image `mongo:9.0.4@sha256:452bdaf2...` (digest via `docker buildx imagetools inspect`), mongosh ping health check (`--health-retries 30`); `.config/compose.yaml` and `docs/testing.md` moved to 9.0.4.
- Commands: `actionlint .github/workflows/ci.yml` -> ok; PR #8 checks 26/26 pass (Mongo driver (integration): pass).
- Follow-up: PR #7 branch updated with main (`gh pr update-branch 7`); PR #6 untouched (owner's), its Mongo job will pass after updating from main.
- Release: not needed (CI-only change).
