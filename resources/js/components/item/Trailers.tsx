import { useEffect, useState } from "react";

import { CONSENT_COOKIE, onConsentChange, readPreferenceCookie } from "../../lib/preferences";
import { useT } from "../../lib/useT";

type Trailer = { youtubeId: string; title?: string };

// Click-to-load: nothing is requested from YouTube until the visitor allowed embeds or clicks.
export default function Trailers({ trailers }: { trailers: Trailer[] }) {
	const t = useT();
	const [loaded, setLoaded] = useState<Set<string>>(new Set());

	useEffect(() => {
		function apply(): void {
			if (readPreferenceCookie(CONSENT_COOKIE) === "all") {
				setLoaded(new Set(trailers.map((tr) => tr.youtubeId)));
			}
		}
		apply();
		return onConsentChange(apply);
	}, [trailers]);

	return (
		<ul className="grid gap-3 sm:grid-cols-2">
			{trailers.map((trailer, i) => {
				const title = trailer.title || t("item.trailerTitle", { n: i + 1 });
				return (
					<li
						key={trailer.youtubeId}
						className="aspect-video overflow-hidden bg-fg text-bg"
					>
						{loaded.has(trailer.youtubeId) ? (
							<iframe
								src={`https://www.youtube-nocookie.com/embed/${trailer.youtubeId}`}
								title={title}
								className="size-full"
								loading="lazy"
								referrerPolicy="strict-origin-when-cross-origin"
								allow="accelerometer; encrypted-media; gyroscope; picture-in-picture"
								allowFullScreen
							/>
						) : (
							<div className="grid size-full place-content-center justify-items-center gap-3 p-4 text-center text-sm">
								<svg
									viewBox="0 0 24 24"
									aria-hidden="true"
									className="size-5 fill-none stroke-current stroke-2"
								>
									<rect x="5" y="11" width="14" height="10" rx="1" />
									<path d="M8 11V7a4 4 0 0 1 8 0v4" />
								</svg>
								<p className="max-w-xs">{t("item.trailerConsent")}</p>
								<button
									type="button"
									aria-label={`${t("item.loadTrailerN", { n: i + 1 })}: ${title}`}
									onClick={() =>
										setLoaded((prev) => new Set(prev).add(trailer.youtubeId))
									}
									className="min-h-11 rounded-xs bg-bg px-4 font-semibold text-fg hover:bg-hover focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-bg"
								>
									{t("item.loadTrailerN", { n: i + 1 })}
								</button>
							</div>
						)}
					</li>
				);
			})}
		</ul>
	);
}
