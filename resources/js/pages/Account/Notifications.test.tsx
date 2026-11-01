import { cleanup, render } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";

import en from "../../../lang/en.json";
import es from "../../../lang/es.json";
import ptBR from "../../../lang/pt-BR.json";
import Notifications from "./Notifications";

const catalogue = vi.hoisted(() => ({ current: {} as Record<string, unknown> }));

vi.mock("@inertiajs/react", () => ({
	Head: () => null,
	InfiniteScroll: ({ children, as: Tag = "ul" }: { children: unknown; as?: "ul" | "div" }) => (
		<Tag>{children as never}</Tag>
	),
	router: { post: vi.fn() },
}));
vi.mock("../../lib/useLocalized", () => ({ useLocalized: () => (p: string) => p }));
vi.mock("../../lib/useT", () => ({
	useT: () => (key: string, params?: Record<string, string>) => {
		const msg = key
			.split(".")
			.reduce<unknown>((o, k) => (o as Record<string, unknown>)?.[k], catalogue.current);
		return String(msg ?? key).replace(/\{(\w+)\}/g, (_, k: string) => params?.[k] ?? "");
	},
}));

afterEach(cleanup);

const row = (kind: string, targetType: string) => ({
	data: [
		{ id: "1", kind, payload: { targetType }, readAt: null, createdAt: "2026-11-01T00:00:00Z" },
	],
});

test.each([
	["en", en, "review", "Your review was hidden"],
	["pt-BR", ptBR, "review", "ocultou seu conteúdo (resenha)"],
	["es", es, "comment", "tu contenido (comentario)"],
])("content_hidden shows the translated target type in %s", (_name, cat, type, text) => {
	catalogue.current = cat as unknown as Record<string, unknown>;
	render(<Notifications notifications={row("content_hidden", type)} />);
	expect(document.body.textContent).toContain(text);
});
