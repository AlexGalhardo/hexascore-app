# REV-P1.I.3.1 (rodada 1, reviewer)
VERDICT: CORRIGIDO_E_APROVADO. Final branch agent/reviewer/P1.I.3.1-fix (on agent/senior/P1.I.3.1). Merged --no-ff into local main as 08257c0. Not pushed.
Fix: CHANGELOG copied the k6 line into every release from 0.1.0 through 0.8.0. Moved it to [Unreleased] only. Merge kept the existing P2.F.4.1 Unreleased line.
Re-run: isolated COMPOSE_PROJECT_NAME=hx-reviewer, image hexascore-app:e2e-reviewer, app 127.0.0.1:3010, grafana/k6:2.3.0 --network host, seed.sh rc=0 (10 keys, 10 lists). hx-senior untouched. dev still contains P1.I.5.1-r2 (80f291d).
api-v1 first run rc=99 (cold): list p95=476ms, detail p95=438.71ms, http_req_failed 0/3005, checks 3001/3001. Immediate re-run rc=0: list p95=13.54ms (<300), detail p95=11.93ms (<150), http_req_failed 0/3006.
public-lists rc=0 p95=14.31ms (<400), http_req_failed 0/1801.
pages-ttfb rc=0 TTFB p75 home 20.17 / entity 19.24 / item 11.9 / listing 10.88 ms, http_req_failed 0/2403.
Not run: search.js, rating-hot-item.js (same skip as the delivery; not in the three acceptance commands). Demo catalogue, not the 500k set.
k6 v2.3.0 started. First launcher miss was Git Bash rewriting /qa; rerun with MSYS_NO_PATHCONV=1. That was not a script failure.
