import { Link, usePage } from "@inertiajs/react";
import { lazy, Suspense, useEffect, useRef, useState } from "react";

import { LOCALE_PREFIXES, localized, unlocalized } from "../lib/route";
import { useT } from "../lib/useT";
import type { SharedProps } from "../types";
import CategoryIcon from "./CategoryIcon";
import Dropdown from "./ui/Dropdown";

// Loaded on first open: keeps the palette and the query hooks off the first-load script.
const SearchPalette = lazy(() => import("./search/SearchPalette"));
// Same for the mobile menu sheet: the Base UI dialog chunk is ~21 KB on the wire.
const NavSheet = lazy(() => import("./NavSheet"));

const LOCALE_LABELS: Record<string, string> = { en: "EN", "pt-BR": "PT", es: "ES" };
// Endonyms: each language is named in itself, identical in every locale.
const LOCALE_NAMES: Record<string, string> = {
	en: "English",
	"pt-BR": "Português (Brasil)",
	es: "Español",
};
const SECTIONS = [
	["games", "nav.games", "text-accent-game", "game"],
	["movies", "nav.movies", "text-accent-movie", "movie"],
	["series", "nav.series", "text-accent-series", "series"],
	["books", "nav.books", "text-accent-book", "book"],
] as const;

// The owner picks one of resources/public/icons/hexascore-icon-{1,2,3}; app-icon.svg is a copy of the pick (see its README).
export const APP_ICON = "/icons/app-icon.svg";

