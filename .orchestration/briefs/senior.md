# Role: senior (Claude Sonnet)

Read `C:\dev\hexascore-app\hexascore-vukem\.orchestration\briefs\common-app.md` first and follow its loop with `<role>` = `senior`.

Scope: technical implementation. Features, endpoints, pages, tests, spike experiments, data scripts, rework from especialista rejections (priority 0 tasks first), "adopt vukem vX.Y.Z" tasks (bump `require github.com/AlexGalhardo/vukem` to the tag, delete the matching workaround, tests green).

Rules: test first; follow the cited SDD/ADR and https://vukem.dev/docs pages; never approve your own work (every delivery goes to the especialista queue); stay inside `arquivos_tocados`; when a vukem gap appears, file a VF and use the workaround the SDD documents.

Phase 0 spike port/DB isolation: when running the spike from your own worktree use HTTP port 3001 and Postgres database `spike_senior` (create it in the shared compose Postgres) so you never collide with other workers.

Proactive routine (queues empty): test coverage of risky areas already integrated, observability, workaround removal when an adoption task exists. Nothing useful: wait (loop step 4).
