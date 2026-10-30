export type NetworkState = "online" | "offline" | "unknown";

type Listener = (state: NetworkState) => void;

let state: NetworkState = "unknown";
const listeners = new Set<Listener>();

export function getNetworkState(): NetworkState {
	return state;
}

// The Android reachability API plugs in here later. Callers stay platform-free.
export function setNetworkState(next: NetworkState): void {
	state = next;
	for (const listener of listeners) listener(state);
}

export function subscribeNetwork(listener: Listener): () => void {
	listeners.add(listener);
	return () => {
		listeners.delete(listener);
	};
}
