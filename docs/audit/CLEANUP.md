# Cleanup ledger (architect audit, 2026-11-02)

Rules (PROMPT-ARCHITECT.md §4): nothing is deleted before (a) it is listed here with reason and evidence, (b) `git status`, `git worktree list` and `git branch --no-merged` were checked, (c) a backup ref exists. Backup: tag `backup/pre-cleanup-2026-11-02` on `hexascore-vukem` and on `vukem` (created in step 0 of the execution plan below). Branch `legacy` is never touched (AGENTS.md).

Owner decision 2026-11-02: delete only branches already merged into `main`; unmerged branches are listed for the owner. Evidence comes from the read-only inventory of 2026-11-02 (`git cherry`, `git merge-base --is-ancestor`, commit subjects and shortstats).

## A. Executed in this audit (commit per category on branch `audit/architect-2026-11-02`)

Status column is updated as each step runs.

| # | Item | Reason | Evidence | Status |
| --- | --- | --- | --- | --- |
| A1 | Dangling junction `sdks/python/.venv/lib64` and the whole `sdks/python/.venv` (moved to the audit scratch folder) | Gitignored `uv` venv created 2026-11-02 13:21 with Linux-style links; broke `docker compose build` twice on Windows because `.dockerignore` does not exclude it | build logs 2026-11-02 ("invalid file request sdks/python/.venv/bin/python") | done (moved, not deleted; regenerate with `cd sdks/python && uv sync`) |
| A2 (done, commit b965c5d5) | `.dockerignore`: add `**/.venv/`, `wt/`, `.ci-local/`, `.hypothesis/`, `.lighthouseci/`, `.playwright-mcp/`, `apps/mobile/android/`, `apps/mobile/ios/`, `qa/reports/` | Build context carried the whole workspace; only `cmd`, `app`, `bootstrap`, `config`, `database`, `internal`, `routes`, `resources`, `openapi`, `go.*`, `package.json`, `bun.lock`, `vite.config.ts`, `tsconfig*.json`, `biome.json`, `.buildtags` are needed by the Dockerfile | `Dockerfile:5-36` COPY lines | done |
| A3 (done, commit b965c5d5) | `local.properties` → add to `.gitignore` (keep the file) | Android SDK path, machine-specific, untracked and not ignored | `git status`: `?? local.properties` | pending |
| A4 | `.orchestration/queue/TASKS.md` rows already merged | Board lists merged work as `ready`; 683 lines | `.orchestration/queue/TASKS.md` vs `git log` (rows cite merge shas) | pending (rewrite with the new PROMPT-WORKERS protocol) |

## B. Worktrees and branches, hexascore-vukem (owner-approved category: merged into main)

Backup first: `git tag backup/pre-cleanup-2026-11-02 main` plus `git bundle create ../backup-history-2026-11-02/pre-cleanup-branches.bundle --all` (the pre-rewrite history is already in `backup-history-2026-11-02/hexascore-all.bundle`, 79.5 MB).

