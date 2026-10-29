export type Mode = "light" | "dark";

const toLinear = (c: number): number => c ** 3;

/** "L C h" (OKLCH, h in degrees) to linear sRGB, clamped to the gamut. */
export function oklchToLinearRgb(value: string): [number, number, number] {
	const [l, c, h] = value.split(" ").map(Number);
	const a = c * Math.cos((h * Math.PI) / 180);
	const b = c * Math.sin((h * Math.PI) / 180);
	const lp = toLinear(l + 0.3963377774 * a + 0.2158037573 * b);
	const mp = toLinear(l - 0.1055613458 * a - 0.0638541728 * b);
	const sp = toLinear(l - 0.0894841775 * a - 1.291485548 * b);
	const rgb = [
		4.0767416621 * lp - 3.3077115913 * mp + 0.2309699292 * sp,
		-1.2684380046 * lp + 2.6097574011 * mp - 0.3413193965 * sp,
		-0.0041960863 * lp - 0.7034186147 * mp + 1.707614701 * sp,
	];
	return rgb.map((v) => Math.min(1, Math.max(0, v))) as [number, number, number];
}

const gamma = (v: number): number => (v <= 0.0031308 ? 12.92 * v : 1.055 * v ** (1 / 2.4) - 0.055);

export function oklchToHex(value: string): string {
	return `#${oklchToLinearRgb(value)
		.map((v) =>
			Math.round(gamma(v) * 255)
				.toString(16)
				.padStart(2, "0"),
		)
		.join("")}`;
}

function luminance(value: string): number {
	const [r, g, b] = oklchToLinearRgb(value);
	return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

/** WCAG 2.x contrast ratio between two "L C h" colours. */
export function contrast(a: string, b: string): number {
	const [hi, lo] = [luminance(a), luminance(b)].sort((x, y) => y - x);
	return (hi + 0.05) / (lo + 0.05);
}
