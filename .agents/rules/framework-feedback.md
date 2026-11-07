# Framework feedback rule

This project is the battle test for vukem 1.0. Every vukem bug, missing feature, awkward API, security issue or wrong/missing docs page you hit gets one entry in the shared `vukem-feedback.md` at the workspace root (`C:\dev\hexascore-app`, outside every repo; per-repo copies are legacy). Other projects use `VE-`, `TR-` and `ST-` ids.

- Cite the docs you consulted with <https://vukem.dev/docs> URLs. Mark gaps `[SITE PAGE MISSING]` (page does not exist) or `[NOT IN VUKEM DOCS]` (page exists, the API or behaviour is not there). Never invent a vukem API.
- Entry: `## HS-<n>: <title>` (next free number for the `HS-` prefix) with the fields `Severity`, `Date`, `Docs`, `Finding`, `Minimal repro`, `Suggested fix`, `Status` (`open` until the vukem maintainer fixes it on the release PR and updates it).
- Apply a local workaround and mark it in code with `// workaround(vukem): <entry title>` (or the `HS-<n>` id), then keep going.
- Never edit the vukem framework checkout from here.