| # | Item | Reason | Evidence | Action |
| --- | --- | --- | --- | --- |
| B1 | Worktrees `wt/docs-i18n`, `wt/item-r2`, `wt/pricing`, `wt/ui-item`, `wt/ui-landing`, `wt/ui-legal`, `wt/ui-listing`, `wt/ui-misc` and their branches `agent/senior-c/DOCS-i18n`, `agent/senior-a/UI-item-r2`, `agent/senior-b/BILLING-plans-r2`, `agent/senior-b/UI-item-page`, `agent/senior-a/UI-landing`, `agent/estagiario/UI-legal-contact`, `agent/orchestrator/UI-listing`, `agent/orchestrator/UI-misc-r2` | Every commit is in `main` (rebased): `git cherry main <branch>` shows `-` for all but four commits, and those four have a `main` commit with identical subject and shortstat; all worktrees clean | inventory 2026-11-02 §A1 | `git worktree remove` (unlink `node_modules` junctions first, AGENTS.md), then `git branch -D` |
| B2 | Worktree `wt/release-linear` (`release/history-linear`, upstream gone) | Staging branch of the history rewrite; tip subject in `main`; pre-rewrite history preserved in the bundle | inventory §A1, §B | `git worktree remove`, `git branch -D` |
| B3 | Directories `wt/apk-build`, `wt/fix-mobile-entity-links`, `wt/routes`, `wt/web-forms`, `wt/web-shell` | Not worktrees (no `.git`), contain only `node_modules` leftovers | inventory §A2 | delete directories |
| B4 | `wt/.esp-rm-comm.log`, `wt/.esp-rm-obs.log`, `wt/.esp-rm-p4b4.log` | Logs of branch deletions already done on 2026-10-31 | inventory §E | delete files |
| B5 | 21 detached worktrees under `C:\Users\aleex\AppData\Local\Temp\tmp.*` whose HEAD is an ancestor of `main` | Clean, merged | inventory §A3 (list of shas) | `git worktree remove --force <path>` or `git worktree prune` after deleting the directories |
| B6 | 17 detached Temp worktrees on the pre-rewrite history (tips c47208c7 ×7, 99c26029 ×2, bdd8ccf6, 9380256b ×3, d4a7816f, 0a6f74c4, 72e13fa3, d642c66b) | Clean; old history is in the backup bundle; every tip but the "session 30 handoff" note is in `main` by subject | inventory §A3 | same as B5 (they hold no branch) |
| B7 | 81 local pre-rewrite branches whose tip subject exists in `main` (list in the inventory §B: `agent/especialista/*`, `agent/estagiario/*`, `agent/senior/*`, `release/history-linear-stage1`, `release/v0.12.0-r2`, …) | Rewritten into `main` by subject; content match not proven because the trees differ by ~1,200 files between histories | inventory §B | **owner call**: they are "merged" only by subject. Recommended: delete after the backup bundle, since the rewrite was the owner's decision and the bundle keeps them |
| B8 | 14 local pre-rewrite branches whose tip subject is NOT in `main`: `agent/especialista/P1.F.4.1-followup-ui-purge-fix`, `agent/especialista/P2.H.2.1-fix`, `agent/estagiario/DOC-vf-apiproblem-code`, `agent/implementer/P1.C.7.1`, `agent/implementer/P2.F-data-migration`, `agent/implementer/P4.A.2.4`, `agent/implementer/import-allow-no-login`, `agent/implementer/manual-access`, `agent/senior/IMPL-web-v1-recent-changes`, `agent/senior/QA-e2e-speed`, `build/apk-sideload`, `checkpoint/audit-worktree-2026-10-31`, `fix/item-page-cover-prop`, `release/history-linear-pre-c2` | Unmerged by the ancestor test and by subject | inventory §B | **keep**; listed for the owner. Notable: `P2.F-data-migration` (S3 cover-upload retry), `manual-access` (CI production environment gate), `import-allow-no-login` (wip commands), `checkpoint/audit-worktree-2026-10-31` |
| B9 | Stashes `stash@{0}` ("wip parity rows E24 E25 E26 E39") and `stash@{1}` ("WIP on agent/especialista/P1.G.3.2d-fix") | Unknown content; not inspected | `git stash list` | **owner call**; keep |
| B10 | Uncommitted: `D docs/review/2026-11-01-architect-brief.md`, `M tsconfig.json`, `?? resources/js/vite-env.d.ts` | Someone's in-progress change (drops `vite/client` from `types` and adds a reference file); the brief deletion matches AGENTS.md ("being moved out of git") | `git status` 2026-11-02 | **owner call**; left untouched and excluded from audit commits |

## C. vukem

