import { getLocale, getTranslations } from "next-intl/server";
import { getPathname, Link } from "@/i18n/navigation";
import { isAdmin } from "@/lib/member";
import { getSessionUser } from "@/lib/session";
import { LocaleMenu } from "./locale-menu";
import { NavIcon } from "./nav-icons";
import { NavShell } from "./nav-shell";
import { SearchAutocomplete } from "./search-autocomplete";
import { ThemeToggle } from "./theme-toggle";
import { UserMenu } from "./user-menu";

const LINKS = [
	{ href: "/games", key: "games" },
	{ href: "/movies", key: "movies" },
	{ href: "/series", key: "series" },
	{ href: "/books", key: "books" },
	{ href: "/docs", key: "api" },
	{ href: "/pricing", key: "pricing" },
] as const;

export async function Navbar() {
	const [t, locale, user] = await Promise.all([getTranslations("nav"), getLocale(), getSessionUser()]);

	return (
		<header className="sticky top-0 z-40 border-b border-border bg-bg/95 backdrop-blur supports-[backdrop-filter]:bg-bg/80">
			<NavShell
				labels={{ search: t("openSearch"), menu: t("menu") }}
				brand={
					// The brand is final ("HexaScore"): two-tone wordmark, green "Hexa" + blue "Score".
					<Link href="/" className="font-serif text-2xl font-extrabold tracking-tight">
						<span className="text-emerald-700 dark:text-emerald-400">Hexa</span>
						<span className="text-blue-600 dark:text-blue-400">Score</span>
					</Link>
				}
				search={
					<SearchAutocomplete action={getPathname({ href: "/search", locale })} placeholder={t("search")} />
				}
				actions={
					<>
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
							<Link
								href="/login"
								className="grid min-h-11 place-items-center rounded-md bg-fg px-3 text-sm font-medium text-bg"
							>
								{t("login")}
							</Link>
						)}
						<LocaleMenu label={t("language")} />
					</>
				}
				links={
					<nav
						aria-label={t("sections")}
						className="grid text-sm sm:order-last sm:flex sm:w-full sm:justify-between"
					>
						{LINKS.map((l) => (
							<Link
								key={l.key}
								href={l.href}
								className="flex min-h-11 items-center gap-2 text-muted underline-offset-4 hover:text-fg hover:underline sm:min-h-9 sm:gap-1.5"
							>
								<NavIcon name={l.key} />
								{t(l.key)}
							</Link>
						))}
					</nav>
				}
			/>
		</header>
	);
}
