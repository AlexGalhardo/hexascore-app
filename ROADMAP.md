# HexaScore ROADMAP

## Release 0.39.0: architect audit 2026-11-02

The audit (`docs/audit/00-project-map.md`, `01-web-report.md`, `02-technical-review.md`, `CLEANUP.md`) replaced the ad-hoc owner rounds with a milestone plan in [docs/ROADMAP-V1.md](docs/ROADMAP-V1.md) and a prioritised backlog in [docs/BACKLOG.md](docs/BACKLOG.md); RFCs for the catalogue worker, trending and offers live in `docs/rfcs/`; ADR 0024 defers Redis and ADR 0025 rejects MongoDB. The team protocol is `PROMPT-WORKERS.md` (orchestrator reviews and merges; senior and intern in standby). Boxes below mirror the backlog; tick them there first.

### P8.A Stabilise (V1-A)

- [x] **P8.A.1** Render cache keyed by locale prefix (B-001 code part): commit 695872f5, test `TestRenderCacheKeysStrippedLocalePrefix`. Status (2026-11-02): deploy and production check pending (B-001).
- [x] **P8.A.2** Docker build context trimmed, `local.properties` ignored: commit b965c5d5.
- [ ] **P8.A.3** Nightly jobs with explicit timeouts and a resumable `RecomputeAll` (B-002, B-010). Acceptance: staging log line `recompute scores items=<all>`; `queue:failed` empty the next morning.
- [x] **P8.A.4** LGPD erasure of scoring and search rows (B-011). Acceptance: `TestSecurity_A03_PurgeRemovesScoringRows` green. Status (2026-11-02): done in local `main` 6bdaa2b8; the user FKs are validated in a separate NoTransaction migration; `rating_flags` stay with the kept ratings.
- [ ] **P8.A.5** e2e helper fixed and `desktop-en` green; CI e2e on release candidates (B-012, OQ-14).
- [ ] **P8.A.6** Retention prunes and `item_status` index (B-024, B-025); dated dependency ignores (B-026); IGDB token cache (B-027).
- [ ] **P8.A.7** Cleanup execution per `docs/audit/CLEANUP.md` after the backup tag (B-033); board rewrite (B-034). Owner gates: OQ-10, OQ-11.
- [x] **P8.A.8** Owner: pre-deploy command `hexascore:predeploy`, staging DNS (OQ-01, OQ-02). Status (2026-11-02): pre-deploy `/app/app hexascore:predeploy` in production and staging (Railway `get-service-config`); `staging.hexascore.app` CNAME to the custom-domain target plus `_railway-verify` TXT, certificate VALID, `/up` 200, `X-Robots-Tag: noindex, nofollow`.

### P8.B Launch polish (V1-B)

- [ ] **P8.B.1** Mobile item page order, 404 pages, `og:image`, server-rendered search results (B-020..B-023, OQ-06).
- [x] **P8.B.2** Copy and SEO details, genre labels in pt-BR/es, storage health probe, alert code reconciliation (B-040..B-044). Status (2026-11-02): B-040 (link back + canonical filters), B-041 (already translated in production), B-042 (copy and home title), B-043 (bucket readiness probe), B-044 (emitters table; missing checks moved to B-049) merged in local `main`.
- [x] **P8.B.3** Documentation to English (B-046, intern): `questions-to-dev.md`, `developer-manual-todos.md`, `cloudflare.md` at the workspace root and the Portuguese answers in `docs/orchestration/QUESTIONS.md`; same facts, links intact. Status (2026-11-02): `docs/owner/cloudflare.md` and the QUESTIONS.md owner quotes translated; `questions-to-dev.md` translated to a local copy (gitignored file); `developer-manual-todos.md` does not exist.

### P8.C Catalogue worker (V1-C, RFC-001)

- [ ] **P8.C.1** Worker service on Railway with the pressure signal and persisted breaker state (CW-01, CW-02). Owner gate: OQ-03.
- [ ] **P8.C.2** Idempotency keys, jittered retries, refresh-existing job, provenance v2 and field reviews (CW-03..CW-05).
- [ ] **P8.C.3** Entity resolution rules, cover hashing and rights, 30-minute Telegram digest (CW-06..CW-08). Owner gate: OQ-04.
- [ ] **P8.C.4** Discovery v2, LLM normaliser behind its budget, storage measures of ADR 0025, SDD 06 update (CW-09..CW-14). Owner gate: OQ-05.

### P8.D Trending and offers (V1-D, RFC-002 and RFC-003)

- [ ] **P8.D.1** Trending v2 score, entry/exit/cooldown/diversity, overrides, home reads snapshots (TR-01..TR-06). Owner gate: OQ-07.
- [ ] **P8.D.2** Offers model, derivation stages, TMDB providers with attributions, verification job, item page section with disclosure (OF-01..OF-07). Owner gates: OQ-08, OQ-09.
