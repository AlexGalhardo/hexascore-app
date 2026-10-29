task: P1.G.1.2 (review)
branch: agent/senior-b/P1.G.1.2
repo: app
report: C:\dev\hexascore-app\hexascore-vukem\.orchestration\reports\P1.G.1.2.md
acceptance: go test ./app/http/middleware -run 'TestSecurity_A05_'; golangci-lint run --build-tags integration ./...; go test ./...
focus: /csp-reports collector deliberately left to P2.D.1.1; confirm acceptable.
rodada: 1
