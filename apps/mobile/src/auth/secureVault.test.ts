import * as SecureStore from "expo-secure-store";
import { createSecureVault } from "./secureVault";
import { createTokenStore } from "./tokenStore";

jest.mock("expo-secure-store", () => {
	const data = new Map<string, string>();
	return {
		AFTER_FIRST_UNLOCK_THIS_DEVICE_ONLY: 6,
		getItemAsync: jest.fn(async (key: string) => data.get(key) ?? null),
		setItemAsync: jest.fn(async (key: string, value: string) => {
			data.set(key, value);
		}),
		deleteItemAsync: jest.fn(async (key: string) => {
			data.delete(key);
		}),
	};
});

describe("secureVault", () => {
	it("round-trips a token through SecureStore with the device-only option", async () => {
		const store = createTokenStore(createSecureVault());
		const token = { token: "t", tokenId: "id", expiresAt: "2027-01-01T00:00:00Z" };
		await store.save(token);
		expect(await store.load()).toEqual(token);
		expect(SecureStore.setItemAsync).toHaveBeenCalledWith("hexascore.token", expect.any(String), {
			keychainAccessible: 6,
		});
	});

	it("clear removes the entry", async () => {
		const store = createTokenStore(createSecureVault());
		await store.save({ token: "t", tokenId: "id", expiresAt: "x" });
		await store.clear();
		expect(await store.load()).toBeNull();
		expect(SecureStore.deleteItemAsync).toHaveBeenCalled();
	});
});
