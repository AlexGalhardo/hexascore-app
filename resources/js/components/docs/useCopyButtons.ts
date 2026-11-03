import { type RefObject, useEffect } from "react";

const BUTTON =
	"absolute top-2 right-2 rounded-sm border border-bg/40 bg-fg px-2.5 py-1 font-sans text-xs font-semibold text-bg hover:bg-bg hover:text-fg focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-bg";

/**
 * Adds a copy button to every <pre> inside `root` (the docs body arrives as built HTML, so the buttons are DOM
 * nodes added after render). The button copies the block's text with the Clipboard API and says so for 1.5 s.
 */
export function useCopyButtons(
	root: RefObject<HTMLElement | null>,
	labels: { copy: string; copied: string },
	key: string,
): void {
	// biome-ignore lint/correctness/useExhaustiveDependencies: `key` (the page body) re-runs the effect when Inertia swaps the HTML in the same element
	useEffect(() => {
		const el = root.current;
		if (!el || !navigator.clipboard) return;
		const timers: ReturnType<typeof setTimeout>[] = [];
		const buttons = [...el.querySelectorAll("pre")].map((pre) => {
			pre.classList.add("relative");
			const button = document.createElement("button");
			button.type = "button";
			button.className = BUTTON;
			button.textContent = labels.copy;
			button.addEventListener("click", () => {
				const code = pre.querySelector("code") ?? pre;
				void navigator.clipboard.writeText(code.textContent ?? "").then(() => {
					button.textContent = labels.copied;
					timers.push(setTimeout(() => (button.textContent = labels.copy), 1500));
				});
			});
			pre.append(button);
			return button;
		});
		return () => {
			for (const t of timers) clearTimeout(t);
			for (const b of buttons) b.remove();
		};
	}, [root, labels.copy, labels.copied, key]);
}
