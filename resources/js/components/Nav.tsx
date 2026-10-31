import { Link, usePage } from "@inertiajs/react";

import { LOCALE_PREFIXES, localized, unlocalized } from "../lib/route";
import { useT } from "../lib/useT";
import type { SharedProps } from "../types";
import SuggestBox from "./search/SuggestBox";

const LOCALE_LABELS: Record<string, string> = { en: "EN", "pt-BR": "PT", es: "ES" };
const SECTIONS = [
	["games", "nav.games", "text-accent-game"],
	["movies", "nav.movies", "text-accent-movie"],
	["series", "nav.series", "text-accent-series"],
	["books", "nav.books", "text-accent-book"],
] as const;

const item =
	"inline-flex min-h-11 items-center rounded-md px-2 hover:text-accent-game focus-visible:outline-2 focus-visible:outline-offset-2";
const iconButton =
	"inline-flex size-11 items-center justify-center rounded-md hover:bg-surface focus-visible:outline-2 focus-visible:outline-offset-2";
const menuItem =
	"flex min-h-11 w-full items-center gap-2 rounded-md px-3 text-left hover:bg-surface focus-visible:outline-2";
const menuPanel =
	"absolute right-0 z-40 mt-1 w-64 rounded-lg border border-border bg-bg p-2 shadow-lg";

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

// Inline icons from the SDD 09 §7.5 set; the label lives in aria-label or next to the icon.
function Icon({ name }: { name: "theme" | "bell" | "user" | "menu" }) {
	const paths: Record<typeof name, string> = {
		theme: "M12 3a9 9 0 1 0 0 18V3z",
		bell: "M6 16V11a6 6 0 1 1 12 0v5l2 2H4l2-2zm4 4h4",
		user: "M12 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8zm-7 8a7 7 0 0 1 14 0",
		menu: "M4 7h16M4 12h16M4 17h16",
	};
	return (
		<svg aria-hidden="true" viewBox="0 0 24 24" className="size-5" fill="none">
			<path d={paths[name]} stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
		</svg>
	);
}

export default function Nav() {
	const t = useT();
	const page = usePage<SharedProps>();
	const { auth, locale, unreadNotifications: unread = 0 } = page.props;
	const user = auth?.user;
	const here = unlocalized(page.url);
	const unreadLabel = unread > 99 ? "99+" : String(unread);

	const localeLinks = (
		<ul aria-label={t("nav.language")} className="flex items-center gap-1">
			{Object.keys(LOCALE_PREFIXES).map((code) => (
				<li key={code}>
					<Link
						href={localized(here, code)}
						hrefLang={code}
						lang={code}
						aria-current={code === locale ? "true" : undefined}
						onClick={() => setCookie("locale", code)}
						className={`${item} text-sm ${code === locale ? "font-semibold underline" : "text-muted"}`}
					>
						{LOCALE_LABELS[code]}
					</Link>
				</li>
			))}
		</ul>
	);

	return (
		<header className="border-b border-border bg-bg">
			<div className="mx-auto max-w-6xl px-4">
				<div className="flex min-h-14 items-center gap-2 sm:gap-3">
					<Link
						href={localized("/", locale)}
						className={`${item} px-0 text-lg font-bold`}
					>
						{page.props.appName}
					</Link>
					<nav aria-label={t("nav.sections")} className="hidden items-center md:flex">
						{SECTIONS.map(([path, key]) => (
							<Link key={path} href={localized(`/${path}`, locale)} className={item}>
								{t(key)}
							</Link>
						))}
					</nav>
					<div className="ml-auto flex min-w-0 flex-wrap items-center justify-end gap-1">
						{/* On phones the search box takes its own full row below the icons. */}
						<div className="order-last basis-full pb-2 sm:order-none sm:basis-auto sm:pb-0">
							<SuggestBox />
						</div>
						<button
							type="button"
							onClick={toggleTheme}
							className={iconButton}
							aria-label={t("nav.theme")}
							title={t("nav.theme")}
						>
							<Icon name="theme" />
						</button>
						{user ? (
							<>
								<Link
									href={localized("/me/notifications", locale)}
									className={`${iconButton} relative`}
									aria-label={`${t("nav.notifications")}${unread > 0 ? ` (${unreadLabel})` : ""}`}
									title={t("nav.notifications")}
								>
									<Icon name="bell" />
									{unread > 0 && (
										<span className="absolute top-1 right-1 min-w-4 rounded-full bg-accent-movie px-1 text-center text-[10px] font-semibold leading-4 text-white">
											{unreadLabel}
										</span>
									)}
								</Link>
								<details className="group relative">
									<summary
										className={`${iconButton} list-none [&::-webkit-details-marker]:hidden`}
										aria-label={t("nav.account")}
										title={user.name}
									>
										<Icon name="user" />
									</summary>
									<div className={menuPanel}>
										<p className="truncate px-3 py-2 text-sm font-semibold">
											{user.name}
										</p>
										<Link
											href={localized("/library", locale)}
											className={menuItem}
										>
											{t("nav.library")}
										</Link>
										<Link
											href={localized("/lists", locale)}
											className={menuItem}
										>
											{t("nav.lists")}
										</Link>
										{user.nickname && (
											<Link
												href={localized(
													`/u/${user.nickname.toLowerCase()}`,
													locale,
												)}
												className={menuItem}
											>
												{t("nav.publicProfile")}
											</Link>
										)}
										<Link href="/profile" className={menuItem}>
											{t("nav.profile")}
										</Link>
										<Link href="/profile/api" className={menuItem}>
											{t("nav.api")}
										</Link>
										{user.role === "admin" && (
											<Link href="/admin" className={menuItem}>
												{t("nav.admin")}
											</Link>
										)}
										<div className="my-1 border-t border-border" />
										<div className="px-1">{localeLinks}</div>
										<Link
											href="/logout"
											method="post"
											as="button"
											className={menuItem}
										>
											{t("nav.logout")}
										</Link>
									</div>
								</details>
							</>
						) : (
							auth && (
								<>
									<Link
										href={localized("/login", locale)}
										className={`${item} font-medium`}
									>
										{t("nav.login")}
									</Link>
									<details className="group relative">
										<summary
											className={`${iconButton} list-none [&::-webkit-details-marker]:hidden`}
											aria-label={t("nav.more")}
										>
											<Icon name="menu" />
										</summary>
										<div className={menuPanel}>
											<div className="px-1">{localeLinks}</div>
											<Link
												href={localized("/signup", locale)}
												className={menuItem}
											>
												{t("auth.signup.title")}
											</Link>
											<Link
												href={localized("/pricing", locale)}
												className={menuItem}
											>
												{t("nav.pricing")}
											</Link>
											<Link href="/docs" className={menuItem}>
												{t("nav.apiDocs")}
											</Link>
										</div>
									</details>
								</>
							)
						)}
					</div>
				</div>
				<nav
					aria-label={t("nav.sections")}
					className="-mx-4 flex overflow-x-auto px-3 pb-1 md:hidden"
				>
					{SECTIONS.map(([path, key, accent]) => (
						<Link
							key={path}
							href={localized(`/${path}`, locale)}
							className={`${item} shrink-0 font-medium ${accent}`}
						>
							{t(key)}
						</Link>
					))}
				</nav>
			</div>
		</header>
	);
}
