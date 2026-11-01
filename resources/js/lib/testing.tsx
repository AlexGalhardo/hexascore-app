import { QueryClientProvider } from "@tanstack/react-query";
import { type RenderOptions, render as rtlRender } from "@testing-library/react";
import type { ReactElement, ReactNode } from "react";

import { makeQueryClient } from "./queryClient";

export * from "@testing-library/react";

/** Testing Library's render inside a fresh app QueryClient (no retries, so errors show at once). */
export function render(ui: ReactElement, options?: RenderOptions): ReturnType<typeof rtlRender> {
	const client = makeQueryClient();
	client.setDefaultOptions({
		queries: { ...client.getDefaultOptions().queries, retry: false },
	});
	const wrapper = ({ children }: { children: ReactNode }) => (
		<QueryClientProvider client={client}>{children}</QueryClientProvider>
	);
	return rtlRender(ui, { wrapper, ...options });
}
