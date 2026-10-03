"use client";

import { setPreferenceCookie, THEME_COOKIE, type Theme } from "@/lib/preferences";

export function ThemeToggle({ label }: { label: string }) {
	function toggle(): void {
		const next: Theme = document.documentElement.dataset.theme === "dark" ? "light" : "dark";
		document.documentElement.dataset.theme = next;
		setPreferenceCookie(THEME_COOKIE, next);
	}

	return (
		<button
			type="button"
			onClick={toggle}
			aria-label={label}
			title={label}
			className="grid size-11 place-items-center rounded-md border border-border hover:bg-surface"
		>
			<svg
				aria-hidden="true"
				viewBox="0 0 24 24"
				className="size-5"
				fill="none"
				stroke="currentColor"
				strokeWidth="2"
			>
				<path d="M21 12.8A9 9 0 1 1 11.2 3a7 7 0 0 0 9.8 9.8z" />
			</svg>
		</button>
	);
}
