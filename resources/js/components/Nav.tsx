import { Link, usePage } from "@inertiajs/react";

import { LOCALE_PREFIXES, localized, unlocalized } from "../lib/route";
import { useT } from "../lib/useT";
import type { SharedProps } from "../types";
import SuggestBox from "./search/SuggestBox";

const LOCALE_LABELS: Record<string, string> = { en: "EN", "pt-BR": "PT", es: "ES" };
const SECTIONS = [
	["games", "nav.games"],
	["movies", "nav.movies"],
	["series", "nav.series"],
	["books", "nav.books"],
] as const;

const item =
	"inline-flex min-h-11 items-center rounded-md px-2 hover:text-accent-game focus-visible:outline-2 focus-visible:outline-offset-2";

function setCookie(name: string, value: string) {
	// biome-ignore lint/suspicious/noDocumentCookie: a plain preference cookie, read server-side
	document.cookie = `${name}=${value}; path=/; max-age=31536000; samesite=lax`;
}

export function toggleTheme() {
	const dark = !document.documentElement.classList.contains("dark");
	document.documentElement.classList.toggle("dark", dark);
	const value = dark ? "dark" : "light";
	setCookie("theme", value);
	try {
		localStorage.setItem("theme", value);
	} catch {
		// Storage can be blocked: the cookie already holds the choice.
	}
}

export default function Nav() {
	const t = useT();
	const page = usePage<SharedProps>();
	const { auth, locale, unreadNotifications: unread = 0 } = page.props;
	const user = auth?.user;
	const here = unlocalized(page.url);

	return (
		<header className="border-b border-border bg-bg">
			<div className="mx-auto flex max-w-6xl flex-wrap items-center gap-x-4 px-4">
				<Link href={localized("/", locale)} className={`${item} font-bold`}>
					{page.props.appName}
				</Link>
				<nav aria-label={t("nav.sections")} className="flex flex-wrap items-center">
					{SECTIONS.map(([path, key]) => (
						<Link key={path} href={localized(`/${path}`, locale)} className={item}>
							{t(key)}
						</Link>
					))}
				</nav>
				<div className="ml-auto flex flex-wrap items-center gap-1">
					<SuggestBox />
					<ul aria-label={t("nav.language")} className="flex items-center">
						{Object.keys(LOCALE_PREFIXES).map((code) => (
							<li key={code}>
								<Link
									href={localized(here, code)}
									hrefLang={code}
									lang={code}
									aria-current={code === locale ? "true" : undefined}
									onClick={() => setCookie("locale", code)}
									className={`${item} ${code === locale ? "font-semibold underline" : ""}`}
								>
									{LOCALE_LABELS[code]}
								</Link>
							</li>
						))}
					</ul>
					<button type="button" onClick={toggleTheme} className={item}>
						{t("nav.theme")}
					</button>
					{user ? (
						<>
							{user.role === "admin" && (
								<Link href="/admin" className={item}>
									{t("nav.admin")}
								</Link>
							)}
							<Link href={localized("/me/notifications", locale)} className={item}>
								{t("nav.notifications")}
								{unread > 0 && (
									<span className="ml-1 text-xs">
										({unread > 99 ? "99+" : unread})
									</span>
								)}
							</Link>
							<Link href="/profile" className={item}>
								{user.name}
							</Link>
							<Link href="/logout" method="post" as="button" className={item}>
								{t("nav.logout")}
							</Link>
						</>
					) : (
						auth && (
							<Link href={localized("/login", locale)} className={item}>
								{t("nav.login")}
							</Link>
						)
					)}
				</div>
			</div>
		</header>
	);
}
