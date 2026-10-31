import * as Device from "expo-device";

const MAX = 64;

export function defaultDeviceName(model: string | null = Device.modelName): string {
	const name = `HexaScore · ${model?.trim() || "Android"}`;
	return name.slice(0, MAX);
}

/** The server accepts 1-64 characters (SDD 10 §6.2). */
export function isValidDeviceName(name: string): boolean {
	const length = name.trim().length;
	return length >= 1 && length <= MAX;
}
