import type { Config } from "jest";

const config: Config = {
	preset: "jest-expo",
	testMatch: ["**/*.test.ts", "**/*.test.tsx"],
	setupFiles: ["<rootDir>/jest.setup.ts"],
	moduleNameMapper: {
		"\\.css$": "<rootDir>/src/test/empty-module.ts",
		"^expo-font$": "<rootDir>/src/test/expo-font.ts",
	},
	// Bun stores packages under node_modules/.bun/<pkg>/node_modules/<pkg>.
	// The jest-expo ignore list does not see through that, so @react-native/jest-preset stays ESM.
	transformIgnorePatterns: [
		"/node_modules/(?!(.bun/[^/]+/node_modules/)?(.pnpm|react-native|@react-native|@react-native-community|expo|@expo|@expo-google-fonts|react-navigation|@react-navigation|nativewind|react-native-css-interop))",
		"/node_modules/react-native-reanimated/plugin/",
		"/node_modules/@react-native/babel-preset/",
	],
};

export default config;
