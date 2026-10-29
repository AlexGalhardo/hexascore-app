# ADR 0002: Host on Railway instead of Vercel

Status: accepted (2026-10-28). Decisions: owner requirement 2; brainstorming D-08, A-03.

## Context

Vercel is a serverless platform optimised for Next.js. The Go rewrite needs a long-running process (HTTP server, scheduler, queue workers, Node SSR sidecar) and a Postgres instance close to it. Legacy monthly estimates on Vercel + Neon were about US$52 at launch and US$160 in growth (`legacy/docs/vercel-costs-monthly.md`, 2026-10-03 prices), with Web Analytics growing linearly with traffic. The owner set a ceiling of US$30/month at launch (D-08).

Railway runs containers with persistent volumes, managed Postgres, cron-like scheduling, private networking, buckets and an MCP/CLI the agents can use (`use-railway` skill available in this environment).

## Decision

Deploy on Railway: one project, environments `production` and `staging`, services `app` (Go binary built from a pinned `golang` base image, multi-stage, with the Node SSR process supervised inside the same container until the ceiling is raised), `postgres` (Railway managed Postgres with a volume), and a bucket or S3-compatible store for covers and backups. Region: the Railway region with the lowest measured latency to Brazilian users at implementation time; US East if no South American region exists (A-03). DNS for `hexascore.app`, `www` and `api` points to Railway; TLS by Railway.

## Alternatives considered

- **Keep Vercel for the web and host the Go API elsewhere.** Two platforms, two deploy pipelines, CORS and session complexity. Rejected.
- **Fly.io / Hetzner VPS / Render.** Comparable prices; Railway was the owner's requirement and has first-class tooling in this environment. Not evaluated further.
- **Kubernetes.** Over-engineering for a single-service app under US$30/month. Rejected.

## Consequences

- Positive: one container, predictable cost, Postgres on the private network, no serverless cold starts, cron and workers inside the app.
- Negative: no global edge CDN for HTML by default; static assets must be cached with long `Cache-Control` and content hashes (Vite manifest) and optionally fronted by a CDN later; single region means higher latency for users far from it; a single app instance means deploys cause a short restart unless Railway's zero-downtime deploy is configured (verify in `11-infra-deploy.md`).
- Obligations: Vercel Analytics and Speed Insights are replaced by self-hosted metrics (SDD 12); BotID is replaced (SDD 03/13); Vercel Blob is replaced by a Railway bucket or S3-compatible store with a migration of existing cover files.
