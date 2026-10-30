# REV FIX-locale-posts-rest
repo: app; branch: agent/senior/FIX-locale-posts-rest (stacked on agent/especialista/FIX-locale-redirects-fix); report: reports/FIX-locale-posts-rest.md
re-run: bun run typecheck && bun run lint && bun run test; go test -tags integration ./tests -run TestLocalePrefix_AccountRedirects
rodada: 1
