import Constants from "expo-constants";
import cacheStorage from "expo-sqlite/kv-store";
import { createAppClient } from "../api/client";
import { resolveLocale, t } from "../i18n";
import { type ListsDeps, listLaneResult } from "../lists/data";
import { subscribeAppActive } from "../platform/appState";
import { getNetworkState, subscribeNetwork } from "../platform/network";
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
import { sendEntry, startQueue } from "../queue/drain";
import { createRatingRuntime, followMe } from "../rating/runtime";
import { showToast } from "../rating/toast";
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

// The offline queue stays out of the wipe: a 401 also wipes, and the queue replays for the same
// user after re-login; another user discards it through the user binding (SDD 10 §8).
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

const ratings = createRatingRuntime({
	storage: cacheStorage,
	queryClient,
	newKey: () => appClient.newIdempotencyKey(),
	send: (entry) => sendEntry(appClient, entry),
	online: () => getNetworkState() !== "offline",
	paused: () => appClient.upgradeRequired(),
	notify: (message) => showToast(t(appLocale, message)),
	onResult: listLaneResult(queryClient, showToast, appLocale),
});
export const ratingQueue = ratings.queue;
export const listsDeps: ListsDeps = {
	read: appClient.read,
	write: appClient.write,
	newKey: () => appClient.newIdempotencyKey(),
	queue: ratings.queue,
	notify: showToast,
};
followMe(session, () => appClient.read.meShow(), ratings.setUser);
startQueue(ratings.drain, { subscribeNetwork, subscribeAppActive });
