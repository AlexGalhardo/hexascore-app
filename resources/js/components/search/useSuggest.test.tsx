import { afterEach, expect, test, vi } from "vitest";

import { cleanup, render, screen, waitFor } from "../../lib/testing";
import { useSuggest } from "./useSuggest";

function Probe({ text, type }: { text: string; type?: string }) {
	const s = useSuggest(text, "en", type);
	return (
		<output>
			{JSON.stringify({
				titles: s.items.map((i) => i.title),
				pending: s.pending,
				answered: s.answered,
			})}
		</output>
	);
}

const state = () => JSON.parse(screen.getByRole("status").textContent ?? "{}");

const answer = (title: string) => ({
	ok: true,
	json: async () => ({
		data: [{ id: title, type: "game", kind: "game", title, year: null, path: `/g/${title}` }],
	}),
});

afterEach(() => {
	cleanup();
	vi.unstubAllGlobals();
});

test("debounces typing into one keyed request and answers it", async () => {
	const fetchMock = vi.fn().mockResolvedValue(answer("Zelda"));
	vi.stubGlobal("fetch", fetchMock);
	const { rerender } = render(<Probe text="ze" />);
	rerender(<Probe text="zel" />);
	expect(state().pending).toBe(true);
	rerender(<Probe text="zeld" />);
	await waitFor(() => expect(state().answered).toBe("zeld"));
	expect(fetchMock).toHaveBeenCalledTimes(1);
	expect(String(fetchMock.mock.calls[0][0])).toBe("/search/suggest?q=zeld");
	expect(state()).toEqual({ titles: ["Zelda"], pending: false, answered: "zeld" });
});

test("aborts the in-flight request when the text changes and drops its answer", async () => {
	const signals: AbortSignal[] = [];
	vi.stubGlobal(
		"fetch",
		vi.fn((url: string, init: RequestInit) => {
			signals.push(init.signal as AbortSignal);
			return url.includes("q=ab") ? new Promise(() => {}) : Promise.resolve(answer("Second"));
		}),
	);
	const { rerender } = render(<Probe text="ab" />);
	await waitFor(() => expect(signals).toHaveLength(1));
	rerender(<Probe text="xy" />);
	await waitFor(() => expect(state().answered).toBe("xy"));
	expect(signals[0].aborted).toBe(true);
	expect(state().titles).toEqual(["Second"]);
});

test("serves a repeated query from the cache", async () => {
	const fetchMock = vi.fn((url: string) => Promise.resolve(answer(url)));
	vi.stubGlobal("fetch", fetchMock);
	const { rerender } = render(<Probe text="halo" />);
	await waitFor(() => expect(state().answered).toBe("halo"));
	rerender(<Probe text="doom" />);
	await waitFor(() => expect(state().answered).toBe("doom"));
	rerender(<Probe text="halo" />);
	await waitFor(() => expect(state().answered).toBe("halo"));
	expect(fetchMock).toHaveBeenCalledTimes(2);
});

test("clears under two characters and answers errors with no items", async () => {
	vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: false, status: 500 }));
	const { rerender } = render(<Probe text="er" type="book" />);
	await waitFor(() => expect(state().answered).toBe("er"));
	expect(state()).toEqual({ titles: [], pending: false, answered: "er" });
	expect(String(vi.mocked(fetch).mock.calls[0][0])).toContain("&type=book");
	rerender(<Probe text="e" />);
	expect(state()).toEqual({ titles: [], pending: false, answered: "" });
});
