# Runbook: dependencies

Owner documents: SDD 13 NFR-13-5, §7.7, §7.8; SDD 11 §6.3.

## Automation

- Dependabot (`.github/dependabot.yml`) opens weekly PRs for `gomod`, `bun`, `docker`, `docker-compose` and `github-actions`. Go and JS updates are grouped.
- CI job `deps` (`.github/workflows/ci.yml`, runs on push, pull request and the weekly schedule) runs `osv-scanner scan source -L go.mod -L bun.lock -L sdks/ts/bun.lock` (the published TypeScript SDK has its own lockfile) and `bun audit --audit-level=high`. Either failing fails the job.
- The CI `vuln` job keeps running `govulncheck` for reachable Go vulnerabilities.
- `osv-scanner` is a CI binary, not an app dependency. It is pinned by release tag and SHA-256 in the workflow `env` (`OSV_SCANNER_VERSION`, `OSV_SCANNER_SHA256`). To bump: pick the new release, copy the `osv-scanner_linux_amd64` line from its `osv-scanner_SHA256SUMS`, update both values together.
- Every GitHub Action is pinned by commit SHA with a version comment; Dependabot proposes the bumps.

## Fix SLA

Counted from the day the advisory is first reported by Dependabot, `osv-scanner` or `bun audit`.

| Severity | Fix within |
| --- | --- |
| Critical | 72 hours |
| High | 7 days |
| Medium | 30 days |

Low severity is fixed with the regular weekly Dependabot PRs. If no fixed version exists, record a dated exception in `questions-to-dev.md` with the mitigation and review it weekly. Pin versions exactly; never use `latest` or pre-releases.

## Checking alerts

```sh
gh api repos/AlexGalhardo/hexascore-app/dependabot/alerts --jq '.[] | [.security_advisory.severity, .dependency.package.name] | @tsv'
```
