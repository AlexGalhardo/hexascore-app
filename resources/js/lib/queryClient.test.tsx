import type { QueryClient } from "@tanstack/react-query";
import type { ReactElement } from "react";
import { expect, test } from "vitest";

import { makeQueryClient, withQueryClient } from "./queryClient";

const clientOf = (el: ReactElement): QueryClient => (el.props as { client: QueryClient }).client;

test("defaults: 30 s stale time, one retry, no refetch on focus", () => {
	expect(makeQueryClient().getDefaultOptions().queries).toMatchObject({
		staleTime: 30_000,
		retry: 1,
		refetchOnWindowFocus: false,
	});
});

test("the browser keeps one client across visits; SSR gets a fresh one per render", () => {
	const app = <main />;
	expect(clientOf(withQueryClient(app, { ssr: false }))).toBe(
		clientOf(withQueryClient(app, { ssr: false })),
	);
	const a = clientOf(withQueryClient(app, { ssr: true }));
	const b = clientOf(withQueryClient(app, { ssr: true }));
	expect(a).not.toBe(b);
	expect(a).not.toBe(clientOf(withQueryClient(app, { ssr: false })));
});
