import Constants from "expo-constants";
import { createAppClient } from "../api/client";
import { createSecureVault } from "./secureVault";
import { createSession } from "./session";

export const apiBaseUrl = process.env.EXPO_PUBLIC_API_BASE_URL ?? "";

// wipeUserData grows with the query cache (P5.A.4.1) and the offline queue (P5.A.4.2).
export const session = createSession({ vault: createSecureVault(), wipeUserData: async () => {} });

export const appClient = createAppClient({
	baseUrl: apiBaseUrl,
	version: Constants.expoConfig?.version ?? "0.0.0",
	locale: "en",
	token: () => session.token(),
});
