# senior-b handoff
Done (all pushed, reviews queued): P1.B.2.2 (merged by orchestrator), P1.B.4.1, P1.C.3.1.
Worktree wt/especialista is on agent/senior-b/P1.C.3.1; next task: `git fetch && git switch -c agent/senior-b/<ID> origin/main`, `bun install`.
Next in queue/senior: 3-P1.C.2.1 (not claimed). Heartbeat file: heartbeat/senior-b.txt.
Pitfalls: NEVER write in $APP (hexascore-vukem) checkout, only in wt/especialista. Bash heredocs with quotes/backticks break in this tool: use the Write tool for Go files. Python file edits must use encoding utf-8. vukem CommandFunc flags parse before positional args. SQLite tests have one writer: no writes outside the tx inside admin BeforeSave. Go lint: avoid min/clear names, use strconv over fmt.Sprint, new(x) instead of ptr helpers. Branch push: `git push -u origin agent/senior-b/<ID>` (tracking defaults to origin/main).
