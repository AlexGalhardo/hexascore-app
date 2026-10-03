// Plain module (not "use client"): server components need the real array, not a client reference.
export const DOC_PAGES = [
	{ href: "/docs", key: "quickstart" },
	{ href: "/docs/authentication", key: "authentication" },
	{ href: "/docs/errors", key: "errors" },
	{ href: "/docs/pagination", key: "pagination" },
	{ href: "/docs/rate-limits", key: "limits" },
	{ href: "/docs/changelog", key: "changelog" },
	{ href: "/docs/reference", key: "reference" },
] as const;

export type DocKey = (typeof DOC_PAGES)[number]["key"];
