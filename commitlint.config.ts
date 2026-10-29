import type { UserConfig } from "@commitlint/types";

const config: UserConfig = {
	extends: ["@commitlint/config-conventional"],
	rules: {
		"scope-case": [2, "always", "lower-case"],
		"header-max-length": [2, "always", 100],
	},
};

export default config;
