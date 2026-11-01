import { QueryClient, QueryClientContext, QueryClientProvider } from "@tanstack/react-query";
import { type ComponentType, type ReactElement, useContext, useState } from "react";

// Only data the client fetches by itself lives here; Inertia page props stay Inertia's (ADR 0020).
export function makeQueryClient(): QueryClient {
	return new QueryClient({
		defaultOptions: {
			queries: { staleTime: 30_000, retry: 1, refetchOnWindowFocus: false },
		},
	});
}

let browserClient: QueryClient | undefined;

/** The browser keeps one client across every page visit; SSR gets a fresh one so no cache crosses requests. */
export function appQueryClient(ssr: boolean): QueryClient {
	if (ssr) return makeQueryClient();
	browserClient ??= makeQueryClient();
	return browserClient;
}

/**
 * Wraps a data-fetching component in the app QueryClient. Only those components wear it, so the
 * query core loads with them and stays off the first-load script of every page (ADR 0020).
 * An outer provider wins, which lets tests pass their own client.
 */
export function withQueryClient<P extends object>(
	Component: ComponentType<P>,
): (props: P) => ReactElement {
	return function WithQueryClient(props: P): ReactElement {
		const outer = useContext(QueryClientContext);
		const [client] = useState(() => outer ?? appQueryClient(typeof window === "undefined"));
		return (
			<QueryClientProvider client={client}>
				<Component {...props} />
			</QueryClientProvider>
		);
	};
}
