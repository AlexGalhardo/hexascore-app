# FIX-ssr-supervise-flake Deterministic supervise restart test
status: concluida
branch: agent/senior/FIX-ssr-supervise-flake @ adab0c8
files: app/ssr/supervisor_test.go
root cause: the test waited at most 5 s for 3 restarts, but each restart spawns the test binary, so under CPU load start-up alone exceeded the budget ("restarted 2 times, want >= 3"). Supervise itself was correct. The 2 s wait for return after cancel had the same weakness (child kill path).
fix: wait for the 3rd run and for Supervise to return without wall-clock deadlines; go test -timeout bounds a real hang. No timeout raised.
acceptance:
- `go test ./app/ssr/ -count=3` -> ok
validation: gofmt clean, go vet ok, tests app/ssr ok
changelog: Fixed: ssr supervisor restart test no longer flakes under load.
roadmap: no box
risks / follow-ups: none

## Review (especialista, 2026-10-31)
verdict: APROVADO
re-run: `go test -count=3 ./app/ssr/` ok; go vet ok
checked: the flake came from a wall-clock budget around child process start-up, not from Supervise; the loop now ends on the observed restarts and `go test -timeout` bounds a real hang (no timeout bump)
PRONTO_PARA_MERGE agent/senior/FIX-ssr-supervise-flake (merged b6c3fe8)
