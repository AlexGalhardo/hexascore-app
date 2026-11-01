import { expect, test } from "bun:test";
import { chmodSync, mkdtempSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

const script = new URL("../../qa/schemathesis/run.sh", import.meta.url).pathname.replace(/^\/([A-Za-z]:)/, "$1");

// A fake docker on PATH prints the arguments run.sh would pass to the scan container.
function dockerArgs(base: string): string[] {
	const bin = mkdtempSync(join(tmpdir(), "hx-fake-docker-"));
	writeFileSync(join(bin, "docker"), '#!/usr/bin/env bash\nprintf "%s\\n" "$@"\n');
	chmodSync(join(bin, "docker"), 0o755);
	const sep = process.platform === "win32" ? ";" : ":";
	const run = Bun.spawnSync(["bash", script, "openapi/api-v1.json"], {
		env: { ...process.env, PATH: `${bin}${sep}${process.env.PATH}`, HX_TEST_KEY: "hx_live_x", QA_BASE_URL: base },
		timeout: 20_000,
	});
	expect(run.exitCode).toBe(0);
	return run.stdout.toString().split("\n");
}

// Production CanonicalHost 301s any other Host to APP_URL, which inside the scan
// container is its own loopback: every call then fails with "connection refused".
test("a loopback base URL is reached through host.docker.internal with the original Host", () => {
	const args = dockerArgs("http://127.0.0.1:3000");
	expect(args[args.indexOf("--url") + 1]).toBe("http://host.docker.internal:3000");
	expect(args).toContain("Host: 127.0.0.1:3000");
});

// api-v1 paths are relative to the API host's /v1, so the app-host base keeps its /api/v1 prefix.
test("the base path survives the host.docker.internal rewrite", () => {
	const args = dockerArgs("http://127.0.0.1:3000/api/v1");
	expect(args[args.indexOf("--url") + 1]).toBe("http://host.docker.internal:3000/api/v1");
	expect(args).toContain("Host: 127.0.0.1:3000");
});

test("a non-loopback base URL is used as is, without a Host override", () => {
	const args = dockerArgs("https://staging.example.test");
	expect(args[args.indexOf("--url") + 1]).toBe("https://staging.example.test");
	expect(args.some((a) => a.startsWith("Host:"))).toBe(false);
});
