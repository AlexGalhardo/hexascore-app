// The one dialog layout of the app: centred on both axes, above the sticky navbar (z-40) and the cookie banner
// (z-50), near full width with margins on mobile, scrolling inside when taller than the viewport.
// Every Base UI Dialog/AlertDialog uses these classes; add padding/gap/layout classes after them.

export const DIALOG_BACKDROP = "fixed inset-0 z-60 bg-black/50";

const WIDTH = { sm: "max-w-md", md: "max-w-xl", lg: "max-w-3xl" } as const;

export function dialogPopup(size: keyof typeof WIDTH = "sm"): string {
	return `fixed top-1/2 left-1/2 z-60 max-h-[min(90dvh,48rem)] w-[calc(100%-2rem)] ${WIDTH[size]} -translate-x-1/2 -translate-y-1/2 overflow-y-auto rounded-lg border border-border bg-bg text-fg shadow-xl`;
}
