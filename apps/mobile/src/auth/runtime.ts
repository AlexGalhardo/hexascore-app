import Constants from "expo-constants";
import cacheStorage from "expo-sqlite/kv-store";
import { createAppClient } from "../api/client";
import { resolveLocale } from "../i18n";
import type { RootDeps } from "../providers/root";
import { createQueryClient, wireQueryManagers } from "../query/client";
import {
	cacheBuster,
	createSqlitePersister,
	dehydrateOptions,
	MAX_AGE_MS,
	readFormulaVersion,
	trackFormulaVersion,
} from "../query/persister";
import { startForegroundRotation } from "./rotate";
import { createSecureVault } from "./secureVault";
import { createSession } from "./session";

export const apiBaseUrl = process.env.EXPO_PUBLIC_API_BASE_URL ?? "";

// Hermes ships Intl, so the device tag needs no localization package.
export const appLocale = resolveLocale([Intl.DateTimeFormat().resolvedOptions().locale]);

const vault = createSecureVault();
const appVersion = Constants.expoConfig?.version ?? "0.0.0";

export const queryClient = createQueryClient();
export const persistOptions = {
	persister: createSqlitePersister(cacheStorage),
	maxAge: MAX_AGE_MS,
	buster: cacheBuster(appVersion, readFormulaVersion(cacheStorage)),
	dehydrateOptions,
};
wireQueryManagers();
trackFormulaVersion(queryClient, cacheStorage);

// wipeUserData grows with the offline queue (P5.A.4.2).
export const session = createSession({
	vault,
	wipeUserData: async () => {
		queryClient.clear();
		await persistOptions.persister.removeClient();
	},
});

let notifyUpgrade = (): void => {};

export const appClient = createAppClient({
	baseUrl: apiBaseUrl,
	version: appVersion,
	locale: appLocale,
	token: () => session.token(),
	onUpgradeRequired: () => notifyUpgrade(),
});

export const rootDeps: RootDeps = {
	session,
	startRotation: () =>
		startForegroundRotation({
			vault,
			session,
			client: {
				tokensRotate: () =>
					appClient.write(appClient.newIdempotencyKey(), (client) => client.tokensRotate()),
			},
		}),
	upgradeRequired: () => appClient.upgradeRequired(),
	onUpgradeRequired(listener) {
		notifyUpgrade = listener;
		return () => {
			notifyUpgrade = () => {};
		};
	},
};
