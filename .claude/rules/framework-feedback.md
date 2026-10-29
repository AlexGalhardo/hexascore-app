# Framework feedback rule

This project is the battle test for vukem 1.0. Every vukem bug, missing feature, awkward API, security issue or wrong/missing docs page you hit gets one entry in `vukem-feedback.md` at the repo root (format in that file's header).

- Cite the docs you consulted with <https://vukem.dev/docs> URLs. Mark gaps `[SITE PAGE MISSING]` (page does not exist) or `[NOT IN VUKEM DOCS]` (page exists, the API or behaviour is not there). Never invent a vukem API.
- Agents never pick a VF number. Title the entry `## VF-NEW-<task-id>: <title>` and add an index row `| VF-NEW-<task-id> | <date> | <severity> | open |`; the orchestrator assigns the real number at merge.
- Apply a local workaround and mark it in code with `// workaround(vukem): <entry title>` (or the `VF-NEW-<task-id>` id), then keep going.
- Never edit the vukem framework checkout from here.
