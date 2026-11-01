export type CoverSize = "sm" | "md";

const DIMENSIONS: Record<CoverSize, { width: number; height: number }> = {
	sm: { width: 32, height: 48 },
	md: { width: 44, height: 64 },
};

// CoverThumb is the list-row cover: 32×48 or 44×64, 96w/600w sources, lazy.
// The item page cover passes priority for eager loading with high fetch priority.
export default function CoverThumb({
	src96,
	src600,
	alt,
	size = "sm",
	priority = false,
}: {
	src96: string;
	src600: string;
	alt: string;
	size?: CoverSize;
	priority?: boolean;
}) {
	const { width, height } = DIMENSIONS[size];

	return (
		<img
			src={src96}
			srcSet={`${src96} 96w, ${src600} 600w`}
			sizes={`${width}px`}
			width={width}
			height={height}
			alt={alt}
			loading={priority ? "eager" : "lazy"}
			fetchPriority={priority ? "high" : "auto"}
			decoding="async"
			className="shrink-0 rounded-sm bg-surface object-cover"
		/>
	);
}
