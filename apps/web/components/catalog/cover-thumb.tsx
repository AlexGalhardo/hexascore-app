/** 2:3 cover thumbnail for lists and suggestions; an empty frame keeps rows aligned when there is no cover. */
export function CoverThumb({ src, size = "sm" }: { src: string | null; size?: "sm" | "md" }) {
	const box = size === "md" ? "h-16 w-11" : "h-12 w-8";
	if (!src)
		return <span aria-hidden="true" className={`${box} shrink-0 rounded-sm border border-border bg-surface`} />;
	return (
		// biome-ignore lint/performance/noImgElement: remote covers from ingestion; next/image needs a host allowlist (later)
		<img
			src={src}
			alt=""
			width={size === "md" ? 44 : 32}
			height={size === "md" ? 64 : 48}
			loading="lazy"
			decoding="async"
			className={`${box} shrink-0 rounded-sm border border-border object-cover`}
		/>
	);
}
