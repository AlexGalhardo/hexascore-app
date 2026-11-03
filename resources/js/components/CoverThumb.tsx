import { usePage } from "@inertiajs/react";
import type { SharedProps } from "../types";

export type CoverSize = "sm" | "md" | "trend" | "lg" | "card" | "compact";

export interface Cover {
	thumb: string;
	full: string;
}

const DIMENSIONS: Record<CoverSize, { width: number; height: number }> = {
	sm: { width: 32, height: 48 },
	md: { width: 44, height: 64 },
	trend: { width: 120, height: 180 },
	lg: { width: 240, height: 360 },
	card: { width: 160, height: 240 },
	compact: { width: 160, height: 213 },
};

const SIZES: Record<CoverSize, string> = {
	sm: "32px",
	md: "44px",
	trend: "(min-width: 48rem) 120px, 100px",
	lg: "(min-width: 40rem) 240px, 100vw",
	card: "(min-width: 48rem) 160px, 30vw",
	compact: "(min-width: 48rem) 25vw, 45vw",
};

// Tailwind classes instead of a style prop: SSR would emit a style attribute, blocked by style-src.
const BOX: Record<CoverSize, string> = {
	sm: "w-8 h-12",
	md: "w-11 h-16",
	trend: "w-25 h-37.5 md:w-30 md:h-45",
	lg: "w-full h-auto aspect-[2/3]",
	card: "w-full h-auto aspect-[2/3]",
	// Shorter than the poster ratio, for the landing carousels; the whole cover still shows (object-contain).
	compact: "w-full h-auto aspect-[3/4]",
};

// CoverThumb is the list-row cover (32×48 or 44×64; "lg" for the item page), shown whole at every size, 96w/600w sources, lazy.
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
			// Every cover is shown whole (object-contain); the transparent letterbox takes the colour of the card around it.
			className={`shrink-0 rounded-sm object-contain ${BOX[size]}`}
		/>
	);
}
