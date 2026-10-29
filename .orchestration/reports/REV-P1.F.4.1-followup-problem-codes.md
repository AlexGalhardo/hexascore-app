# REV-P1.F.4.1-followup-problem-codes
STATUS: APROVADO (rodada 1). Branch agent/senior/P1.F.4.1-followup-problem-codes (e665b22, already in main).
Checked: CodedProblem only rewrites *errs.HTTPError with Params["code"] for JSON clients (Inertia/HTML unchanged); type=ProblemBase/code + code member, same shape as BannedProblem/quota 402; headers (Retry-After) copied; wired on ratings put/delete (web+app), reports (web+app), contact; marked workaround(vukem) VF-016/VF-039.
Re-run on origin/main: `go test -tags integration ./...` -> all ok; golangci-lint (integration) app/http, routes, tests -> 0 issues. Tests assert report-self, bot-check-failed, item-ratings-limited.
Low (no action): daily-limit code path untested (same wrapper); coded problems skip vukem's instance/request_id members (house pattern already does).
