# Urgent framework task FW-CI-MONGO (priority 0)

Objective: make vukem CI job "Mongo driver (integration)" green again; it blocks every PR merge (seen on PR #6, run 40795887049, two attempts).

Evidence: `.github/workflows/ci.yml` line ~151 uses service `image: mongo:8` with `ports: ["57017:27017"]`; tests fail with `dial tcp [::1]:57017: connect: connection refused` after 10 s (server never reachable). Last green CI on main: 2026-10-25. `mongo:8` is a floating tag (also against the pinned-image rule), so a new 8.x image is the prime suspect; check the service container logs in the job ("Initialize containers" step) before assuming.

Do: root-cause it (service logs, health options, IPv6 `::1` vs `127.0.0.1`, image change); fix with an exact pinned image tag (or digest) plus a health check if missing; branch `fix/ci-mongo-service` from `origin/main`; Conventional Commit `ci(mongo): ...`; PR; merge when all checks are green. Do not touch PR #6 (owner's); after your fix merges, run `gh pr update-branch 6` is NOT allowed either — just report, the orchestrator handles PR #6.
Report: `$ORCH/reports/FW-CI-MONGO.md`.
