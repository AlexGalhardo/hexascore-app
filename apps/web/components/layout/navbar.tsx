import { getLocale, getTranslations } from "next-intl/server";
import { getPathname, Link } from "@/i18n/navigation";
import { APP_NAME } from "@/lib/config";
import { isAdmin } from "@/lib/member";
import { getSessionUser } from "@/lib/session";
import { SearchAutocomplete } from "./search-autocomplete";
import { ThemeToggle } from "./theme-toggle";
import { UserMenu } from "./user-menu";

const SECTIONS = [
	{ href: "/games", key: "games", accent: "decoration-game" },
	{ href: "/movies", key: "movies", accent: "decoration-movie" },
	{ href: "/series", key: "series", accent: "decoration-series" },
	{ href: "/books", key: "books", accent: "decoration-book" },
] as const;

export async function Navbar() {
	const [t, locale, user] = await Promise.all([getTranslations("nav"), getLocale(), getSessionUser()]);

	return (
		<header className="border-b border-border">
			<div className="mx-auto flex max-w-6xl flex-wrap items-center gap-3 px-4 py-3">
				<Link href="/" className="font-serif text-xl font-bold tracking-tight">
					{APP_NAME}
				</Link>
				<search className="w-full sm:w-72">
					<SearchAutocomplete action={getPathname({ href: "/search", locale })} placeholder={t("search")} />
				</search>
				<nav aria-label="Sections" className="flex gap-3 text-sm">
					{SECTIONS.map((s) => (
						<Link key={s.key} href={s.href} className={`underline-offset-4 hover:underline ${s.accent}`}>
							{t(s.key)}
						</Link>
					))}
				</nav>
				<div className="ml-auto flex items-center gap-2">
					<Link href="/docs" className="text-sm underline-offset-4 hover:underline">
						{t("docs")}
					</Link>
					<ThemeToggle label={t("theme")} />
					{user ? (
						<UserMenu
							initial={(user.username ?? user.name).charAt(0).toUpperCase() || "?"}
							nickname={user.username}
							admin={isAdmin(user)}
							labels={{
								account: t("account"),
								profile: t("profile"),
								api: t("api"),
								logout: t("logout"),
								feed: t("feed"),
								publicProfile: t("publicProfile"),
							}}
						/>
					) : (
						<>
							<Link href="/login" className="grid min-h-11 place-items-center px-2 text-sm">
								{t("login")}
							</Link>
							<Link
								href="/signup"
								className="grid min-h-11 place-items-center rounded-md bg-fg px-3 text-sm font-medium text-bg"
							>
								{t("signup")}
							</Link>
						</>
					)}
				</div>
			</div>
		</header>
	);
}
