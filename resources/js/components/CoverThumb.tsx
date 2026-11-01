import { usePage } from "@inertiajs/react";
import type { SharedProps } from "../types";

export type CoverSize = "sm" | "md" | "trend" | "lg" | "card";

export interface Cover {
	thumb: string;
	full: string;
}

const DIMENSIONS: Record<CoverSize, { width: number; height: number }> = {
	sm: { width: 32, height: 48 },
	md: { width: 44, height: 64 },
	trend: { width: 96, height: 144 },
	lg: { width: 240, height: 360 },
	card: { width: 160, height: 240 },
};

const SIZES: Record<CoverSize, string> = {
	sm: "32px",
	md: "44px",
	trend: "(min-width: 48rem) 96px, 80px",
	lg: "(min-width: 48rem) 240px, 96px",
	card: "(min-width: 48rem) 160px, 30vw",
};

// Tailwind classes instead of a style prop: SSR would emit a style attribute, blocked by style-src.
const BOX: Record<CoverSize, string> = {
	sm: "w-8 h-12",
	md: "w-11 h-16",
	trend: "w-20 h-30 md:w-24 md:h-36",
	lg: "w-24 h-36 md:w-60 md:h-90",
	card: "w-full h-auto aspect-[2/3]",
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
	const { coversBase } = usePage<SharedProps>().props;

	if (!cover) {
		return (
			<span
				aria-hidden="true"
				data-testid="cover-placeholder"
				className={`inline-block shrink-0 rounded-sm bg-surface ${BOX[size]}`}
			/>
		);
	}

	return (
		<img
			src={`${coversBase}${size === "lg" ? cover.full : cover.thumb}`}
			srcSet={`${coversBase}${cover.thumb} 96w, ${coversBase}${cover.full} 600w`}
			sizes={SIZES[size]}
			width={width}
			height={height}
			alt={alt}
			loading={priority ? "eager" : "lazy"}
			fetchPriority={priority ? "high" : "auto"}
			decoding="async"
			className={`shrink-0 rounded-sm bg-surface object-cover ${BOX[size]}`}
		/>
	);
}
