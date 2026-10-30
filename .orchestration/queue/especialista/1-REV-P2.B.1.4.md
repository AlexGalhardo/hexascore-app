# Review P2.B.1.4

Re-run and confirm:

- `bash scripts/openapi-check.sh` → exit 0
- `bunx redocly lint openapi/api-v1.json openapi/app-v1.json` → exit 0
- `.github/workflows/ci.yml` job `openapi` needs `build`, downloads `image.tar`, and has no `docker build`
- `qa/schemathesis/ci.sh` is the Schemathesis step; `--network host` is Linux-only, so a Windows host cannot run it as written

Report: `.orchestration/reports/P2.B.1.4.md` (PRONTO_PARA_MERGE).
