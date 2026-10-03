type Props = { src: string | null; thumb?: string | null; size?: "sm" | "md" };

/** 2:3 cover thumbnail for lists and suggestions; an empty frame keeps rows aligned when there is no cover. */
export function CoverThumb({ src, thumb = null, size = "sm" }: Props) {
	const box = size === "md" ? "h-16 w-11" : "h-12 w-8";
	const width = size === "md" ? 44 : 32;
	const url = thumb ?? src;
	if (!url)
		return <span aria-hidden="true" className={`${box} shrink-0 rounded-sm border border-border bg-surface`} />;
	return (
		// biome-ignore lint/performance/noImgElement: remote covers from ingestion; next/image needs a host allowlist (later)
		<img
			src={url}
			// Up to ~2x DPR the browser picks the 96w thumb; denser screens get the 600w full cover.
			srcSet={thumb && src ? `${thumb} 96w, ${src} 600w` : undefined}
			sizes={thumb && src ? `${width}px` : undefined}
			alt=""
			width={width}
			height={size === "md" ? 64 : 48}
			loading="lazy"
			decoding="async"
			className={`${box} shrink-0 rounded-sm border border-border object-cover`}
		/>
	);
}
