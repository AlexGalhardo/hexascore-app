import { Link, usePage } from "@inertiajs/react";
import { openCookieSettings } from "../lib/preferences";
import { localized } from "../lib/route";
import { useT } from "../lib/useT";
import type { SharedProps } from "../types";
import { APP_ICON, BRAND } from "./Nav";

const SOURCE_URL = "https://github.com/AlexGalhardo/hexascore-app";
const linkClass =
	"inline-flex min-h-9 items-center text-sm text-footer-fg hover:underline hover:underline-offset-4 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-footer-fg";
const headingClass = "mb-4 text-xs font-bold tracking-[0.08em] text-footer-muted uppercase";

export default function Footer() {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	const l = (path: string) => localized(path, locale);
	const categories = [
		["/games", "nav.games"],
		["/movies", "nav.movies"],
		["/series", "nav.series"],
		["/books", "nav.books"],
	] as const;
	const company = [
		["/about", "footer.about"],
		["/contact", "footer.contact"],
		["/pricing", "footer.pricing"],
		["/how-score-works", "footer.howScores"],
	] as const;
	const legal = [
		["/terms", "footer.terms"],
		["/privacy", "footer.privacy"],
	] as const;

	return (
		<footer className="mt-16 bg-footer-bg text-footer-fg">
			<div className="shell py-12 md:py-16">
				<div className="flex flex-col gap-6 lg:flex-row lg:items-center lg:justify-between">
					<div>
						<h2 className={headingClass}>{t("footer.newsletterTitle")}</h2>
						<p id="newsletter-soon" className="text-sm text-footer-muted">
							{t("footer.newsletterSoon")}
						</p>
					</div>
					<div className="flex w-full gap-2 lg:max-w-xl">
						<input
							type="email"
							disabled
							aria-label={t("footer.newsletterPlaceholder")}
							aria-describedby="newsletter-soon"
							placeholder={t("footer.newsletterPlaceholder")}
							className="min-h-11 min-w-0 flex-1 rounded-sm border border-footer-border bg-footer-border/30 px-3 text-sm text-footer-fg placeholder:text-footer-muted disabled:cursor-not-allowed"
						/>
						<button
							type="button"
							disabled
							className="min-h-11 shrink-0 rounded-sm bg-footer-fg px-5 text-sm font-bold text-footer-bg disabled:cursor-not-allowed disabled:opacity-60"
						>
							{t("footer.newsletterSubmit")}
						</button>
					</div>
				</div>

				<div className="grid gap-10 pt-10 sm:grid-cols-2 lg:grid-cols-[minmax(0,1.8fr)_repeat(4,minmax(0,1fr))] lg:gap-8">
					<div className="sm:col-span-2 lg:col-span-1">
						<Link
							href={l("/")}
							className="inline-flex items-center gap-3 text-2xl font-bold focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-footer-fg"
						>
							<img src={APP_ICON} alt="" width={36} height={36} className="size-9" />
							{BRAND}
						</Link>
						<p className="mt-5 max-w-md text-sm leading-7 text-footer-muted">
							{t("footer.description")}
							<span className="block mt-2">{t("footer.affiliateDisclosure")}</span>
						</p>
					</div>
					<nav
						aria-label={t("footer.label")}
						className="grid gap-8 sm:col-span-2 sm:grid-cols-2 lg:col-span-4 lg:grid-cols-4"
					>
						<div>
							<h2 className={headingClass}>{t("footer.categoriesTitle")}</h2>
							<ul className="grid gap-1">
								{categories.map(([href, key]) => (
									<li key={href}>
										<Link href={l(href)} className={linkClass}>
											{t(key)}
										</Link>
									</li>
								))}
							</ul>
						</div>
						<div>
							<h2 className={headingClass}>{t("footer.companyTitle")}</h2>
							<ul className="grid gap-1">
								{company.map(([href, key]) => (
									<li key={href}>
										<Link href={l(href)} className={linkClass}>
											{t(key)}
										</Link>
									</li>
								))}
								<li>
									<a
										href={SOURCE_URL}
										target="_blank"
										rel="noopener noreferrer"
										className={linkClass}
									>
										{t("footer.sourceCode")}
									</a>
								</li>
							</ul>
						</div>
						<div>
							<h2 className={headingClass}>{t("footer.apiTitle")}</h2>
							<ul className="grid gap-1">
								<li>
									<Link href="/docs" className={linkClass}>
										{t("footer.apiDocs")}
									</Link>
								</li>
								<li>
									<Link href="/docs/reference" className={linkClass}>
										{t("footer.apiReference")}
									</Link>
								</li>
							</ul>
						</div>
						<div>
							<h2 className={headingClass}>{t("footer.legalTitle")}</h2>
							<ul className="grid gap-1">
								{legal.map(([href, key]) => (
									<li key={href}>
										<Link href={l(href)} className={linkClass}>
											{t(key)}
										</Link>
									</li>
								))}
								<li>
									<button
										type="button"
										onClick={openCookieSettings}
										className={linkClass}
									>
										{t("footer.cookieSettings")}
									</button>
								</li>
							</ul>
						</div>
					</nav>
				</div>
			</div>
		</footer>
	);
}
