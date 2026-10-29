# REV-P1.G.4.2
STATUS: CORRIGIDO_E_APROVADO (rodada 0). Reviewed b9c2432 (already in main). Final branch: agent/especialista/P1.G.4.2-fix (from origin/main).
Fixes:
1. Merge conflicts (self, already-redirected, type mismatch) answered 422; SDD 01 §7.5 and app/catalog/merge.go comment say 409 -> now 409; test updated (b180cb4).
2. Affiliate upsert audit lost the previous `active` value -> audited as Before (2nd commit).
Checked ok: all new routes gated (bans.Authorize + CSRF), member/visitor 404 covered; https-only URL check + store allowlist; uuid path params; FOR UPDATE + tx audit; html/template escaping; ReportButton removal justified (item reports stub removed earlier, CHANGELOG Removed).
Re-run: `go test -tags integration ./tests -run 'TestSecurity_A01_AdminPagesNotFoundForMembers|TestAdminMerge|TestAffiliateLinks'` -> ok; golangci-lint (integration) app/admin, tests -> 0 issues.
Nits (no action): mapItemErr matches FK name by string; social.report.* i18n keys now unused until list/user report UI lands.
