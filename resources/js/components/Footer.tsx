import { Link, usePage } from "@inertiajs/react";

import { openCookieSettings } from "../lib/preferences";
import { localized } from "../lib/route";
import { useT } from "../lib/useT";
import type { SharedProps } from "../types";

export default function Footer() {
	const t = useT();
	const { appName, locale } = usePage<SharedProps>().props;
	const link =
		"inline-flex min-h-11 items-center justify-center rounded-xs text-[15px] hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fg md:text-[13px]";
	const links: [string, string][] = [
		[localized("/terms", locale), "footer.terms"],
		[localized("/policy", locale), "footer.privacy"],
		[localized("/contact", locale), "footer.contact"],
		["/docs", "footer.apiDocs"],
		[localized("/pricing", locale), "footer.pricing"],
		[localized("/help/score", locale), "footer.howScores"],
	];

	// Language lives in the navbar only (design v1, round 2 item 14).
	return (
		<footer className="mt-12 border-t border-fg">
			<div className="shell flex flex-col items-center gap-2 pt-4 pb-6 md:min-h-[72px] md:flex-row md:gap-6 md:py-0">
				<nav aria-label={t("footer.label")} className="w-full md:w-auto">
					<ul className="flex flex-col items-center md:flex-row md:gap-6">
						{links.map(([href, key]) => (
							<li key={key}>
								<Link href={href} className={link}>
									{t(key)}
								</Link>
							</li>
						))}
						<li>
							<button type="button" onClick={openCookieSettings} className={link}>
								{t("footer.cookieSettings")}
							</button>
						</li>
					</ul>
				</nav>
				<p className="text-[13px] text-muted md:ml-auto">
					© {new Date().getFullYear()} {appName}
				</p>
			</div>
		</footer>
	);
}
