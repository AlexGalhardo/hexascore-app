import { expect, test } from "bun:test";

type Job = { steps: { name?: string; run?: string; "working-directory"?: string }[] };

const workflow = Bun.YAML.parse(await Bun.file(new URL("../../.github/workflows/ci.yml", import.meta.url)).text()) as {
	jobs: Record<string, Job>;
};
const script = (workflow.jobs.sdk?.steps ?? []).map((s) => s.run ?? "").join("\n");

test("ci.yml has an sdk job that regenerates both SDKs and fails on a diff", () => {
	expect(workflow.jobs.sdk).toBeDefined();
	expect(script).toContain("bun run sdk:generate");
	expect(script).toContain("bun run sdk:generate:python");
	expect(script).toMatch(/git diff --exit-code[^\n]*sdks/);
});

test("the sdk job type-checks, tests and dry-run packages both SDKs", () => {
	for (const cmd of ["bun run typecheck", "bun test", "npm pack --dry-run", "jsr@0.13.5 publish --dry-run", "uv run pytest", "uv build", "twine@6.2.0 check"]) {
		expect(script).toContain(cmd);
	}
});
