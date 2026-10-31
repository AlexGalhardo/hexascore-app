import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

export default defineConfig({
	plugins: [react()],
	test: {
		environment: "jsdom",
		// Vitest 5 gives each worker a fixed 60 s to start (not configurable); jsdom
		// startup exceeds it when every core boots a worker on a loaded host, so cap the pool.
		maxWorkers: 4,
		include: ["resources/js/**/*.test.{ts,tsx}", "packages/**/*.test.{ts,tsx}", "scripts/**/*.test.ts"],
		// These files import bun:test. CI already runs them with `bun test`
		// (migration policy, projected-spend workflow, app-client package).
		// Setting exclude replaces Vitest's defaults, so node_modules stays out.
		exclude: [
			"**/node_modules/**",
			"**/dist/**",
			"scripts/ci/migration-policy.test.ts",
			"scripts/ops/projected-spend.test.ts",
			"packages/app-client/**",
		],
	},
});