const focus = "focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fg";
const item = `inline-flex min-h-11 items-center rounded-xs hover:text-primary ${focus}`;
// Pressed: 1 px down onto an inset shadow (the move is dropped under prefers-reduced-motion).
export const iconButton = `inline-flex size-11 shrink-0 items-center justify-center rounded-xs transition-[background-color,box-shadow,translate] duration-100 hover:bg-hover active:translate-y-px active:bg-surface active:shadow-pressed motion-reduce:transition-none motion-reduce:active:translate-y-0 ${focus}`;
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
	close: "M6 6 L18 18 M18 6 L6 18",
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
	const [menuOpen, setMenuOpen] = useState(false);
	// Mounted on intent (pointerdown/focus on the hamburger) or on open, so the chunk is usually there by the click.
	const [sheetMounted, setSheetMounted] = useState(false);
	if (menuOpen && !sheetMounted) setSheetMounted(true);
	const burger = useRef<HTMLButtonElement>(null);
	const sheet = useRef<HTMLDivElement>(null);
	const unreadBadge = unread > 0 && (
		<span className="absolute top-1 right-0 min-w-5 rounded-full bg-primary px-1 text-center text-[10px] font-bold leading-[18px] text-on-primary tabular-nums">
			{unreadLabel}
		</span>
	);

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

	// Base UI's focus guards and initial focus move focus in a requestAnimationFrame, so a Tab
	// handled before the next frame (keys right after opening, a busy main thread) reached the
	// page behind the sheet. Wrapping synchronously keeps focus inside on every keypress.
	useEffect(() => {
		if (!menuOpen) return;
		const onKey = (e: KeyboardEvent) => {
			const root = sheet.current;
			if (!root) {
				// The sheet chunk is still loading: hold focus on the hamburger and let Escape cancel.
				if (e.key === "Tab") e.preventDefault();
				if (e.key === "Escape") setMenuOpen(false);
				return;
			}
			if (e.key !== "Tab") return;
			const items = [...root.querySelectorAll<HTMLElement>("a[href],button:not([disabled])")];
			const active = document.activeElement;
			if (!root.contains(active) || active === (e.shiftKey ? items[0] : items.at(-1))) {
				e.preventDefault();
				(e.shiftKey ? items.at(-1) : items[0])?.focus();
			}
		};
		document.addEventListener("keydown", onKey);
		return () => document.removeEventListener("keydown", onKey);
	}, [menuOpen]);

	const languageList = (
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
	);

	const accountLinks = user && (
		<>
			<Link href={localized("/library", locale)} className={menuItem}>
				{t("nav.library")}
			</Link>
			<Link href={localized("/lists", locale)} className={menuItem}>
				{t("nav.lists")}
			</Link>
			{user.nickname && (
				<Link
					href={localized(`/u/${user.nickname.toLowerCase()}`, locale)}
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
		</>
	);

	const guestLinks = (
		<>
			<Link href={localized("/signup", locale)} className={menuItem}>
				{t("auth.signup.title")}
			</Link>
			<Link href={localized("/pricing", locale)} className={menuItem}>
				{t("nav.pricing")}
			</Link>
			<Link href="/docs" className={menuItem}>
				{t("nav.apiDocs")}
			</Link>
		</>
	);

	const avatar = user && (
		<span
			aria-hidden="true"
			className="inline-flex size-9 shrink-0 items-center justify-center rounded-full bg-fg font-bold text-bg"
		>
			{(user.name || "?").slice(0, 1).toUpperCase()}
		</span>
	);

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
			{languageList}
		</Dropdown>
	);

	return (
		<header
			data-scrolled={scrolled || undefined}
			className="sticky top-0 z-50 border-b border-fg bg-bg transition-shadow motion-reduce:transition-none data-scrolled:shadow-nav dark:data-scrolled:shadow-nav-dark"
		>
			<div className="shell flex h-14 items-center gap-2 lg:h-[72px] lg:gap-7">
				<Link
					href={localized("/", locale)}
					className={`${item} shrink-0 gap-2 font-display text-2xl font-bold hover:text-fg lg:text-[28px]`}
				>
					<img
						src={APP_ICON}
						alt=""
						width={32}
						height={32}
						className="size-7 lg:size-8"
					/>
					{page.props.appName}
				</Link>
				<nav
					aria-label={t("nav.sections")}
					className="hidden items-center gap-7 whitespace-nowrap lg:flex"
				>
					{SECTIONS.map(([path, key, , type]) => (
						<Link
							key={path}
							href={localized(`/${path}`, locale)}
							className={`${item} gap-1.5 text-[15px] font-medium`}
						>
							<CategoryIcon type={type} className="size-[18px]" />
							{t(key)}
						</Link>
					))}
					{feedHref && (
						<Link
							href={feedHref}
							aria-current={here === "/feed" ? "page" : undefined}
							className={`hidden min-h-11 items-center rounded-xs border-l border-border pl-7 text-[15px] font-medium hover:text-primary aria-[current=page]:font-bold xl:inline-flex ${focus}`}
						>
							{t("social.feed.title")}
						</Link>
					)}
				</nav>
				<div className="flex min-w-0 flex-1 items-center gap-1 lg:ml-auto lg:flex-none lg:gap-7">
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
						className={`flex h-11 min-w-0 flex-1 items-center justify-start gap-2 rounded-xs border border-field bg-field-bg px-3 text-muted hover:bg-hover lg:w-56 lg:flex-none xl:w-[280px] ${focus}`}
					>
						<Icon name="search" className="size-[18px] shrink-0" />
						<span className="truncate text-[15px]">{t("nav.searchShort")}</span>
						<kbd className="ml-auto hidden h-[22px] items-center rounded-sm border border-border bg-surface px-1.5 font-sans text-xs font-semibold lg:inline-flex">
							{hint}
						</kbd>
					</a>
					<div data-desktop="" className="hidden items-center gap-7 lg:flex">
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
									{unreadBadge}
								</Link>
								<Dropdown
									label={t("nav.account")}
									triggerClassName={`inline-flex min-h-11 items-center gap-2 rounded-xs pl-1 pr-1 hover:bg-hover ${focus}`}
									panelClassName="w-64"
									trigger={
										<>
											{avatar}
											{/* with the inline Feed link the 1152 px bar has no room for the name; the menu still shows it */}
											<span
												className={`hidden max-w-32 truncate font-semibold ${feedHref ? "" : "xl:inline"}`}
											>
												{user.nickname || user.name}
											</span>
										</>
									}
								>
									<p className="truncate px-3 py-2 text-sm font-semibold">
										{user.name}
									</p>
									{feedHref && (
										<Link href={feedHref} className={`${menuItem} xl:hidden`}>
											{t("social.feed.title")}
										</Link>
									)}
									{accountLinks}
								</Dropdown>
							</div>
						) : (
							auth && (
								<div className="flex items-center gap-4">
									<Link
										href={localized("/login", locale)}
										className={`${item} whitespace-nowrap px-1 text-[15px] font-semibold text-primary hover:underline`}
									>
										{t("nav.login")}
									</Link>
									<Dropdown
										label={t("nav.more")}
										triggerClassName={`inline-flex min-h-11 min-w-11 items-center justify-center gap-1.5 rounded-xs px-2 hover:bg-hover ${focus}`}
										trigger={
											<>
												<Icon name="menu" />
												<span className="hidden font-semibold xl:inline">
													{t("nav.more")}
												</span>
											</>
										}
									>
										{guestLinks}
									</Dropdown>
								</div>
							)
						)}
					</div>
				</div>
				<button
					ref={burger}
					type="button"
					aria-haspopup="dialog"
					aria-expanded={menuOpen}
					aria-label={
						user && unread > 0 ? t("nav.menuUnread", { count: unread }) : t("nav.menu")
					}
					onClick={() => setMenuOpen(true)}
					onPointerDown={() => setSheetMounted(true)}
					onFocus={() => setSheetMounted(true)}
					className={`${iconButton} relative lg:hidden`}
				>
					<Icon name="menu" />
					{user && unreadBadge}
				</button>
			</div>
			{sheetMounted && (
				<Suspense fallback={null}>
					<NavSheet
						open={menuOpen}
						onOpenChange={setMenuOpen}
						sheetRef={sheet}
						finalFocus={burger}
						title={t("nav.menu")}
						closeLabel={t("nav.closeMenu")}
					>
						{user && (
							<div className="flex items-center gap-3 border-b border-border px-4 py-3">
								{avatar}
								<span className="min-w-0 truncate font-semibold">{user.name}</span>
							</div>
						)}
						<nav aria-label={t("nav.sections")} className="border-b border-border py-1">
							{SECTIONS.map(([path, key, accent, type]) => (
								<Link
									key={path}
									href={localized(`/${path}`, locale)}
									className={`${menuItem} font-semibold ${accent}`}
								>
									<CategoryIcon type={type} />
									{t(key)}
								</Link>
							))}
							{feedHref && (
								<Link
									href={feedHref}
									aria-current={here === "/feed" ? "page" : undefined}
									className={`${menuItem} font-semibold`}
								>
									{t("social.feed.title")}
								</Link>
							)}
						</nav>
						{user ? (
							<div className="border-b border-border py-1">
								<Link
									href={localized("/me/notifications", locale)}
									className={menuItem}
								>
									<Icon name="bell" />
									<span className="flex-1">{t("nav.notifications")}</span>
									{unread > 0 && (
										<span className="min-w-5 rounded-full bg-primary px-1 text-center text-xs font-bold leading-5 text-on-primary tabular-nums">
											{unreadLabel}
										</span>
									)}
								</Link>
								{accountLinks}
							</div>
						) : (
							auth && (
								<div className="border-b border-border py-1">
									<Link
										href={localized("/login", locale)}
										className={`${menuItem} font-semibold text-primary`}
									>
										{t("nav.login")}
									</Link>
									{guestLinks}
								</div>
							)
						)}
						<div className="border-b border-border py-1">
							<p className="px-3 pt-2 pb-1 text-xs font-bold text-muted uppercase">
								{t("nav.language")}
							</p>
							{languageList}
						</div>
						<div className="py-1">
							<button
								type="button"
								data-theme-toggle=""
								onClick={toggleTheme}
								className={menuItem}
							>
								<Icon name="moon" className="size-5 dark:hidden" />
								<Icon name="sun" className="hidden size-5 dark:block" />
								{t("nav.theme")}
							</button>
						</div>
					</NavSheet>
				</Suspense>
			)}
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
