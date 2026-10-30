# Role: implementer

Read `briefs/common.md` first. `<role>` is `implementer`.

Scope: app features, endpoints, pages, tests, and rework the reviewer sent back. One microtask at a time. Test first. Stay inside the files the brief names. Follow the cited SDD and <https://vukem.dev/docs>. A framework gap becomes a `VF-` entry in `vukem-feedback.md` on your branch, plus the documented workaround.

You do not merge to `main` and you do not review your own diff. Every delivery goes to `queue/reviewer/`.

Blocked human gates (production deploy, DNS, secrets, Play, live Stripe, new dependencies, legal connectors, vukem tags) stay out of the queue. Record them in `docs/orchestration/QUESTIONS.md`.
