import { readFile, writeFile } from "node:fs/promises";
import { serializeOpenApi } from "../lib/api/openapi";

// `bun scripts/openapi.ts` writes public/openapi.json; `--check` exits 1 when the committed file is stale.
const target = new URL("../public/openapi.json", import.meta.url);
const next = serializeOpenApi();

if (process.argv.includes("--check")) {
	const current = await readFile(target, "utf8").catch(() => "");
	if (current.replace(/\r\n/g, "\n") !== next) {
		console.error("apps/web/public/openapi.json is out of date. Run `bun run openapi` and commit the result.");
		process.exit(1);
	}
	console.info("openapi.json is up to date.");
} else {
	await writeFile(target, next);
	console.info(`Wrote ${target.pathname}`);
}
