import { usePage } from "@inertiajs/react";
import type { SharedProps } from "../types";

export type CoverSize = "sm" | "md";

export interface Cover {
	thumb: string;
	full: string;
}

const DIMENSIONS: Record<CoverSize, { width: number; height: number }> = {
	sm: { width: 32, height: 48 },
	md: { width: 44, height: 64 },
};

// CoverThumb is the list-row cover: 32×48 or 44×64, 96w/600w sources, lazy.
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
	const { coversBase } = usePage<SharedProps>().props;

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
			src={`${coversBase}${cover.thumb}`}
			srcSet={`${coversBase}${cover.thumb} 96w, ${coversBase}${cover.full} 600w`}
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
