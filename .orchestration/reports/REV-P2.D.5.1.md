REV P2.D.5.1: CORRIGIDO_E_APROVADO (base already merged ee77cc1). Final branch: agent/especialista/P2.D.5.1-fix (5d3ba1b).
Verified: OSV_SCANNER_SHA256 160c6253...218b == osv-scanner_linux_amd64 line of v2.7.0 osv-scanner_SHA256SUMS (fetched); bun.lock is a supported lockfile in v2.7.0 docs/supported_languages_and_lockfiles.md; actions pinned by SHA; install uses --ignore-scripts.
Re-run: bun audit --audit-level=high (root) -> "No vulnerabilities found (checked 249 packages)"; markdownlint docs/runbooks/dependencies.md -> 0 issues; YAML parses.
Fixed: sdks/ts/bun.lock (added by P1.H.4.1) was outside both osv-scanner and Dependabot -> `-L sdks/ts/bun.lock` in the deps job, dependabot bun `directories: ["/", "/sdks/ts"]`, runbook line updated.
Finding: sdks/ts audit -> 3 high on js-yaml@4.2.0 (dev, via @hey-api 0.99.0 exact pin) -> task 2-P2.D.5.2 (senior). Until it lands, the deps job will fail on main after merging the -fix.
Not run: osv-scanner binary locally (auto-mode denied executing a downloaded binary).
