export interface PreferenceStore {
	get(key: string): string | null;
	set(key: string, value: string): void;
}

const memory = new Map<string, string>();

// Process-local until Expo SQLite (named in SDD 10, wired in P5.A.4.1) holds
// the preferences row. No other storage package is approved for this task.
export const memoryPreferences: PreferenceStore = {
	get(key: string): string | null {
		return memory.get(key) ?? null;
	},
	set(key: string, value: string): void {
		memory.set(key, value);
	},
};
