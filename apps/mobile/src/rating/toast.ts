// In-app toast store: showToast is callable from plain modules (queue, rating), ToastHost renders it.
type Listener = () => void;

let current: { id: number; message: string } | null = null;
let counter = 0;
const listeners = new Set<Listener>();

export function showToast(message: string): void {
	counter += 1;
	current = { id: counter, message };
	for (const listener of listeners) listener();
}

export function dismissToast(): void {
	current = null;
	for (const listener of listeners) listener();
}

export function subscribeToast(listener: Listener): () => void {
	listeners.add(listener);
	return () => listeners.delete(listener);
}

export function currentToast(): { id: number; message: string } | null {
	return current;
}
