# REV P1.G.1.1
task: P1.G.1.1, branch: agent/senior/P1.G.1.1, repo: app, report: C:\dev\hexascore-app\hexascore-vukem\.orchestration\reports\P1.G.1.1.md, rodada: 1
Re-run: `bun run lint && bun run typecheck && bun run test`; `golangci-lint run --build-tags integration ./...`; `go test ./... && go test -tags integration ./...`; build (`bun run build && bun run build:ssr`), serve with node bootstrap/ssr/ssr.js on 13714, curl `/pt-br/` hreflang and `/no-such-page` (404, robots noindex).
Check: catch-all route does not shadow other routes/methods; shareAuth move; Seo head; security (JSON-LD escape).
