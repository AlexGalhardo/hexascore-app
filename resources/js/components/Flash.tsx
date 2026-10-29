import { useEffect, useRef, useState } from "react";

import { useT } from "../lib/useT";

const DISMISS_MS = 6000;

// Flash shows a flash message as a toast: role=status, or role=alert for
// errors; it closes after 6 s, paused while hovered or focused (SDD 09 §7.9).
export default function Flash({ message, kind }: { message: string; kind: "status" | "error" }) {
	const t = useT();
	const [open, setOpen] = useState(true);
	const [paused, setPaused] = useState(false);
	const timer = useRef<ReturnType<typeof setTimeout>>(undefined);

	// biome-ignore lint/correctness/useExhaustiveDependencies: a new message restarts the countdown
	useEffect(() => {
		setOpen(true);
	}, [message]);

	useEffect(() => {
		if (!open || paused) return;
		timer.current = setTimeout(() => setOpen(false), DISMISS_MS);
		return () => clearTimeout(timer.current);
	}, [open, paused]);

	if (!open) return null;

	const tone = kind === "error" ? "border-danger text-danger" : "border-border text-fg";

	return (
		// biome-ignore lint/a11y/noStaticElementInteractions: hover and focus only pause the countdown, nothing is activated here
		<div
			role={kind === "error" ? "alert" : "status"}
			onMouseEnter={() => setPaused(true)}
			onMouseLeave={() => setPaused(false)}
			onFocus={() => setPaused(true)}
			onBlur={() => setPaused(false)}
			className={`fixed right-4 bottom-4 z-[70] flex max-w-sm items-start gap-3 rounded-md border bg-surface px-4 py-3 shadow-lg ${tone}`}
		>
			<p>{message}</p>
			<button
				type="button"
				onClick={() => setOpen(false)}
				aria-label={t("nav.dismiss")}
				className="-my-2 -mr-3 inline-flex min-h-11 min-w-11 items-center justify-center"
			>
				×
			</button>
		</div>
	);
}
