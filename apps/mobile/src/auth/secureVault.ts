import * as SecureStore from "expo-secure-store";
import type { KeyValueVault } from "./tokenStore";

// THIS_DEVICE_ONLY keeps the token out of device backups and restores (SDD 10 §5).
const options: SecureStore.SecureStoreOptions = {
	keychainAccessible: SecureStore.AFTER_FIRST_UNLOCK_THIS_DEVICE_ONLY,
};

export function createSecureVault(): KeyValueVault {
	return {
		get: (key) => SecureStore.getItemAsync(key, options),
		set: (key, value) => SecureStore.setItemAsync(key, value, options),
		delete: (key) => SecureStore.deleteItemAsync(key, options),
	};
}
