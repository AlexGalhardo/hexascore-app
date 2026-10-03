import type { ReactNode } from "react";

const PATHS = {
	games: (
		<>
			<rect x="2" y="6" width="20" height="12" rx="4" />
			<path d="M6 12h4M8 10v4M15 13h.01M18 11h.01" />
		</>
	),
	movies: (
		<>
			<rect x="3" y="3" width="18" height="18" rx="2" />
			<path d="M7 3v18M17 3v18M3 7.5h4M3 12h18M3 16.5h4M17 7.5h4M17 16.5h4" />
		</>
	),
	series: (
		<>
			<rect x="2" y="7" width="20" height="14" rx="2" />
			<path d="m17 2-5 5-5-5" />
		</>
	),
	books: (
		<>
			<path d="M2 4h6a4 4 0 0 1 4 4v13a3 3 0 0 0-3-3H2z" />
			<path d="M22 4h-6a4 4 0 0 0-4 4v13a3 3 0 0 1 3-3h7z" />
		</>
	),
	api: <path d="m16 18 6-6-6-6M8 6l-6 6 6 6" />,
	pricing: (
		<>
			<path d="M12.6 2.6A2 2 0 0 0 11.2 2H4a2 2 0 0 0-2 2v7.2a2 2 0 0 0 .6 1.4l8.7 8.7a2.4 2.4 0 0 0 3.4 0l6.6-6.6a2.4 2.4 0 0 0 0-3.4z" />
			<circle cx="7.5" cy="7.5" r="1.5" />
		</>
	),
	language: (
		<>
			<circle cx="12" cy="12" r="10" />
			<path d="M2 12h20M12 2a15 15 0 0 1 0 20M12 2a15 15 0 0 0 0 20" />
		</>
	),
} satisfies Record<string, ReactNode>;

export type NavIconName = keyof typeof PATHS;

export function NavIcon({ name, className = "size-4" }: { name: NavIconName; className?: string }) {
	return (
		<svg
			aria-hidden="true"
			viewBox="0 0 24 24"
			fill="none"
			stroke="currentColor"
			strokeWidth="2"
			strokeLinecap="round"
			strokeLinejoin="round"
			className={`shrink-0 ${className}`}
		>
			{PATHS[name]}
		</svg>
	);
}
