import { expect, test } from "bun:test";

const script = new URL("./local.sh", import.meta.url).pathname.replace(/^\/([A-Za-z]:)/, "$1");

// openapi and lighthouse need the image of the build job; without build selected they used to wait forever.
test.each([
	["--only", "openapi"],
	["--only", "lighthouse,mobile"],
	["--skip", "build"],
])("local.sh %s %s fails fast without the build job", (flag, jobs) => {
	const run = Bun.spawnSync(["bash", script, flag, jobs], { timeout: 20_000 });
	expect(run.exitCode).toBe(2);
	expect(run.stderr.toString()).toContain("needs the build job");
});
