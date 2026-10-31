# VF-NEW-AUDIT-6 ratelimit burst allowance
status: concluida
branch: fw: fix/vf-ratelimit-burst @ bcbfbd04 (not pushed, no PR; worktree wt/fw-VF-NEW-AUDIT-6)
files: ratelimit/ratelimit.go, ratelimit/ratelimit_test.go, docs/subsystems.md, site/content/docs/security/rate-limiting.mdx, site/content/last-updated.json
acceptance:
- `go test ./ratelimit/` → ok (TestMemoryBurst: 5 at once then 1/s, RateLimit-Limit = burst; TestMemoryBurstClampsInputs)
- `task ci` → EXIT=0, no FAIL; `task api:check` → rc 0 (additive API)
validation: golangci-lint ok, vet ok, tests ok
changelog: Added - ratelimit.NewMemoryBurst(rate, window, burst): in-process GCRA limiter with a burst apart from the steady rate.
roadmap: no box
risks / follow-ups: only the Memory limiter has it (Window and the Redis GCRA limiter do not; app.RateLimiters() Factory unchanged, apps use NewMemoryBurst directly). RateLimit-Policy header not added (request mentioned it; RateLimit-Limit reports burst). Specialist: push, PR, merge on green CI. last-updated.json already refreshed in the commit.
vukem-feedback status line: VF-NEW-AUDIT-6 -> fixed in vukem branch fix/vf-ratelimit-burst (NewMemoryBurst), pending PR/release; adoption after release as ADOPT-vukem-<version>.
