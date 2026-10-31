# FIX-community-create-ux create/merge UX
status: concluida
branch: agent/senior/FIX-community-create-ux @ c186bcf
files: app/community/status.go, tests/community_status_test.go, resources/js/pages/Community/{SuggestCreate,SuggestionStatus}(.test).tsx, resources/js/pages/{Catalog,Search}/Index.tsx, resources/js/components/community/{ItemPicker,SuggestNewLink}(.test).tsx, resources/lang/{en,pt-BR,es}.json
acceptance:
- `go test -tags integration -run 'TestCommunityStatusNames|TestCommunitySubmitWithdrawFlow' ./tests/` -> ok (suggestion.names.<entity id> = default_name)
- `bunx vitest run resources/js/pages/Community resources/js/components/community` -> 4 files / 13 tests pass (merge picks via SuggestBox, submit disabled until both set; names in diff; SuggestNewLink href)
- tsc clean, biome clean, golangci-lint ./app/community 0 issues, go test ./tests/i18n ok
validation: browser not run (QA-community-e2e row covers it; community-create.spec still types ids in the merge form and needs updating, see risks)
changelog: Changed: merge/duplicate form picks records with the search box instead of raw ids; relation and merge diffs on the suggestion status page show entity/item names; Added: "Suggest a new <type>" links on section listings and empty search results.
roadmap: no box (follow-up of P4.A.3.2/3.3)
risks / follow-ups:
- qa/e2e/specs/community-create.spec.ts may fill the old id inputs of the merge form; QA-community-e2e row must adjust it.
- Link shows on every listing even with the flag off (the page shows the closed notice); intentional, no flag in shared props.
- Names come from entities.default_name / items.original_title, not localized.
