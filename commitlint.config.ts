import type { UserConfig } from "@commitlint/types";

const config: UserConfig = {
	extends: ["@commitlint/config-conventional"],
	// Orchestration lands each microtask with `merge: <id> ...`. That prefix is
	// not a Conventional Commit type; the commits stay in history.
	ignores: [(commit) => commit.startsWith("merge:")],
	rules: {
		"scope-case": [2, "always", "lower-case"],
		"header-max-length": [2, "always", 100],
	},
};

export default config;
