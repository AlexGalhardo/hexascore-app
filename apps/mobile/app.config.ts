import type { ExpoConfig } from "expo/config";

// Expo SDK 58 ships Hermes only. Native projects stay generated (CNG); do not commit android/ or ios/.
const config: ExpoConfig = {
	name: "HexaScore",
	slug: "hexascore",
	scheme: "hexascore",
	version: "1.0.0",
	orientation: "portrait",
	userInterfaceStyle: "automatic",
	android: {
		package: "app.hexascore",
	},
	plugins: ["expo-router"],
	experiments: {
		typedRoutes: true,
	},
	web: {
		bundler: "metro",
	},
};

export default config;
