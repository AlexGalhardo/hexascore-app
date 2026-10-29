# senior handoff (2026-10-29, fourteenth session)
Delivered: P1.G.4.1 (agent/senior/P1.G.4.1 pushed, report + 1-REV queued).
Queue left: 2-P1.G.4.2 (big admin task, claim fresh), 3-P1.E.3.2 (k6 via pinned grafana/k6 image, see brief end), 3-P1.F.2.1, 4-P1.G.2.4.
Worktree wt/senior on agent/senior/P1.G.4.1; next: git fetch && git switch -c agent/senior/<ID> origin/main.
Pitfalls: TEST_DATABASE_URL inline on every go test; never grep resources/public/build; biome: run `bunx biome check --write resources/js packages` then lint (output is huge, pipe through sed/tail); bash heredocs containing "\n" in Go strings get mangled: use Edit tool; resources/public had a static robots.txt that shadowed routes (removed); docs.json is committed output of scripts/build-docs.ts (run `bun run build:docs` after editing resources/docs/*.md); lang JSON edit textually; gofmt/gci: run `golangci-lint fmt <file>`.
