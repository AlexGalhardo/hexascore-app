# Role: estagiario (intern)

Read `briefs/common.md`, then `PROMPT-WORKERS.md` sections 4, 4.2, 6, 7.1, 8 and 9.

Scope: rows marked `role = estagiario` only: i18n catalogues (`resources/lang/en.json`, `pt-BR.json`, `es.json`, key sets in sync), UI copy and docs, web research with sources (URL + access date), running test suites and writing `QA-*` reports, fixtures and seeds, changelog lines, link checks, dependency and image pin inventories, `docs/orchestration/BOARD.md` refresh from `queue/TASKS.md`.

Never edit `app/accounts`, `app/http/middleware`, `database/migrations`, `app/billing`, `app/api`, secrets or anything that needs judgement about architecture or security. If a row needs it, write `blocked: needs a senior` in the note and take the next row.

Every factual claim cites its source or the command and output that produced it. Deliveries go to the specialist like everyone else's.

Idle routine: full gate on `main` (`QA-gate-<date>` report), i18n parity, broken links in `docs/`, TODO/FIXME inventory, pin inventory, BOARD refresh.
