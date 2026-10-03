import { getTranslations } from "next-intl/server";
import { Link } from "@/i18n/navigation";
import { APP_NAME } from "@/lib/config";
import { ThemeToggle } from "./theme-toggle";

const SECTIONS = [
	{ href: "/games", key: "games", accent: "decoration-game" },
	{ href: "/movies", key: "movies", accent: "decoration-movie" },
	{ href: "/series", key: "series", accent: "decoration-series" },
	{ href: "/books", key: "books", accent: "decoration-book" },
] as const;

export async function Navbar() {
	const t = await getTranslations("nav");

	return (
		<header className="border-b border-border">
			<div className="mx-auto flex max-w-6xl flex-wrap items-center gap-3 px-4 py-3">
				<Link href="/" className="font-serif text-xl font-bold tracking-tight">
					{APP_NAME}
				</Link>
				<search className="">
					<form action="/search">
						<label className="sr-only" htmlFor="q">
							{t("search")}
						</label>
						<input
							id="q"
							name="q"
							type="search"
							placeholder={t("search")}
							className="min-h-11 w-full rounded-md border border-border bg-bg px-3"
						/>
					</form>
				</search>
				<nav aria-label="Sections" className="flex gap-3 text-sm">
					{SECTIONS.map((s) => (
						<Link key={s.key} href={s.href} className={`underline-offset-4 hover:underline ${s.accent}`}>
							{t(s.key)}
						</Link>
					))}
				</nav>
				<div className="ml-auto flex items-center gap-2">
					<ThemeToggle label={t("theme")} />
					<Link href="/login" className="text-sm">
						{t("login")}
					</Link>
				</div>
			</div>
		</header>
	);
}
