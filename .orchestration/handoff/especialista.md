# Handoff especialista (2026-10-29, session 11)
STATUS: IDLE. No task in progress. Queue: only blocked 9-APP-PASSKEYS-b, 9-P2.F.2.2b.
Done this session: REV-P1.I.2.2 APROVADO, REV-P1.I.2.3 APROVADO, DOC-processed-events (self, branch agent/especialista/DOC-processed-events), P2.E.3.1 (self, branch agent/especialista/P2.E.3.1; staging drill deferred).
Tools of this session (scratchpad of session 27958981..., recreate if gone):
- Isolated e2e stack: override file with postgres/rustfs `ports: !reset []`, app-init/app image hexascore-app:e2e-esp, app port 127.0.0.1:3200; COMPOSE_PROJECT_NAME=hx-especialista, APP_KEY=base64:$(openssl rand -base64 32); playwright with E2E_BASE_URL=http://127.0.0.1:3200 E2E_APP_CONTAINER=hx-especialista-app-1.
- Tests needing pg_dump/pg_restore: docker run golang:1.27.3-alpine3.24, mount worktree at /src and C:/Users/aleex/go/pkg/mod at /go/pkg/mod, volume hx-esp-gocache, GOFLAGS=-buildvcs=false, TEST_DATABASE_URL=...@host.docker.internal:5433/postgres, `apk add postgresql18-client`, MSYS_NO_PATHCONV=1.
- Integration scratch DB: docker exec hexascore-postgres-1 psql -U hexascore -d postgres (create/drop hx_esp_test).
Pitfalls: golangci-lint --allow-parallel-runners and `golangci-lint fmt` for golines; .env reads denied; staging runs APP_ENV=production (cannot gate on env); vukem cron ORs day-of-month and day-of-week; pg_restore needs PATH in its env.
