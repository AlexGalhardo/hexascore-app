# REV-P2.A.1.1 (rodada 1) — CORRIGIDO_E_APROVADO
Final branch: agent/especialista/P2.A.1.1-fix (c5bbbe2, on top of estagiario 2919706).
Fixes: (1) script committed 100644 → git mode 100755 (report claimed +x; `scripts/infra/region-latency.sh ...` failed on Linux/CI). (2) curl `-s` → `-fsS`: a 5xx/4xx was recorded as a latency sample. (3) ADR used Portuguese `[NÃO VERIFICADO]` in an English doc → `[NOT VERIFIED]`.
Re-run:
- `SAMPLES=5 scripts/infra/region-latency.sh https://app-production-785f.up.railway.app/up http://127.0.0.1:1/none` → prod `5 0.039359 0.221255 0.214910 0.430346 0`; dead URL `0 nan nan nan nan 5`
- `bash -n` → OK; `go build ./...` → OK (no Go change, lint not needed)
- `bunx markdownlint-cli2@0.24.0 ...` → `Linting: 67 files` / `Summary: 0 issues in 0 files` (config glob lints the repo; conclusive)
Partial acceptance accepted: script + addendum are correct; two-region comparison, 3 probes and region choice stay owner-gated (scratch Railway project). ROADMAP checkbox must stay unchecked.
