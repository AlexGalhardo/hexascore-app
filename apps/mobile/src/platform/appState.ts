import { AppState } from "react-native";

export function subscribeAppActive(listener: (active: boolean) => void): () => void {
	const subscription = AppState.addEventListener("change", (state) => listener(state === "active"));
	return () => subscription.remove();
}
