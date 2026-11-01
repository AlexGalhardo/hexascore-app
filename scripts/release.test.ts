import { execFileSync, spawnSync } from "node:child_process";
import { chmodSync, copyFileSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { afterAll, describe, expect, it } from "vitest";

// release.sh against a scratch repository that reproduces the squash-merge release: the tag sits on a branch
// main never reaches, and main only gets an equal tree through a merge-back. A fake git-cliff records its arguments.
const repo = mkdtempSync(join(tmpdir(), "release-sh-"));
afterAll(() => rmSync(repo, { recursive: true, force: true }));

const git = (...args: string[]): string =>
	execFileSync("git", ["-c", "core.hooksPath=/dev/null", "-c", "commit.gpgsign=false", ...args], {
		cwd: repo,
		encoding: "utf8",
	}).trim();

const commit = (msg: string, file: string): string => {
	writeFileSync(join(repo, file), msg);
	git("add", "-A");
	git("commit", "-q", "-m", msg);
	return git("rev-parse", "HEAD");
};

const release = (...args: string[]) => {
	writeFileSync(join(repo, "bin", "args"), "");
	const run = spawnSync("bash", ["-c", 'PATH="$PWD/bin:$PATH" scripts/release.sh "$@"', "release", ...args], {
		cwd: repo,
		encoding: "utf8",
	});
	return { code: run.status, err: run.stderr, cliff: readFileSync(join(repo, "bin", "args"), "utf8").split("\n") };
};

const changelog = (): string => readFileSync(join(repo, "CHANGELOG.md"), "utf8");
const url = "https://github.com/AlexGalhardo/hexascore-app/compare";

git("init", "-q", "-b", "main");
git("config", "user.email", "release-test@example.test");
git("config", "user.name", "release test");
git("config", "core.autocrlf", "false");
// release.sh commits and tags with the repository's config: no hooks, no signing.
git("config", "core.hooksPath", "/dev/null");
git("config", "commit.gpgsign", "false");
git("config", "tag.gpgsign", "false");
mkdirSync(join(repo, "scripts"));
mkdirSync(join(repo, ".config"));
mkdirSync(join(repo, "bin"));
writeFileSync(join(repo, ".git", "info", "exclude"), "bin/\n");
copyFileSync(new URL("./release.sh", import.meta.url), join(repo, "scripts", "release.sh"));
writeFileSync(join(repo, ".config", "cliff.toml"), "# fake\n");
writeFileSync(
	join(repo, "bin", "git-cliff"),
	[
		"#!/usr/bin/env bash",
		'printf "%s\\n" "$@" >>"$(dirname "$0")/args"',
		'while [ $# -gt 0 ]; do [ "$1" = --tag ] && v=$2; shift; done',
		'printf "## [%s] - 2026-11-01\\n\\n### Added\\n\\n- stub\\n" "${v#v}"',
		"",
	].join("\n"),
);
chmodSync(join(repo, "bin", "git-cliff"), 0o755);
writeFileSync(
	join(repo, "CHANGELOG.md"),
	[
		"# Changelog",
		"",
		"## [Unreleased]",
		"",
		"## [0.2.0] - 2026-10-31",
		"",
		"- two",
		"",
		"## [0.1.0] - 2026-10-30",
		"",
		"- one",
		"",
		`[Unreleased]: ${url}/v0.2.0...HEAD`,
		`[0.2.0]: ${url}/v0.1.0...v0.2.0`,
		`[0.1.0]: https://github.com/AlexGalhardo/hexascore-app/releases/tag/v0.1.0`,
		"",
	].join("\n"),
);
git("add", "-A");
git("commit", "-q", "-m", "feat: one");
git("tag", "-a", "v0.1.0", "-m", "v0.1.0");
commit("feat: two", "two.txt");
// v0.2.0 lives on a rebuilt linear branch; origin got a squash with the same tree, merged back into main.
const tree = git("rev-parse", "HEAD^{tree}");
const linear = git("commit-tree", tree, "-p", "v0.1.0^{commit}", "-m", "chore(release): v0.2.0");
git("tag", "-a", "v0.2.0", "-m", "v0.2.0", linear);
const squash = git("commit-tree", tree, "-p", "v0.1.0^{commit}", "-m", "chore(release): land v0.2.0 (#1)");
git("merge", "-q", "--no-ff", "-m", "chore(release): merge squashed release back into main", squash);
const mergeBack = git("rev-parse", "HEAD");
commit("feat: three", "three.txt");

describe("scripts/release.sh range", () => {
	it("starts after the commit whose tree matches the highest tag, not at git describe", () => {
		expect(git("describe", "--tags", "--abbrev=0")).toBe("v0.1.0");

		const r = release("v0.3.0");
		expect(r.err).toBe("");
		expect(r.code).toBe(0);
		expect(r.cliff).toContain(`${mergeBack}..HEAD`);
		// git-cliff drops the range when --unreleased is also passed.
		expect(r.cliff).not.toContain("--unreleased");
		expect(r.cliff).toContain("v0.3.0");

		const log = changelog();
		expect(log).toMatch(/## \[Unreleased\]\n\n## \[0\.3\.0\] - 2026-11-01\n\n### Added\n\n- stub\n\n## \[0\.2\.0\]/);
		expect(log).toContain(`[Unreleased]: ${url}/v0.3.0...HEAD\n[0.3.0]: ${url}/v0.2.0...v0.3.0\n[0.2.0]: ${url}/v0.1.0...v0.2.0`);
		expect(git("describe", "--tags", "--exact-match", "HEAD")).toBe("v0.3.0");
		expect(git("log", "-1", "--format=%s")).toBe("chore(release): v0.3.0");
	});

	it("uses the tag itself once it is on main's first-parent line", () => {
		const tagged = git("rev-parse", "v0.3.0^{commit}");
		commit("fix: four", "four.txt");

		const r = release("v0.3.1");
		expect(r.code).toBe(0);
		expect(r.cliff).toContain(`${tagged}..HEAD`);
		expect(changelog()).toContain(`[Unreleased]: ${url}/v0.3.1...HEAD\n[0.3.1]: ${url}/v0.3.0...v0.3.1\n`);
	});

	it("takes an explicit --since and refuses a version that is not above the highest tag", () => {
		const three = git("rev-parse", "HEAD~2");
		commit("feat: five", "five.txt");

		const low = release("v0.3.1");
		expect(low.code).toBe(1);
		expect(low.err).toContain("exists");

		const older = release("v0.2.9");
		expect(older.code).toBe(1);
		expect(older.err).toContain("not above v0.3.1");

		const r = release("v0.4.0", "--since", three);
		expect(r.code).toBe(0);
		expect(r.cliff).toContain(`${three}..HEAD`);
	});
});
