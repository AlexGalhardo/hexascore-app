// Jest runs on Node. Expo SDK 58 does not ship JSC; the Android binary defines
// HermesInternal. Mirror that global so the engine assertion can run in CI.
const runtime = globalThis as { HermesInternal?: object };
if (runtime.HermesInternal === undefined) runtime.HermesInternal = {};

// react-native exposes its components through lazy getters, so the first render
// in a test file would transform and load them inside the 5 s test budget
// (6.5 s on a cold Jest cache). Touch them here, which no test timer covers.
const { FlatList, Pressable, ScrollView, Text, View } =
	require("react-native") as typeof import("react-native");
void [FlatList, Pressable, ScrollView, Text, View];

// Jest has no native SQLite; screens that import the runtime get an in-memory kv-store.
jest.mock("expo-sqlite/kv-store", () => {
	const data = new Map<string, string>();
	return {
		__esModule: true,
		default: {
			getItem: async (key: string) => data.get(key) ?? null,
			getItemSync: (key: string) => data.get(key) ?? null,
			setItem: async (key: string, value: string) => {
				data.set(key, value);
			},
			removeItem: async (key: string) => {
				data.delete(key);
			},
		},
	};
});
