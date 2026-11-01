import { expect, test } from "bun:test";

const read = async (path: string): Promise<{ version: string }> => Bun.file(new URL(path, import.meta.url)).json();

test("package.json, jsr.json and the OpenAPI contract carry the same version", async () => {
	const npm = await read("../package.json");
	const jsr = await read("../jsr.json");
	const contract = (await read("../../../openapi/api-v1.json")) as unknown as { info: { version: string } };
	expect(jsr.version).toBe(npm.version);
	expect(npm.version).toBe(contract.info.version);
});

test("the README documents the quota headers, not the legacy RateLimit ones", async () => {
	const readme = await Bun.file(new URL("../README.md", import.meta.url)).text();
	expect(readme).toContain("x-quota-remaining");
	expect(readme.toLowerCase()).not.toContain("ratelimit-");
});
