import type { KeyValueStorage } from "../query/persister";

/** LGPD opt-in (SDD 10 §7.7): asked at first launch, default off, revocable in Settings. */
export type CrashConsent = "unset" | "granted" | "denied";
export type ConsentStorage = Pick<KeyValueStorage, "getItemSync" | "setItem">;

const KEY = "hexascore.crashConsent";

export function readCrashConsent(storage: ConsentStorage): CrashConsent {
	const value = storage.getItemSync(KEY);
	return value === "granted" || value === "denied" ? value : "unset";
}

export function writeCrashConsent(storage: ConsentStorage, granted: boolean): Promise<void> {
	return storage.setItem(KEY, granted ? "granted" : "denied");
}
