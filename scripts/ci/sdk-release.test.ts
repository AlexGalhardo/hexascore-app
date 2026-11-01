import { expect, test } from "bun:test";

type Step = { uses?: string; run?: string; with?: Record<string, string> };
type Job = {
	if?: string;
	needs?: string | string[];
	environment?: string;
	permissions?: Record<string, string>;
	steps: Step[];
};

const text = await Bun.file(new URL("../../.github/workflows/sdk-release.yml", import.meta.url)).text();
const workflow = Bun.YAML.parse(text) as {
	on: Record<string, { inputs?: Record<string, { type: string; default?: unknown }> }>;
	permissions?: Record<string, string>;
	jobs: Record<string, Job>;
};
const script = (job: string): string => (workflow.jobs[job]?.steps ?? []).map((s) => s.run ?? "").join("\n");

test("the workflow is manual only and has no long-lived tokens", () => {
	expect(Object.keys(workflow.on)).toEqual(["workflow_dispatch"]);
	expect(text).not.toMatch(/secrets\.(?!GITHUB_TOKEN)/);
	expect(text).not.toMatch(/NPM_TOKEN|NODE_AUTH_TOKEN|UV_PUBLISH_TOKEN|PYPI_TOKEN/);
});

test("publishing is off by default and every publish job is gated on it", () => {
	const inputs = workflow.on.workflow_dispatch?.inputs ?? {};
	expect(inputs.publish?.type).toBe("boolean");
	expect(inputs.publish?.default).toBe(false);
	for (const job of ["publish-npm", "publish-jsr", "publish-pypi"]) {
		expect(workflow.jobs[job]?.if).toContain("inputs.publish");
		expect(workflow.jobs[job]?.needs).toBe("verify");
	}
});

test("only publish jobs get id-token: write, each behind its own environment", () => {
	expect(workflow.permissions?.["id-token"]).toBeUndefined();
	expect(workflow.jobs.verify?.permissions?.["id-token"]).toBeUndefined();
	const envs: Record<string, string> = { "publish-npm": "npm", "publish-jsr": "jsr", "publish-pypi": "pypi" };
	for (const [job, env] of Object.entries(envs)) {
		expect(workflow.jobs[job]?.permissions?.["id-token"]).toBe("write");
		expect(workflow.jobs[job]?.environment).toBe(env);
	}
});

test("verify rehearses every registry without publishing", () => {
	const s = script("verify");
	for (const cmd of ["npm pack --dry-run", "npm publish --dry-run", "jsr@0.13.5 publish --dry-run", "uv build", "twine@6.2.0 check"]) {
		expect(s).toContain(cmd);
	}
});

test("each registry publishes through OIDC", () => {
	expect(script("publish-npm")).toContain("publish --provenance --access public");
	expect(script("publish-jsr")).toContain("jsr@0.13.5 publish");
	expect(script("publish-pypi")).toContain("uv publish --trusted-publishing always");
	expect(script("publish-pypi")).toContain("https://test.pypi.org/legacy/");
});

test("every action is pinned by commit SHA", () => {
	for (const job of Object.values(workflow.jobs)) {
		for (const step of job.steps) {
			if (step.uses) expect(step.uses).toMatch(/@[0-9a-f]{40}$/);
		}
	}
});
