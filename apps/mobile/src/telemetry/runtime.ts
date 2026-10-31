import Constants from "expo-constants";
import kvStore from "expo-sqlite/kv-store";
import { Platform } from "react-native";
import { createClient } from "../../../../packages/app-client/src/api";
import { readCrashConsent, writeCrashConsent } from "./consent";
import { createCrashReporter, type ErrorUtilsLike, installGlobalHandler } from "./crash";

const appVersion = Constants.expoConfig?.version ?? "0.0.0";

// No Authorization header: crash reports are anonymous (SDD 10 §7.7).
const telemetryClient = createClient({
	baseUrl: process.env.EXPO_PUBLIC_API_BASE_URL ?? "",
	headers: { "X-App-Version": appVersion },
});

let screen = "/";
export function setCrashScreen(next: string): void {
	screen = next;
}

export const crashConsent = (): ReturnType<typeof readCrashConsent> => readCrashConsent(kvStore);
export const setCrashConsent = (granted: boolean): Promise<void> =>
	writeCrashConsent(kvStore, granted);

export const reportCrash = createCrashReporter({
	consent: crashConsent,
	send: (body) => telemetryClient.clientErrors({ body }),
	context: {
		appVersion,
		// shortcut: the static config value; EAS remote versionCode needs expo-application, add it if the owner approves.
		versionCode: Constants.expoConfig?.android?.versionCode ?? null,
		androidApi: Platform.OS === "android" ? Number(Platform.Version) : null,
	},
	screen: () => screen,
});

const errorUtils = (globalThis as { ErrorUtils?: ErrorUtilsLike }).ErrorUtils;
if (errorUtils) installGlobalHandler(reportCrash, errorUtils);
