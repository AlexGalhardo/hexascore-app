# SDK release

How to publish the Hexascore SDKs: `@hexascore/sdk` (TypeScript, folder `sdks/ts`) to npm and JSR, and `hexascore` (Python, folder `sdks/python`) to PyPI. Publishing is automated by `.github/workflows/sdk-release.yml` through trusted publishing (OIDC): no registry token is stored in the repository, in secrets or on a laptop.

| | npm | JSR | PyPI |
| --- | --- | --- | --- |
| Package | `@hexascore/sdk` | `@hexascore/sdk` | `hexascore` |
| Folder | `sdks/ts` | `sdks/ts` | `sdks/python` |
| What is published | `dist/` (ESM + `.d.ts`), README, LICENSE, `package.json` | TypeScript source `src/**/*.ts` without tests, README, LICENSE, `jsr.json` | sdist and wheel built by `uv build` |
| Version field | `package.json` | `jsr.json` | `pyproject.toml` |
| Trust | npm trusted publisher + provenance | JSR repository link (OIDC) | PyPI trusted publisher |
| GitHub environment | `npm` | `jsr` | `pypi` |

Published versions are immutable on all three registries. A mistake means a patch bump.

## Who does what

Everything under "One-time owner setup" is owner-only (accounts, 2FA, registry settings). The workflow cannot publish until it is done, and it never publishes unless someone starts it by hand with `publish` ticked. The rehearsal on TestPyPI and the first real release are owner-triggered.

## One-time owner setup

### npm

1. Create the account at <https://www.npmjs.com/signup>, confirm the e-mail, and enable 2FA at *Account → Two-Factor Authentication* with `auth-and-writes`.
2. Create the organization `hexascore` (free plan, public packages) at <https://www.npmjs.com/org/create>. Without it `npm publish` answers `404 Scope not found`.
3. The package must exist before a trusted publisher can be attached. Publish version `1.0.0` once from a laptop (`cd sdks/ts && npm publish --otp=<code>`) or ask the maintainers for a bootstrap release, then open the package *Settings → Trusted Publisher* and set repository `AlexGalhardo/hexascore-app`, workflow file `sdk-release.yml`, environment `npm`.
4. Afterwards set the package publishing access to "Require two-factor authentication and disallow tokens".

### JSR

1. Sign in at <https://jsr.io> with GitHub (no password).
2. Create the scope `hexascore` and the package `sdk` at <https://jsr.io/new>.
3. At `jsr.io/@hexascore/sdk` → *Settings → GitHub repository*, link `AlexGalhardo/hexascore-app`. A job with `id-token: write` in that repository then publishes without a token and JSR records provenance.
4. Fill the description and the runtime compatibility (Deno, Node, Bun, browsers, Workers) in *Settings*.

### PyPI and TestPyPI

Two separate registries with separate accounts: <https://test.pypi.org/account/register/> and <https://pypi.org/account/register/>. In each:

1. Create the account, confirm the e-mail, enable 2FA.
2. Add a pending trusted publisher (it works before the project exists, so no account-wide token is ever created): *Publishing → Add a new pending publisher → GitHub* with owner `AlexGalhardo`, repository `hexascore-app`, workflow `sdk-release.yml`, environment `pypi`, project name `hexascore`.

### GitHub environments

In the repository *Settings → Environments* create `npm`, `jsr` and `pypi`. Add the owner as a required reviewer on each, so every publish job waits for an explicit approval. Deployment branches: `main` only.

## Release procedure

1. On `main`, bump the SDK version in the three files, keeping them identical (SemVer; the SDK version is independent of the site version):
   - `sdks/ts/package.json` (`npm version X.Y.Z --no-git-tag-version`)
   - `sdks/ts/jsr.json`
   - `sdks/python/pyproject.toml` (`uv version X.Y.Z`)
2. Regenerate and check nothing drifts: `bun run openapi:check`, `bun run sdk:generate`, `bun run sdk:generate:python`, then `git status` must show no unexpected diff. CI job `sdk` enforces the same.
3. Merge. In *Actions → SDK release → Run workflow* choose the commit and leave `publish` unticked. The `verify` job checks that the three versions agree, then runs typecheck, tests, build, `npm pack --dry-run`, `npm publish --dry-run`, `jsr publish --dry-run`, `uv build` and `twine check`. It must be green.
4. Rehearse on TestPyPI: run the workflow again with `publish` ticked and `pypi_index = testpypi`. This also runs the npm and JSR publish jobs, so for a pure rehearsal approve only the `pypi` environment and reject the `npm` and `jsr` approvals. Check <https://test.pypi.org/project/hexascore/> (README rendered, version, links) and install from it:

   ```sh
   uv run --no-project --isolated \
     --index https://test.pypi.org/simple/ --index-strategy unsafe-best-match \
     --with hexascore==X.Y.Z python -c "import hexascore; print('ok')"
   ```

5. Real release: run the workflow with `publish` ticked and `pypi_index = pypi`. Approve `npm`, `jsr` and `pypi`.
6. Verify the three installs:

   ```sh
   npm view @hexascore/sdk version dist-tags
   bunx jsr info @hexascore/sdk
   uv run --no-project --isolated --with hexascore==X.Y.Z python -c "import hexascore; print('ok')"
   ```

7. Tag the commit `sdk-vX.Y.Z` and push the tag. Run the staging smoke tests (`qa/sdk-smoke/`, P3.B.2.2) and record the release in `CHANGELOG.md`.

## Local rehearsal (no network, no login)

```sh
cd sdks/ts
bun install --frozen-lockfile --ignore-scripts
bun run typecheck && bun test && bun run build
npm pack --dry-run                         # only dist/**, README, LICENSE, package.json
bunx jsr@0.13.5 publish --dry-run --allow-dirty   # ends with "Success Dry run complete"
cd ../python
uv run pytest -q && rm -rf dist && uv build && uvx twine@6.2.0 check dist/*
```

## Common errors

| Registry | Error | Cause and fix |
| --- | --- | --- |
| npm | `404 Scope not found` | The `hexascore` organization does not exist or the publisher is not a member |
| npm | `403 You cannot publish over the previously published versions` | The version exists; bump it |
| npm | `ENEEDAUTH` or OIDC token exchange fails in CI | Trusted publisher missing or mismatched (repository, workflow file name, environment `npm`), or npm older than 11.5.1 |
| JSR | `Aborting due to uncommitted changes` | Dirty tree on a local run; commit or use `--allow-dirty` for the rehearsal only |
| JSR | Slow types error on an exported symbol | Add an explicit type on the public API |
| JSR | Not authorized in CI | Repository not linked in the package settings |
| PyPI | `400 File already exists` | The version was already uploaded; bump it and rebuild |
| PyPI | `invalid-publisher` / trusted publishing exchange failed | Publisher entry does not match owner, repository, workflow `sdk-release.yml` and environment `pypi` exactly |
| PyPI | README renders as plain text | Invalid Markdown for PyPI; `twine check` shows it |

## Fixing a bad release

- npm: within 72 hours and with no dependents, `npm unpublish @hexascore/sdk@X.Y.Z`; otherwise `npm deprecate @hexascore/sdk@X.Y.Z "use X.Y.Z+1"`. The number stays burned.
- JSR: versions cannot be deleted; use *Yank* at `jsr.io/@hexascore/sdk/versions`.
- PyPI: use *Manage → Releases → Options → Yank*. Deleting also burns the number.

In all cases publish the corrected patch version.
