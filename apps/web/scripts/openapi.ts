import { readFile, writeFile } from "node:fs/promises";
import { routing } from "../i18n/routing";
import { serializeOpenApi } from "../lib/api/openapi";
import { openApiPath } from "../lib/api/spec-path";

// `bun scripts/openapi.ts` writes public/openapi.json (en) and openapi.<locale>.json; `--check` exits 1 when any
// committed file is stale.
const check = process.argv.includes("--check");
let stale = false;

for (const locale of routing.locales) {
	const path = openApiPath(locale);
	const target = new URL(`../public${path}`, import.meta.url);
	const next = serializeOpenApi(locale);
	if (check) {
		const current = await readFile(target, "utf8").catch(() => "");
		if (current.replace(/\r\n/g, "\n") !== next) {
			console.error(`apps/web/public${path} is out of date. Run \`bun run openapi\` and commit the result.`);
			stale = true;
		}
	} else {
		await writeFile(target, next);
		console.info(`Wrote ${target.pathname}`);
	}
}

if (stale) process.exit(1);
if (check) console.info("OpenAPI documents are up to date.");
