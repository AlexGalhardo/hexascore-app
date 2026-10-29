import { Link, usePage } from "@inertiajs/react";

import Seo from "../components/Seo";
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

	return (
		<>
			<Seo title={`${status} ${heading}`} noindex />
			<section className="py-12 text-center">
				<p className="text-sm font-semibold text-accent-game">{status}</p>
				<h1 className="mt-2 text-3xl font-semibold tracking-tight">{heading}</h1>
				{body && <p className="mt-3 text-muted">{body}</p>}
				<Link
					href={localized("/", locale)}
					className="mt-6 inline-flex min-h-11 items-center text-accent-game hover:underline"
				>
					{t("errors.home")}
				</Link>
			</section>
		</>
	);
}