| # | Item | Reason | Evidence | Action |
| --- | --- | --- | --- | --- |
| C1 | Branches `chore/nested-v0.37.0`, `docs/readme-and-updates` | Squash-merged (#142; all subjects in `main`) | inventory §C | delete after backup tag |
| C2 | Branch `release/v0.37.0` | Release commit in `main`; 4 changelog edits (3 files, +36/-30) are not | inventory §C | **owner call**: cherry-pick the changelog edits or drop |
| C3 | Branches `fix/railway-otlp-traces`, `agent/implementer/railway-observability` | "keep Railway OTLP export to traces", not in `main`; superseded? D-35 keeps Railway-native observability and production has `OTEL_*_EXPORTER` variables | inventory §C, Railway variable names | **owner call**; recommended: open a PR from `fix/railway-otlp-traces` if traces still matter, delete the older variant |

## D. Outside git (owner decides, nothing executed)

| # | Item | Evidence | Recommendation |
| --- | --- | --- | --- |
| D1 | Docker: 38 images 12.8 GB (16 `hexascore*` QA/CI images), build cache 25.9 GB (21.2 GB reclaimable), 35 volumes of which ~20 `hexascore-qa-*_postgres` | `docker system df`, `docker images`, `docker volume ls` 2026-11-02 | `docker builder prune`, remove `hexascore-qa-*` projects and volumes older than the current session; keep `hexascore` (dev) and `hexascore-qa-audit` until this audit ends |
| D2 | Compose project `hexascore` started from the removed path `wt/especialista-ADOPT-VF-001/docker-compose.yml` | `docker compose ls` | harmless; restart from the main checkout at the next `docker compose up` |
| D3 | Root files `hexascore-spike.bundle` (203 KB, 2026-10-31), `apply-button.png`, `PROMPT.md` (2026-10-28), `cloudflare.md`, `developer-manual-todos.md`, `questions-to-dev.md`, `PROMPT-WORKERS.md` (root copy, 2026-10-31, superseded by the repo copy) | root listing | archive under `backup-history-2026-11-02/` or delete the spike bundle and the root `PROMPT-WORKERS.md` copy |
| D4 | `vercel-env/` with `.env.local` and `.env.development.local` | root listing; contents not read | contains Vercel-era secrets: rotate what is still valid, then delete |
| D5 | Audit stack: compose project `hexascore-qa-audit`, image `hexascore-app:audit`, database `hexascore_qa_audit`, Chrome isolated contexts | created by this audit | removed at the end of the audit |

## E. Execution log

- 2026-11-02 18:20: A1 done (venv moved to the scratch folder; restored at the end of the audit).
- 2026-11-02 19:05: A2, A3 committed (b965c5d5).
- 2026-11-02 19:40: backup tags `backup/pre-cleanup-2026-11-02` on both repos; bundles `backup-history-2026-11-02/pre-cleanup-branches-hexascore.bundle` (80 MB, all refs) and `pre-cleanup-branches-vukem.bundle` (14 MB).
- 2026-11-02 19:50: B1, B2 done (9 worktree directories removed with `rmdir /s /q`, `git worktree prune`, 9 branches deleted); B3 done (5 stale directories); B4 done (3 logs). `wt/` is empty; `git worktree list` shows the main checkout plus the Temp entries of B5/B6.
- 2026-11-02 19:55: C1 done on vukem (`chore/nested-v0.37.0`, `docs/readme-and-updates` deleted).
- B5/B6 NOT executed: deleting the 38 `C:\Users\aleex\AppData\Local\Temp\tmp.*` checkouts was refused by the tool permission classifier ("irreversible local destruction"). Owner command, safe because every entry is clean and bundled: list them with `git -C C:\dev\hexascore-app\hexascore-vukem worktree list --porcelain | findstr Temp`, then `git worktree remove --force <path>` for each (or delete the directories and run `git worktree prune`).
- 2026-11-02 20:20 (owner answers OQ-10..13): B7 done (81 pre-rewrite branches deleted; 14 of B8 kept for the senior review task B-033); B9 done (both stashes dropped); B10 done (loose changes discarded, brief restored); C2 done (`release/v0.37.0` deleted, changelog commits dropped); C3 done (`fix/railway-otlp-traces` was already squash-merged as vukem PR #104; the duplicate PR #143 opened during reconciliation was closed; both OTLP branches deleted); D4 done (`vercel-env/` deleted without rotation, owner decision).
- Still open: B5/B6 (owner command above), D1 Docker prune, D2, D3 root files.

### E.1 Review of the 14 kept branches (B-033, 2026-11-02)

Method: `release/history-linear-pre-c2` and `build/apk-sideload` are two parallel copies of the pre-rewrite `main` (identical under `app`, `internal`, `routes`, `resources/js`, `apps/mobile`); `git diff --name-status release/history-linear-pre-c2 main` shows only intended product changes (11 files deleted in `main`: more-like-this, entity pages, the old 2FA page, the Playfair font), so that lineage is in `main` by content. For each other branch the commits not in either copy (`git log --no-merges <branch> --not release/history-linear-pre-c2 build/apk-sideload`) were checked against `main` by file and by trying `git cherry-pick -n`. The bundle `backup-history-2026-11-02/pre-cleanup-branches-hexascore.bundle` keeps every ref.

| Branch | Decision | Reason | Cherry-pick |
| --- | --- | --- | --- |
| `agent/especialista/P1.F.4.1-followup-ui-purge-fix` | delete | its one unique commit (4750ae71, export ZIP built after a purge) is in `main` in another form: `app/library/export.go` updates `WHERE status = 'running'` and skips ZIP and mail when the row is gone | none |
| `agent/especialista/P2.H.2.1-fix` | delete | no commit outside the pre-rewrite `main` lineage | none |
| `agent/estagiario/DOC-vf-apiproblem-code` | delete | no commit outside the lineage (docs and orchestration notes only) | none |
| `agent/implementer/P1.C.7.1` | delete | passkeys (register, sign in, UI, legacy import) are in `main` (`app/accounts/passkeys.go`, `database/legacyimport/passkeys.go`) | none |
| `agent/implementer/P2.F-data-migration` | delete | both cover-retry commits are in `main` (`scripts/legacy/blob-export.ts` has the retry loop and `not seekable`; 656f426a cherry-picks to an empty change) | none |
| `agent/implementer/P4.A.2.4` | delete | community cover and source-URL guards are in `main` (`app/community/uploads.go`, `allowedCover`); the rest are release-check fixes of v0.11.0 already tagged | none |
| `agent/implementer/import-allow-no-login` | delete | `IgnoreNoLogin` and its test are in `main`; the `wip` commit only adds two one-off import helpers (`cmd/activity`, `cmd/dbcounts`), the cutover is done | none |
| `agent/implementer/manual-access` | delete (owner declined the change, 2026-11-02) | `environment: Production` on the release job (e5cfbfdc) was not in `main`; the rest of the commit was an orchestration report | 4ed184d1 (`release.yml` only). Owner call: the GitHub environment `Production` must exist and may have required reviewers, which makes every release wait for approval. Owner answer 2026-11-02: no; 4ed184d1 not merged |
| `agent/senior/IMPL-web-v1-recent-changes` | delete | no commit outside the lineage | none |
| `agent/senior/QA-e2e-speed` | delete | no commit outside the lineage | none |
| `build/apk-sideload` | delete | copy of the pre-rewrite `main` (see method); the APK build lives on `main` and the sideload APK is a release asset | none |
| `checkpoint/audit-worktree-2026-10-31` | delete | hydration waits, inert-until-Inertia (`page_script.go`) and the osv-scanner argument order are in `main`; the checkpoint commit is Playwright `test-results` output | none |
| `fix/item-page-cover-prop` | delete | `"cover": page.Cover` is in `app/controllers/item_controller.go` on `main`; the rest are release commits | none |
| `release/history-linear-pre-c2` | delete | the pre-rewrite `main` lineage, in `main` by content | none |

The orchestrator deletes all 14 branches after merging B-033 (`manual-access` too, its one change is cherry-picked); `legacy` is untouched.
