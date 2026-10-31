import Constants from "expo-constants";
import { createAppClient } from "../api/client";
import { resolveLocale } from "../i18n";
import type { RootDeps } from "../providers/root";
import { startForegroundRotation } from "./rotate";
import { createSecureVault } from "./secureVault";
import { createSession } from "./session";

export const apiBaseUrl = process.env.EXPO_PUBLIC_API_BASE_URL ?? "";

// Hermes ships Intl, so the device tag needs no localization package.
export const appLocale = resolveLocale([Intl.DateTimeFormat().resolvedOptions().locale]);

const vault = createSecureVault();

// wipeUserData grows with the query cache (P5.A.4.1) and the offline queue (P5.A.4.2).
export const session = createSession({ vault, wipeUserData: async () => {} });

let notifyUpgrade = (): void => {};

export const appClient = createAppClient({
	baseUrl: apiBaseUrl,
	version: Constants.expoConfig?.version ?? "0.0.0",
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
