export type CoverSize = "sm" | "md" | "lg";

export interface Cover {
	thumb: string;
	full: string;
}

const DIMENSIONS: Record<CoverSize, { width: number; height: number }> = {
	sm: { width: 32, height: 48 },
	md: { width: 44, height: 64 },
	lg: { width: 144, height: 208 },
};

// CoverThumb is the list-row cover (32×48 or 44×64; "lg" 144×208 for the item page header), 96w/600w sources, lazy.
// The item page cover passes priority for eager loading with high fetch priority.
// Without a cover it keeps the same box so the row does not shift.
export default function CoverThumb({
	cover,
	alt,
	size = "sm",
	priority = false,
}: {
	cover: Cover | null | undefined;
	alt: string;
	size?: CoverSize;
	priority?: boolean;
}) {
	const { width, height } = DIMENSIONS[size];

	if (!cover) {
		return (
			<span
				aria-hidden="true"
				data-testid="cover-placeholder"
				style={{ width, height }}
				className="inline-block shrink-0 rounded-sm bg-surface"
			/>
		);
	}

	return (
		<img
			src={size === "lg" ? cover.full : cover.thumb}
			srcSet={`${cover.thumb} 96w, ${cover.full} 600w`}
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
