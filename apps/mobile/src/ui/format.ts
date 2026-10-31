export function formatScore(score: number | null, locale: string): string {
	if (score === null) return "–";
	return new Intl.NumberFormat(locale, {
		minimumFractionDigits: 1,
		maximumFractionDigits: 1,
	}).format(score);
}
