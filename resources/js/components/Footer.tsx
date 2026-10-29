import { Link, usePage } from "@inertiajs/react";

import { localized } from "../lib/route";
import { useT } from "../lib/useT";
import type { SharedProps } from "../types";

export default function Footer() {
	const t = useT();
	const { appName, locale } = usePage<SharedProps>().props;
	const link = "inline-flex min-h-11 items-center hover:underline";

	return (
		<footer className="mt-12 border-t border-border text-sm text-muted">
			<div className="mx-auto flex max-w-6xl flex-wrap items-center gap-x-6 px-4 py-2">
				<p>
					© {new Date().getFullYear()} {appName}. {t("footer.rights")}
				</p>
				<nav className="ml-auto flex gap-4">
					<Link href={localized("/terms", locale)} className={link}>
						{t("footer.terms")}
					</Link>
					<Link href={localized("/privacy", locale)} className={link}>
						{t("footer.privacy")}
					</Link>
					<Link href={localized("/contact", locale)} className={link}>
						{t("footer.contact")}
					</Link>
				</nav>
			</div>
		</footer>
	);
}
