function parts(version: string): number[] {
	return (version.split("-")[0] ?? "").split(".").map((part) => Number.parseInt(part, 10) || 0);
}

// Numeric MAJOR.MINOR.PATCH comparison; pre-release suffixes are ignored (store builds are plain SemVer).
export function compareVersions(a: string, b: string): number {
	const left = parts(a);
	const right = parts(b);
	for (let i = 0; i < 3; i++) {
		const diff = (left[i] ?? 0) - (right[i] ?? 0);
		if (diff !== 0) return diff;
	}
	return 0;
}

export function isBelowMinimum(running: string, minimum: string | null): boolean {
	return minimum !== null && compareVersions(minimum, running) > 0;
}
