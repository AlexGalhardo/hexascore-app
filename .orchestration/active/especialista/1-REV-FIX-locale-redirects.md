# REV FIX-locale-redirects (rodada: 1)
Task: FIX-locale-redirects. Branch: agent/senior/FIX-locale-redirects. Repo: app. Report: reports/FIX-locale-redirects.md
Re-run: go test -tags integration ./tests -run TestLocalePrefix_AccountRedirects; golangci-lint --build-tags integration; bun lint/typecheck/test; isolated stack playwright auth profile api-keys (E2E_IMAGE_TAG per role).
