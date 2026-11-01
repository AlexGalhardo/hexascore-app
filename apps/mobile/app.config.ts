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
	plugins: [
		"expo-router",
		"expo-secure-store",
		"expo-sqlite",
		"expo-image",
		// Fonts are embedded at build time (no runtime load, no flash of fallback text).
		// Static OFL instances: Inter 4.1 (rsms/inter release), Playfair Display 700 (Google Fonts).
		[
			"expo-font",
			{
				android: {
					fonts: [
						{
							fontFamily: "Inter",
							fontDefinitions: [
								{ path: "./assets/fonts/Inter-Regular.ttf", weight: 400 },
								{ path: "./assets/fonts/Inter-SemiBold.ttf", weight: 600 },
								{ path: "./assets/fonts/Inter-Bold.ttf", weight: 700 },
							],
						},
						{
							fontFamily: "Playfair Display",
							fontDefinitions: [{ path: "./assets/fonts/PlayfairDisplay-Bold.ttf", weight: 700 }],
						},
					],
				},
			},
		],
	],
	experiments: {
		typedRoutes: true,
	},
	web: {
		bundler: "metro",
	},
};

export default config;
