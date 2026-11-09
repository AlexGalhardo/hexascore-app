# Framework feedback rule

This project is the battle test for vukem 1.0. Every vukem bug, missing feature, awkward API, security issue or wrong/missing docs page you hit gets one entry in the single shared `vukem-feedback.md` at the workspace root (the folder that holds `vukem/` and every app, outside every repo). Never create a copy inside a repo. Other projects use the `VE-`, `TR-`, `ST-`, `MU-` and `TL-` ids.

- Cite the docs you consulted with <https://vukem.dev/docs> URLs. Mark gaps `[SITE PAGE MISSING]` (page does not exist) or `[NOT IN VUKEM DOCS]` (page exists, the API or behaviour is not there). Never invent a vukem API.
- Entry: `## HS-<n>: <title>` (next free number for the `HS-` prefix) with the fields `Severity`, `Date`, `Docs`, `Finding`, `Minimal repro`, `Suggested fix`, `Status` (`open` until the vukem maintainer fixes it on the release PR and updates it).
- Apply a local workaround and mark it in code with `// workaround(vukem): <entry title>` (or the `HS-<n>` id), then keep going.
- Never edit the vukem framework checkout from here.
