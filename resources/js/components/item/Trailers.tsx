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
		<div className="grid gap-3">
			<p className="text-sm text-muted">{t("item.trailerConsent")}</p>
			<ul className="grid gap-3 sm:grid-cols-2">
				{trailers.map((trailer, i) => {
					const title = trailer.title || t("item.trailerTitle", { n: i + 1 });
					return (
						<li
							key={trailer.youtubeId}
							className="aspect-video overflow-hidden rounded-md border border-border"
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
								<button
									type="button"
									onClick={() =>
										setLoaded((prev) => new Set(prev).add(trailer.youtubeId))
									}
									className="grid size-full place-content-center gap-1 bg-surface text-sm hover:bg-border"
								>
									<span aria-hidden="true" className="text-3xl">
										▶
									</span>
									<span className="font-medium">{t("item.loadTrailer")}</span>
									<span className="text-xs text-muted">{title}</span>
								</button>
							)}
						</li>
					);
				})}
			</ul>
		</div>
	);
}
