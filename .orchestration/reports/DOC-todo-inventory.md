# DOC-todo-inventory TODO / FIXME / workaround(vukem) inventory
status: concluida (report only; no edits)
branch: none (main @ 246ced8 at read time)
files: .orchestration/reports/DOC-todo-inventory.md
acceptance:
- `rg -n "TODO|FIXME|workaround\(vukem\)" app resources/js internal database` → 7 lines, all `workaround(vukem)`; 0 `TODO`, 0 `FIXME`
validation: no code touched; lint n/a
changelog: none
roadmap: no box
risks / follow-ups: the search covers only the four directories named in the row. `TODO`/`FIXME` anywhere else in the repo was not searched. VF status in `vukem-feedback.md` was not checked here, only the ids cited in code.

| file:line | text (abridged) | matching id |
| --- | --- | --- |
| app\billing\stripe_calls.go:14 | `workaround(vukem): VF-051` contrib/stripe has no subscription list, no checkout session | VF-051 |
| app\controllers\exports_controller.go:54 | `workaround(vukem): VF-016/VF-039` errs.HTTPError has no ... (429 export cooldown) | VF-016, VF-039 |
| app\http\middleware\request_log.go:11 | `workaround(vukem): VF-022` (route and user recorded by inner middleware) | VF-022 |
| app\http\middleware\api_edge.go:25 | `APIEdge is the workaround(vukem) VF-015/VF-016` for the public API edge | VF-015, VF-016 |
| app\http\middleware\admin_guard.go:73 | AdminGuard, SDD 13 FR-13-5, `workaround(vukem):` (id on next line) | not read; check the next line |
| app\controllers\publicapi\items.go:46 | `workaround(vukem): VF-049` typed handlers have no documented way to answer 304 | VF-049 |
| database\legacyimport\twofactor.go:122 | `workaround(vukem): VF-060` totp has no exported hash for ... | VF-060 |

Notes:
- `admin_guard.go:73` was matched on the marker, but the id is on a line the search did not print. Not confirmed here.
- `app\billing\stripe_calls.go` is in the senior-only area for `app/billing`; this report only lists it, it does not edit it.
- ROADMAP ids: none found in these markers.

## Review (especialista, 2026-10-31)

verdict: APROVADO (report only)
re-run on main @ ad5287f: same `rg` → 8 lines, all `workaround(vukem)`, 0 TODO/FIXME. The eighth line is new since 246ced8: `app/controllers/publicapi/items.go:376` `workaround(vukem): VF-NEW-P2.H.3.1` (P2.H.3.1 merge); the orchestrator still has to number that VF-NEW entry
checked: command quoted, scope limited to the four directories as the row asked
