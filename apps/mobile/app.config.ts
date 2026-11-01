import type { ExpoConfig } from "expo/config";

// Expo SDK 58 ships Hermes only. Native projects stay generated (CNG); do not commit android/ or ios/.
const config: ExpoConfig = {
	name: "HexaScore",
	slug: "hexascore",
	owner: "alexgalhardo",
	scheme: "hexascore",
	version: "1.0.0",
	orientation: "portrait",
	userInterfaceStyle: "automatic",
	android: {
		package: "app.hexascore",
		// EAS sets versionCode remotely; only the Gradle fallback (.github/workflows/mobile-gradle.yml) passes it.
		...(process.env.ANDROID_VERSION_CODE
			? { versionCode: Number(process.env.ANDROID_VERSION_CODE) }
			: {}),
		// Verified App Links (SDD 10 §7.6); the Go app serves /.well-known/assetlinks.json.
		// Expo loads this file without TS module resolution, so the list is inlined;
		// src/linking.test.ts keeps it equal to APP_LINK_PREFIXES.
		intentFilters: [
			{
				action: "VIEW",
				autoVerify: true,
				data: [
					"/games",
					"/movies",
					"/series",
					"/books",
					"/entities",
					"/lists",
					"/pt-br",
					"/es",
				].map((pathPrefix) => ({
					scheme: "https",
					host: "hexascore.app",
					pathPrefix,
				})),
				category: ["BROWSABLE", "DEFAULT"],
			},
		],
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
	// EAS project @alexgalhardo/hexascore (public id, not a secret; docs/runbooks/mobile-release.md).
	extra: {
		eas: {
			projectId: "b1206657-bbbb-474f-8dce-336a773d9dd6",
		},
	},
	web: {
		bundler: "metro",
	},
};

export default config;
