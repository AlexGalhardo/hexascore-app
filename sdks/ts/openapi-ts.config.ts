import { defineConfig } from "@hey-api/openapi-ts";

// Generated from the committed contract; never edit src/gen by hand — run `bun run sdk:generate`.
export default defineConfig({
	input: "../../openapi/api-v1.json",
	// Relative imports carry the `.ts` extension: JSR publishes the sources as they are and needs it, and the npm
	// build rewrites it to `.js` (`rewriteRelativeImportExtensions` in tsconfig.json).
	output: { path: "src/gen", importFileExtension: ".ts" },
	plugins: ["@hey-api/client-fetch", "@hey-api/typescript", "@hey-api/sdk"],
});
