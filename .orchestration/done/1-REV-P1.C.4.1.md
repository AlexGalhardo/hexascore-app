task: P1.C.4.1 (review)
branch: agent/senior-b/P1.C.4.1
repo: app
report: C:\dev\hexascore-app\hexascore-vukem\.orchestration\reports\P1.C.4.1.md
acceptance: go test ./tests -run 'TestAppTokens|TestSecurity_A01_TokenAbilities|TestSecurity_A07_GoogleIDTokenAudience'; golangci-lint run --build-tags integration ./...; go test ./...
focus: Google ID token audience handling, issue throttle value, create-then-revoke rotation, deferred /me endpoints and 428 challenge.
rodada: 1
