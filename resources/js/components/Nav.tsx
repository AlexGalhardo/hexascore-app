import { Link, usePage } from "@inertiajs/react";
import { lazy, Suspense, useEffect, useRef, useState } from "react";

import { LOCALE_PREFIXES, localized, unlocalized } from "../lib/route";
import { useT } from "../lib/useT";
import type { SharedProps } from "../types";
import Dropdown from "./ui/Dropdown";

// Loaded on first open: keeps the palette and the query hooks off the first-load script.
const SearchPalette = lazy(() => import("./search/SearchPalette"));

const LOCALE_LABELS: Record<string, string> = { en: "EN", "pt-BR": "PT", es: "ES" };
// Endonyms: each language is named in itself, identical in every locale.
const LOCALE_NAMES: Record<string, string> = {
	en: "English",
	"pt-BR": "Português (Brasil)",
	es: "Español",
};
const SECTIONS = [
	["games", "nav.games", "text-accent-game"],
	["movies", "nav.movies", "text-accent-movie"],
	["series", "nav.series", "text-accent-series"],
	["books", "nav.books", "text-accent-book"],
] as const;

const focus = "focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fg";
const item = `inline-flex min-h-11 items-center rounded-xs hover:text-primary ${focus}`;
const iconButton = `inline-flex size-11 shrink-0 items-center justify-center rounded-xs hover:bg-hover ${focus}`;
const menuItem = `flex min-h-11 w-full items-center gap-2.5 px-3 text-left text-[15px] hover:bg-hover ${focus}`;

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

const ICONS = {
	search: "M11 4a7 7 0 1 0 0 14a7 7 0 1 0 0-14 M16 16 L21 21",
	moon: "M20 14.5 A8.5 8.5 0 1 1 9.5 4 A7 7 0 0 0 20 14.5 Z",
	sun: "M12 8 a4 4 0 1 0 0 8 a4 4 0 1 0 0-8 M12 2 V4.5 M12 19.5 V22 M2 12 H4.5 M19.5 12 H22 M4.9 4.9 L6.7 6.7 M17.3 17.3 L19.1 19.1 M4.9 19.1 L6.7 17.3 M17.3 6.7 L19.1 4.9",
	bell: "M6 16 V11 a6 6 0 0 1 12 0 V16 L19.5 18 H4.5 Z M10 20.5 a2 2 0 0 0 4 0",
	menu: "M4 7 H20 M4 12 H20 M4 17 H20",
	down: "M6 9 L12 15 L18 9",
	check: "M5 12 L10 17 L19 7",
} as const;

// Icons from the v1 set; the label lives in aria-label or next to the icon.
export function Icon({
	name,
	className = "size-5",
}: {
	name: keyof typeof ICONS;
	className?: string;
}) {
	return (
		<svg aria-hidden="true" viewBox="0 0 24 24" className={className} fill="none">
			<path d={ICONS[name]} stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" />
		</svg>
	);
}

/** True once the page has scrolled, for the shadow under the fixed navbar. */
function useScrolled(): boolean {
	const [scrolled, setScrolled] = useState(false);
	useEffect(() => {
		const onScroll = () => setScrolled(window.scrollY > 0);
		onScroll();
		window.addEventListener("scroll", onScroll, { passive: true });
		return () => window.removeEventListener("scroll", onScroll);
	}, []);
	return scrolled;
}

/** "⌘K" on Apple platforms, "Ctrl K" elsewhere; the server always renders "Ctrl K". */
function useShortcutHint(): string {
	const [hint, setHint] = useState("Ctrl K");
	useEffect(() => {
		if (/Mac|iPhone|iPad/.test(navigator.platform)) setHint("⌘K");
	}, []);
	return hint;
}

