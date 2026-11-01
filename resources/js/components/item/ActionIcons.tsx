// Stroke icons for the item page action buttons; decorative, the button text names the action.
const svg =
	"size-5 shrink-0 fill-none stroke-current stroke-2 [stroke-linecap:round] [stroke-linejoin:round]";

export function ChartIcon(): React.ReactElement {
	return (
		<svg viewBox="0 0 24 24" aria-hidden="true" className={svg}>
			<path d="M4 20V10M10 20V4M16 20v-7M22 20H2" />
		</svg>
	);
}

export function StarIcon(): React.ReactElement {
	return (
		<svg viewBox="0 0 24 24" aria-hidden="true" className={svg}>
			<path d="m12 3 2.7 5.6 6.1.9-4.4 4.3 1 6.1L12 17l-5.4 2.9 1-6.1-4.4-4.3 6.1-.9z" />
		</svg>
	);
}

export function ListIcon(): React.ReactElement {
	return (
		<svg viewBox="0 0 24 24" aria-hidden="true" className={svg}>
			<path d="M9 6h11M9 12h11M9 18h11M4 6h.01M4 12h.01M4 18h.01" />
		</svg>
	);
}

export function PencilIcon(): React.ReactElement {
	return (
		<svg viewBox="0 0 24 24" aria-hidden="true" className={svg}>
			<path d="M4 20h4L19 9l-4-4L4 16z" />
		</svg>
	);
}

export function HistoryIcon(): React.ReactElement {
	return (
		<svg viewBox="0 0 24 24" aria-hidden="true" className={svg}>
			<path d="M3 12a9 9 0 1 0 3-6.7L3 8M3 3v5h5M12 7v5l3 2" />
		</svg>
	);
}
