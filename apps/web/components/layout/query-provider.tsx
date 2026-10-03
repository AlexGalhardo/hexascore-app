"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { type ReactNode, useState } from "react";

// Client-side fetches only; server-rendered data stays in Server Components. One client per browser session
// (useState), so the cache is never shared between requests on the server render.
export function QueryProvider({ children }: { children: ReactNode }): ReactNode {
	const [client] = useState(
		() =>
			new QueryClient({
				defaultOptions: {
					queries: { staleTime: 30_000, retry: 1, refetchOnWindowFocus: false },
				},
			}),
	);
	return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}