export default function Nav() {
	const t = useT();
	const page = usePage<SharedProps>();
	const { auth, locale, unreadNotifications: unread = 0 } = page.props;
	const user = auth?.user;
	const feedHref = user && page.props.flags?.social ? localized("/feed", locale) : null;
	const here = unlocalized(page.url);
	const unreadLabel = unread > 99 ? "99+" : String(unread);
	const scrolled = useScrolled();
	const hint = useShortcutHint();
	const searchHref = localized("/search", locale);
	const [paletteOpen, setPaletteOpen] = useState(false);
	const [paletteMounted, setPaletteMounted] = useState(false);
	if (paletteOpen && !paletteMounted) setPaletteMounted(true);
	const searchTrigger = useRef<HTMLAnchorElement>(null);

	useEffect(() => {
		const onKey = (e: KeyboardEvent) => {
			if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === "k") {
				e.preventDefault();
				setPaletteOpen(true);
			}
		};
		document.addEventListener("keydown", onKey);
		return () => document.removeEventListener("keydown", onKey);
	}, []);

	const language = (
		<Dropdown
			label={t("nav.language")}
			triggerClassName={`inline-flex min-h-11 items-center gap-1 rounded-xs px-2 text-sm font-bold hover:bg-hover aria-expanded:bg-hover ${focus}`}
			trigger={
				<>
					{LOCALE_LABELS[locale] ?? "EN"}
					<Icon name="down" className="size-3.5" />
				</>
			}
		>
			<ul aria-label={t("nav.language")}>
				{Object.keys(LOCALE_PREFIXES).map((code) => (
					<li key={code}>
						<Link
							href={localized(here, code)}
							hrefLang={code}
							lang={code}
							aria-current={code === locale ? "true" : undefined}
							onClick={() => setCookie("locale", code)}
							className={`${menuItem} ${code === locale ? "bg-hover font-bold" : ""}`}
						>
							<span className="size-4 text-primary">
								{code === locale && <Icon name="check" className="size-4" />}
							</span>
							<span className="flex-1">{LOCALE_NAMES[code]}</span>
							<span className="text-xs text-muted">{LOCALE_LABELS[code]}</span>
						</Link>
					</li>
				))}
			</ul>
		</Dropdown>
	);

	return (
		<header
			data-scrolled={scrolled || undefined}
			className="sticky top-0 z-50 border-b border-fg bg-bg transition-shadow motion-reduce:transition-none data-scrolled:shadow-nav dark:data-scrolled:shadow-nav-dark"
		>
			<div className="shell flex h-14 items-center gap-1 lg:h-[72px] lg:gap-7">
				<Link
					href={localized("/", locale)}
					className={`${item} shrink-0 font-display text-2xl font-bold hover:text-fg lg:text-[28px]`}
				>
					{page.props.appName}
				</Link>
				<nav
					aria-label={t("nav.sections")}
					className="hidden items-center gap-7 whitespace-nowrap lg:flex"
				>
					{SECTIONS.map(([path, key]) => (
						<Link
							key={path}
							href={localized(`/${path}`, locale)}
							className={`${item} text-[15px] font-medium`}
						>
							{t(key)}
						</Link>
					))}
					{feedHref && (
						<Link
							href={feedHref}
							aria-current={here === "/feed" ? "page" : undefined}
							className={`${item} border-l border-border pl-7 text-[15px] font-medium aria-[current=page]:font-bold`}
						>
							{t("social.feed.title")}
						</Link>
					)}
				</nav>
				<div className="ml-auto flex items-center gap-1 lg:gap-7">
					{/* A plain link to /search; the command palette takes over the click and Ctrl/⌘ K. */}
					<a
						ref={searchTrigger}
						href={searchHref}
						data-search-trigger=""
						aria-haspopup="dialog"
						onClick={(e) => {
							// A modified click still opens /search in a new tab.
							if (e.ctrlKey || e.metaKey || e.shiftKey || e.button !== 0) return;
							e.preventDefault();
							setPaletteOpen(true);
						}}
						aria-label={t("nav.openSearch")}
						aria-keyshortcuts="Control+K Meta+K"
						className={`${iconButton} lg:h-11 lg:w-56 lg:justify-start lg:gap-2 lg:border lg:border-field lg:bg-field-bg lg:px-3 lg:text-muted xl:w-[280px]`}
					>
						<Icon name="search" className="size-5 lg:size-[18px]" />
						<span className="hidden text-[15px] lg:inline">{t("nav.searchShort")}</span>
						<kbd className="ml-auto hidden h-[22px] items-center rounded-sm border border-border bg-surface px-1.5 font-sans text-xs font-semibold lg:inline-flex">
							{hint}
						</kbd>
					</a>
					<div className="flex items-center gap-1">
						{language}
						<button
							type="button"
							onClick={toggleTheme}
							className={iconButton}
							aria-label={t("nav.theme")}
							title={t("nav.theme")}
						>
							<Icon name="moon" className="size-5 dark:hidden" />
							<Icon name="sun" className="hidden size-5 dark:block" />
						</button>
					</div>
					{user ? (
						<div className="flex items-center gap-1">
							<Link
								href={localized("/me/notifications", locale)}
								className={`${iconButton} relative`}
								aria-label={`${t("nav.notifications")}${unread > 0 ? ` (${unreadLabel})` : ""}`}
								title={t("nav.notifications")}
							>
								<Icon name="bell" />
								{unread > 0 && (
									<span className="absolute top-1 right-0 min-w-5 rounded-full bg-primary px-1 text-center text-[10px] font-bold leading-[18px] text-on-primary tabular-nums">
										{unreadLabel}
									</span>
								)}
							</Link>
							<Dropdown
								label={t("nav.account")}
								triggerClassName={`inline-flex min-h-11 items-center gap-2 rounded-xs pl-1 pr-1 hover:bg-hover ${focus}`}
								panelClassName="w-64"
								trigger={
									<>
										<span
											aria-hidden="true"
											className="inline-flex size-9 items-center justify-center rounded-full bg-fg font-bold text-bg"
										>
											{(user.name || "?").slice(0, 1).toUpperCase()}
										</span>
										<span className="hidden max-w-32 truncate font-semibold xl:inline">
											{user.nickname || user.name}
										</span>
									</>
								}
							>
								<p className="truncate px-3 py-2 text-sm font-semibold">
									{user.name}
								</p>
								{feedHref && (
									<Link href={feedHref} className={`${menuItem} lg:hidden`}>
										{t("social.feed.title")}
									</Link>
								)}
								<Link href={localized("/library", locale)} className={menuItem}>
									{t("nav.library")}
								</Link>
								<Link href={localized("/lists", locale)} className={menuItem}>
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
								<Link href="/logout" method="post" as="button" className={menuItem}>
									{t("nav.logout")}
								</Link>
							</Dropdown>
						</div>
					) : (
						auth && (
							<div className="flex items-center gap-1 lg:gap-4">
								<Link
									href={localized("/login", locale)}
									className={`${item} whitespace-nowrap px-1 text-[15px] font-semibold text-primary hover:underline`}
								>
									{t("nav.login")}
								</Link>
								<Dropdown
									label={t("nav.more")}
									triggerClassName={`inline-flex min-h-11 min-w-11 items-center justify-center gap-1.5 rounded-xs hover:bg-hover lg:px-2 ${focus}`}
									trigger={
										<>
											<Icon name="menu" />
											<span className="hidden font-semibold xl:inline">
												{t("nav.more")}
											</span>
										</>
									}
								>
									<Link href={localized("/signup", locale)} className={menuItem}>
										{t("auth.signup.title")}
									</Link>
									<Link href={localized("/pricing", locale)} className={menuItem}>
										{t("nav.pricing")}
									</Link>
									<Link href="/docs" className={menuItem}>
										{t("nav.apiDocs")}
									</Link>
								</Dropdown>
							</div>
						)
					)}
				</div>
			</div>
			<nav
				aria-label={t("nav.sections")}
				className="shell flex justify-between border-t border-border lg:hidden"
			>
				{SECTIONS.map(([path, key, accent]) => (
					<Link
						key={path}
						href={localized(`/${path}`, locale)}
						className={`${item} text-[15px] font-semibold ${accent}`}
					>
						{t(key)}
					</Link>
				))}
			</nav>
			{paletteMounted && (
				<Suspense fallback={null}>
					<SearchPalette
						open={paletteOpen}
						onOpenChange={setPaletteOpen}
						returnFocus={searchTrigger}
					/>
				</Suspense>
			)}
		</header>
	);
}
