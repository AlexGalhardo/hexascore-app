import type { UserConfig } from "@commitlint/types";

const config: UserConfig = {
	extends: ["@commitlint/config-conventional"],
	// Orchestration lands each microtask with `merge: <id> ...`. That prefix is
	// not a Conventional Commit type; the commits stay in history.
	ignores: [(commit) => commit.startsWith("merge:")],
	rules: {
		"scope-case": [2, "always", "lower-case"],
		"header-max-length": [2, "always", 100],
		// This branch already has bodies longer than the conventional 100.
		// Those commits stay. The header cap above still applies.
		"body-max-line-length": [0],
	},
};

export default config;
