STATUS: concluida
REV: PRONTO_PARA_MERGE
Branch: agent/implementer/FIX-page-script-loader

The inline loader ended in `})})})`, a SyntaxError. `open()` then waited until the 120s test timeout on every page with a form. Signup HTML on :3000 and :3200 never requested `/build/assets/app-*.js`.

Command: go test ./app/http/middleware -count=1 -run TestMovePageScript
Output: ok github.com/AlexGalhardo/hexascore-app/app/http/middleware 0.031s

node --check on the corrected loader exits 0. The old string exits 1 with Unexpected token ')'.

Risks: the containers still run the old binary. The signup e2e on :3200 stays red until a new process serves this commit. P1.I.7.1 stays open until that suite is green.
