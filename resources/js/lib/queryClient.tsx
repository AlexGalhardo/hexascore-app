import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactElement } from "react";

// Only data the client fetches by itself lives here; Inertia page props stay Inertia's (ADR 0020).
export function makeQueryClient(): QueryClient {
	return new QueryClient({
		defaultOptions: {
			queries: { staleTime: 30_000, retry: 1, refetchOnWindowFocus: false },
		},
	});
}

let browserClient: QueryClient | undefined;

/**
 * Inertia's `withApp` hook. The browser mounts the app once, so one client
 * outlives every page visit; SSR calls this per render and gets a fresh
 * client, so no cached data crosses requests (nothing is prefetched there).
 */
export function withQueryClient(app: ReactElement, { ssr }: { ssr: boolean }): ReactElement {
	if (ssr) return <QueryClientProvider client={makeQueryClient()}>{app}</QueryClientProvider>;
	browserClient ??= makeQueryClient();
	return <QueryClientProvider client={browserClient}>{app}</QueryClientProvider>;
}
