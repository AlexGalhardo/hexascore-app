import { Link, usePage } from "@inertiajs/react";
import { fieldClass } from "../components/Field";
import Seo from "../components/Seo";
import Btn, { btnClass } from "../components/ui/Btn";
import { localized } from "../lib/route";
import { useT } from "../lib/useT";
import type { SharedProps } from "../types";

// Rendered by the server for errors on Inertia visits and for unmatched
// pages (404, 403, 500…). message is safe to show: 5xx only ever carry the
// status text, so the page shows the catalogue copy for 404 and 5xx.
export default function ErrorPage({ status, message }: { status: number; message: string }) {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	const missing = status === 404;
	const server = status >= 500;
	const heading = missing
		? t("errors.notFoundTitle")
		: server
			? t("errors.serverTitle")
			: message;
	const body = missing ? t("errors.notFoundBody") : server ? t("errors.serverBody") : "";
	const home = (
		<Link href={localized("/", locale)} className={btnClass(server ? "secondary" : "primary")}>
			{t("errors.home")}
		</Link>
	);

	return (
		<>
			<Seo title={`${status} ${heading}`} noindex />
			<section className="mx-auto flex max-w-2xl flex-col items-center py-16 text-center md:py-24">
				<p className="font-display text-8xl leading-none text-primary md:text-[10rem]">
					{status}
				</p>
				<div aria-hidden="true" className="mt-8 h-0.5 w-48 bg-fg" />
				<h1 className="mt-6 text-4xl md:text-5xl">{heading}</h1>
				{body && <p className="mt-4 text-lg text-muted">{body}</p>}
				<div className="mt-8 flex w-full flex-col items-center justify-center gap-3 sm:flex-row">
					{server ? (
						<>
							<Btn onClick={() => window.location.reload()}>{t("errors.retry")}</Btn>
							{home}
						</>
					) : (
						<>
							{home}
							<search className="w-full sm:w-90">
								<form action={localized("/search", locale)} method="get">
									<label htmlFor="error-search" className="sr-only">
										{t("search.label")}
									</label>
									<input
										id="error-search"
										type="search"
										name="q"
										placeholder={t("nav.search")}
										className={fieldClass}
									/>
								</form>
							</search>
						</>
					)}
				</div>
			</section>
		</>
	);
}
