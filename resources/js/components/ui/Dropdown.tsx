import { type ReactNode, useEffect, useId, useRef, useState } from "react";

interface Props {
	/** Content of the trigger button (icon, short label). */
	trigger: ReactNode;
	/** Accessible name of the trigger. */
	label: string;
	triggerClassName: string;
	panelClassName?: string;
	children: ReactNode;
}

// A disclosure menu of links: the button toggles the panel; Escape, a click outside or following a
// link closes it, and Escape returns focus to the button. Plain DOM positioning (absolute, below the
// button) keeps it free of inline styles under the CSP.
export default function Dropdown({
	trigger,
	label,
	triggerClassName,
	panelClassName = "",
	children,
}: Props) {
	const [open, setOpen] = useState(false);
	const root = useRef<HTMLDivElement>(null);
	const button = useRef<HTMLButtonElement>(null);
	const id = useId();

	useEffect(() => {
		if (!open) return;
		const onKey = (e: KeyboardEvent) => {
			if (e.key !== "Escape") return;
			setOpen(false);
			button.current?.focus();
		};
		const onPointer = (e: PointerEvent) => {
			if (!root.current?.contains(e.target as Node)) setOpen(false);
		};
		document.addEventListener("keydown", onKey);
		document.addEventListener("pointerdown", onPointer);
		return () => {
			document.removeEventListener("keydown", onKey);
			document.removeEventListener("pointerdown", onPointer);
		};
	}, [open]);

	return (
		<div ref={root} className="relative">
			<button
				ref={button}
				type="button"
				aria-label={label}
				aria-expanded={open}
				aria-controls={id}
				onClick={() => setOpen((o) => !o)}
				className={triggerClassName}
			>
				{trigger}
			</button>
			{/* biome-ignore lint/a11y/noStaticElementInteractions: closes the menu after a link inside is followed */}
			{/* biome-ignore lint/a11y/useKeyWithClickEvents: Enter on a link or button fires click, so keyboard use is covered */}
			<div
				id={id}
				hidden={!open}
				onClick={(e) => {
					if ((e.target as HTMLElement).closest("a,button")) setOpen(false);
				}}
				className={`absolute right-0 z-50 mt-1 min-w-52 border border-fg bg-bg py-1 shadow-[0_4px_0_var(--color-border)] ${panelClassName}`}
			>
				{children}
			</div>
		</div>
	);
}
