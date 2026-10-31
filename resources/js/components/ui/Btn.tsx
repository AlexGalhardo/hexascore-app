import type { ButtonHTMLAttributes } from "react";

export type BtnKind = "primary" | "secondary" | "ghost";

// One button look for every page (design v1, round 2 item 12): a 2 px ledge at rest, hover fill,
// a 2 px ink focus ring with a 2 px gap, pressed moves 1 px down onto a 1 px ledge. With
// prefers-reduced-motion the button does not move; only the ledge shrinks.
const base =
	"inline-flex min-h-11 items-center justify-center gap-2 rounded-xs border px-5 text-[15px] font-semibold leading-tight transition-[background-color,box-shadow,translate] duration-100 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fg active:translate-y-px motion-reduce:transition-none motion-reduce:active:translate-y-0 disabled:cursor-not-allowed disabled:translate-y-0 disabled:border-border disabled:bg-surface disabled:text-muted disabled:shadow-none aria-disabled:cursor-not-allowed aria-disabled:translate-y-0 aria-disabled:border-border aria-disabled:bg-surface aria-disabled:text-muted aria-disabled:shadow-none";

const kinds: Record<BtnKind, string> = {
	primary:
		"border-primary bg-primary text-on-primary shadow-ledge-primary hover:border-primary-hover hover:bg-primary-hover active:shadow-ledge-primary-pressed",
	secondary:
		"border-fg bg-bg text-fg shadow-ledge-secondary hover:bg-hover active:bg-hover active:shadow-ledge-secondary-pressed",
	ghost: "border-transparent bg-transparent text-primary hover:bg-hover active:bg-surface",
};

/** Class names for a button or a link that looks like one. */
export function btnClass(kind: BtnKind = "primary", extra = ""): string {
	return `${base} ${kinds[kind]}${extra ? ` ${extra}` : ""}`;
}

interface Props extends ButtonHTMLAttributes<HTMLButtonElement> {
	kind?: BtnKind;
}

export default function Btn({ kind = "primary", className = "", type = "button", ...rest }: Props) {
	return <button {...rest} type={type} className={btnClass(kind, className)} />;
}
