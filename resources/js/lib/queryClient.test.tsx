import { type QueryClient, QueryClientProvider, useQueryClient } from "@tanstack/react-query";
import { render } from "@testing-library/react";
import { expect, test } from "vitest";

import { appQueryClient, makeQueryClient, withQueryClient } from "./queryClient";

test("defaults: 30 s stale time, one retry, no refetch on focus", () => {
	expect(makeQueryClient().getDefaultOptions().queries).toMatchObject({
		staleTime: 30_000,
		retry: 1,
		refetchOnWindowFocus: false,
	});
});

test("the browser keeps one client across visits; SSR gets a fresh one per render", () => {
	expect(appQueryClient(false)).toBe(appQueryClient(false));
	const a = appQueryClient(true);
	expect(a).not.toBe(appQueryClient(true));
	expect(a).not.toBe(appQueryClient(false));
});

test("withQueryClient gives the app client, and an outer provider wins", () => {
	const seen: QueryClient[] = [];
	const Probe = withQueryClient(() => {
		seen.push(useQueryClient());
		return null;
	});
	render(<Probe />);
	render(<Probe />);
	expect(seen[0]).toBe(appQueryClient(false));
	expect(seen[1]).toBe(seen[0]);

	const outer = makeQueryClient();
	render(
		<QueryClientProvider client={outer}>
			<Probe />
		</QueryClientProvider>,
	);
	expect(seen[2]).toBe(outer);
});
