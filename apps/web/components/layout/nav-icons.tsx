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
	user: (
		<>
			<circle cx="12" cy="8" r="4" />
			<path d="M4 21a8 8 0 0 1 16 0" />
		</>
	),
	publicProfile: (
		<>
			<rect x="3" y="4" width="18" height="16" rx="2" />
			<circle cx="9" cy="11" r="2.5" />
			<path d="M5.5 17a3.5 3.5 0 0 1 7 0M15 10h3M15 14h3" />
		</>
	),
	feed: (
		<>
			<path d="M4 11a9 9 0 0 1 9 9M4 4a16 16 0 0 1 16 16" />
			<circle cx="5" cy="19" r="1" />
		</>
	),
	key: (
		<>
			<circle cx="7.5" cy="15.5" r="4.5" />
			<path d="m10.7 12.3 9.3-9.3M17 6l3 3M14 9l2 2" />
		</>
	),
	admin: <path d="M12 3 4 6v6c0 5 3.5 8 8 9 4.5-1 8-4 8-9V6z" />,
	logout: <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4M16 17l5-5-5-5M21 12H9" />,
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
