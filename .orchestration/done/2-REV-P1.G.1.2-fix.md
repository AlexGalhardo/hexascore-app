task: P1.G.1.2-fix (follow-up bug fix to merged P1.G.1.2, found while testing P1.D.3.1b)
branch: agent/senior-b/P1.G.1.2-fix
repo: app
report: none; commit message explains: CSPSources starts worker-src from script-src (nonce + strict-dynamic copied) and frame-src from default-src ('self'); now written into the base CSP string.
acceptance: go test ./app/http/middleware -run 'TestSecurity_A05_'
rodada: 1
