# VF-NEW-OWNER-schemathesis native Schemathesis support (api:fuzz)
status: concluida
branch: fw: fix/vf-schemathesis @ 5f45c467 (not pushed, no PR; worktree wt/fw-VF-NEW-OWNER-schemathesis)
files: fuzz.go, fuzz_hook.py, fuzz_test.go, fuzz_integration_test.go (tag fuzzintegration), command.go, app.go, site/content/docs/rest-apis/fuzz-testing.mdx, cli/reference*.mdx, navigation.ts, docs/cli.md, examples/api/README.md, CHANGELOG.md, last-updated.json
design: one command `api:fuzz` (flags -prefix -url -image -report-dir -max-examples -workers -rate-limit-bypass -dry-run -docker); one public option `WithFuzzAuth(func(ctx,*App)(FuzzAuth,error))` + type FuzzAuth (no FuzzConfig, no QUESTIONS entry: surface is small); fallback env VUKEM_FUZZ_TOKEN = bearer. Runner is an unexported interface (fake in unit tests, exec in prod). Credential goes via env VUKEM_FUZZ_AUTH, referenced as ${VUKEM_FUZZ_AUTH} in the generated toml, never in argv. Config derived from the doc: GET/HEAD/OPTIONS only disables use_after_free and ensure_resource_availability; unsupported_method and allow_header_conformance off; embedded hook rotates X-Forwarded-For (needs TRUSTED_PROXIES). Image schemathesis/schemathesis:4.29.0 (override with -image digest). No Go dependency.
acceptance:
- go test -run Fuzz . -> ok (argv/env no-leak, dry-run, read-only detection, toml, bad header/url/empty doc, hook error)
- task ci -> EXIT=0 (also fixed TestCLIReferenceDocumentsEveryCommandAndFlag by documenting flags); task api:check -> rc 0 (additive)
changelog: Added - vukem api:fuzz + WithFuzzAuth.
risks / follow-ups: NOT run against a real Schemathesis container (docker not exercised here); toml `headers` with ${VAR} interpolation and `--report junit,vcr` follow the 4.x docs from memory: run `go test -tags fuzzintegration -run TestFuzzRealSchemathesis .` where Docker exists before release. Image is a tag, not a digest (no registry access to resolve one). Command targets a running app; it does not start one. HexaScore adoption (replace qa/schemathesis/) is a separate ADOPT row after release. site check:mdx/lint/build not run (no node_modules in worktree).
vukem-feedback status line: VF-NEW-OWNER-schemathesis -> implemented in vukem branch fix/vf-schemathesis (api:fuzz, WithFuzzAuth), pending PR/release; HexaScore adopts after release as ADOPT-vukem-<version>.

## Review (especialista, 2026-10-31)
verdict: CORRIGIDO_E_APROVADO (full run on GitHub CI per owner rule 8.2)
re-run: `go test -tags fuzzintegration -run TestFuzzRealSchemathesis -count=1 .` (Docker Desktop, real image) → ok, both tests; `go test -run Fuzz .` ok; site `npm ci`, `check:mdx` 0 errors (12 existing warnings), `lint` ok, `build` ok
checked: `--report junit,vcr` and `--report-dir` against `schemathesis run --help` of 4.29.0 (formats junit, vcr, har, ndjson, json, wfc, allure); `${VUKEM_FUZZ_AUTH}` toml interpolation proven on the real image (every /api call carried the header); credential never in argv or files (env only); config accepted by the 4.29.0 schema; image user is non-root (`schemathesis`)
fixes: 1. image pinned by digest: default `schemathesis/schemathesis:4.29.0@sha256:5586e94460479271714d47613175b3620c44b38b3e1babc936636108f1672a9e` (RepoDigests after pull), docs tables and CI example updated 2. work dir 0755 and files 0644: with 0700/0600 the non-root image user cannot read the bind mount on Linux (Docker Desktop hides this); the files hold no secret 3. new `TestFuzzRealSchemathesisSendsAuthHeader` (tag fuzzintegration) 4. merged origin/main, last-updated refreshed
not verified: a Linux host run (the permission fix is reasoned, not observed)
PR: AlexGalhardo/vukem#113
