# Review P2.B.1.5

Re-run and confirm:

- `.github/workflows/ci.yml` jobs `e2e` and `lighthouse` need `build`, download `image.tar`, and contain no `docker build`
- `scripts/ci/qa-up.sh` uses `--no-build` and compose profile `app`
- `bunx playwright test -c qa/e2e/playwright.config.ts --list` shows desktop and mobile for en, pt-BR and es
- `lighthouserc.json` `collect.url` has the five paths from `qa/lighthouse/urls.txt`
- failure upload includes `qa/e2e/test-results/`

Report: `.orchestration/reports/P2.B.1.5.md` (PRONTO_PARA_MERGE).
