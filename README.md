<div align="center">

<h1><a href="https://hexascore.app" target="_blank">HexaScore.app</a></h1>

[![CI](https://github.com/AlexGalhardo/hexascore-app/actions/workflows/ci.yml/badge.svg)](https://github.com/AlexGalhardo/hexascore-app/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go 1.27.3](https://img.shields.io/badge/go-1.27.3-00ADD8?logo=go&logoColor=white)](go.mod)

Movies, TV Shows, Books and Games Open Source Almanac

</div>

## Introduction

HexaScore is an open source almanac and community score platform for games, movies, TV shows and books. Each item is rated by the community in six categories per type, and the ratings combine into a published score. The application is built on the [vukem](https://vukem.dev) Go framework.

## Summary

- [Introduction](#introduction)
- [Stability](#stability)
- [Contributing](#contributing)
- [Code of conduct](#code-of-conduct)
- [Security vulnerabilities](#security-vulnerabilities)
- [License](#license)

## Stability

HexaScore follows [Semantic Versioning](https://semver.org) and tags releases as `vX.Y.Z` on `main`. The first release is `v0.1.0`. `v1.0.0` marks parity with the legacy application and a successful cutover ([ADR 0013](docs/adrs/0013-versioning-releases-cutover.md)).

Changes are listed in [CHANGELOG.md](CHANGELOG.md), which follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## Contributing

Thank you for considering a contribution to HexaScore.

- Read the design documents in `docs/` (SDDs in `docs/sdd/`, decisions in `docs/adrs/`) before changing behaviour.
- Write the test first. Run the specific test, lint and typecheck that cover your change.
- Security-relevant behaviour needs a test named `TestSecurity_<Axx>_<Behaviour>`.
- Update the relevant SDD, `CHANGELOG.md` (under `[Unreleased]`) and the documentation in the same change when behaviour, commands or conventions change.

### Commits

Use [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) in English: `type(scope): description`.

- Types: `feat`, `fix`, `perf`, `refactor`, `docs`, `test`, `build`, `ci`, `chore`, `style`, `revert`.
- Scopes: `auth`, `catalog`, `ratings`, `lists`, `billing`, `api`, `admin`, `ui`, `infra`, `deps`, `repo`, `docs`.

The commit-msg hook enabled by the setup script checks the header locally, and commitlint checks the commits and the PR title in CI.

### Pull requests

1. Create a branch from `main`.
2. Make one logical change per commit.
3. Open a pull request against `main`.

`main` accepts changes only through a pull request with green required checks and linear history.

The full guide (setup, checks, commit style) is in [CONTRIBUTING.md](CONTRIBUTING.md).

## Code of conduct

Keep discussions respectful and constructive. This repository does not publish a separate code of conduct file yet.

## Security vulnerabilities

Please do not open a public issue for a security vulnerability. Report it privately as described in [SECURITY.md](SECURITY.md).

## License

HexaScore is open source software released under the [MIT license](LICENSE).
