# DOC-pin-inventory Dependency, image and pseudo-version pins
status: concluida (report only; no edits)
branch: none (main @ 246ced8 at read time)
files: .orchestration/reports/DOC-pin-inventory.md
acceptance:
- `git ls-files '*package.json'` → 6 files; `rg -n '[\^~]|latest|"\*"'` over them → 0 matches (dependency ranges)
- `rg -n "^\s*(FROM|image:)"` in Dockerfile and docker-compose.yml → every image has a tag or digest; 2 are `hexascore-app:${E2E_IMAGE_TAG:-e2e}` (built locally, not pulled)
- `rg -n 'v[0-9.]+-[0-9]{14}-[0-9a-f]{12}' go.mod` → 3 pseudo-versions in `go.mod`, all `// indirect`
validation: no code touched; lint n/a
changelog: none
roadmap: no box
risks / follow-ups: `sdks/ts/package.json` has `"engines": { "node": ">=18" }`, which is an engine range, not a dependency pin. Not counted as a violation.

## Dependency ranges
None found in the 6 tracked `package.json` files.

## Docker images

| file:line | image | pin |
| --- | --- | --- |
| docker-compose.yml:8 | postgres:18.6-alpine3.24 | fixed tag |
| docker-compose.yml:24 | rustfs/rustfs:1.0.4 | fixed tag |
| docker-compose.yml:43, 69 | hexascore-app:${E2E_IMAGE_TAG:-e2e} | local build; default tag `e2e` is not a pulled image |
| Dockerfile:5 | oven/bun:1.4.4-alpine@sha256:... | tag and digest |
| Dockerfile:26 | golang:1.27.3-alpine3.24@sha256:... | tag and digest |
| Dockerfile:40 | node:26.13.0-alpine3.24@sha256:... | tag and digest |

No `latest`, `edge` or `nightly` found.

## Go pseudo-versions (`go.mod`, all `// indirect`)
- `github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761` (go.mod:25)
- `github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec` (go.mod:30)
- `golang.org/x/telemetry v0.0.0-20261010193011-f71dc92da45e` (go.mod:38)

Pseudo-versions are untagged commits. The task asks for an inventory, not a fix. Whether to replace them with tagged releases is a senior/owner decision.

## Review (especialista, 2026-10-31)

verdict: APROVADO (report only)
re-run (sampled on main @ 57497a9): no `^`/`~` in the root, mobile or TS SDK `package.json`; `docker-compose.yml` images `postgres:18.6-alpine3.24`, `rustfs/rustfs:1.0.4`, and the two locally built `hexascore-app:${E2E_IMAGE_TAG:-e2e}`; pseudo-version count in `go.mod` re-counted (see the command in the report)
checked: commands quoted, no edits; `engines` range correctly not counted
