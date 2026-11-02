# Contributing to HexaScore

Thank you for considering a contribution. HexaScore is a Go application on the [vukem](https://vukem.dev) framework, with an Inertia + React (SSR) web frontend and an Expo Android app in `apps/mobile`.

## Before you start

- Read the design documents in [docs/sdd/](docs/sdd/) and the decisions in [docs/adrs/](docs/adrs/) and [docs/brainstorming.md](docs/brainstorming.md). Decisions are final until they are edited: change the decision first, then the SDD, then the code.
- For anything bigger than a small fix, open an issue first so the approach can be agreed.
- Framework facts come from <https://vukem.dev/docs>. If vukem lacks something, report it on the [vukem repository](https://github.com/AlexGalhardo/vukem) instead of working around it silently.

## Local setup

1. Run the setup script for your system: `setups/setup-windows.ps1` or `setups/setup-unix.sh`. It installs the tools, creates `.env` from [.env.example](.env.example), generates `APP_KEY`, starts the Docker Compose services (Postgres 18.6 on `127.0.0.1:5433`, RustFS), runs `bun install` and the migrations, and enables the commit-msg hook.
2. Start the app with `vukem serve` (Go server with Vite HMR).
3. Optional integrations (analytics, Turnstile, Telegram, IndexNow) stay off while their variables are unset.

Setup guides, step by step:

- [Windows 11 with Docker](docs/setups/setup-local-with-docker-windows.md)
- [Linux and macOS with Docker](docs/setups/setup-local-with-docker-unix.md)
- [Windows 11 without Docker](docs/setups/setup-local-without-docker-windows.md)
- [Linux and macOS without Docker](docs/setups/setup-local-without-docker-unix.md)

Self-hosting: [Ubuntu 24.04 VPS guide](docs/setups/setup-vps-ubuntu.md) and its [setup script](docs/setups/setup-vps-script.sh), an alternative to Railway.

## Making a change

- Write the test first and keep the change focused on one concern.
- Go: standard library plus `go-cmp`, table-driven tests, no testify. Integration tests carry `//go:build integration` and read `TEST_DATABASE_URL`.
- Security-relevant behaviour needs a test named `TestSecurity_<Axx>_<Behaviour>`; security defaults are never weakened.
- UI strings live in `resources/lang/en.json`, `pt-BR.json` and `es.json`, and the three stay in sync.
- No new Go or npm dependency without prior agreement; versions are pinned exactly.
- Update the relevant SDD and the documentation in the same change when behaviour, commands or conventions change.

## Checks

Run the checks that cover your change before you push:

| Command | What it checks |
| --- | --- |
| `golangci-lint run --build-tags integration ./...` | Go lint and formatting |
| `go test ./...` and `go test -tags integration ./...` | Go unit and integration tests |
| `bun run lint`, `bun run typecheck`, `bun run test` | web lint, types and Vitest |
| `bun run --cwd apps/mobile lint`, `typecheck`, `test`, `bundle:check` | Android app |
| `bunx markdownlint-cli2@0.24.0 "**/*.md"` | Markdown |
| `scripts/ci/local.sh` | every CI job locally (Docker required) |

## Commits

Use [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) in English, in the style of the [vukem history](https://github.com/AlexGalhardo/vukem/commits/main/): `type(scope): lowercase imperative subject`, header at most 100 characters.

- Types: `feat`, `fix`, `perf`, `refactor`, `docs`, `test`, `build`, `ci`, `chore`, `style`, `revert`.
- Scopes: `auth`, `catalog`, `ratings`, `lists`, `billing`, `api`, `admin`, `ui`, `infra`, `deps`, `repo`, `docs`.
- A user-visible change is `feat`, `fix`, `perf` or `refactor`: release notes are generated from these subjects.

The commit-msg hook checks the header locally and commitlint checks it in CI.

## Pull requests

1. Fork the repository and create a branch from `main`.
2. Make one logical change per commit, with tests.
3. Open a pull request against `main` describing what changed and how you tested it.
4. CI must be green before review.

## Reporting bugs and security issues

- Bugs and feature ideas: open a [GitHub issue](https://github.com/AlexGalhardo/hexascore-app/issues) with steps to reproduce.
- Security vulnerabilities: do not open a public issue; follow [SECURITY.md](.github/SECURITY.md).

## License

By contributing you agree that your contributions are licensed under the [MIT license](LICENSE).
