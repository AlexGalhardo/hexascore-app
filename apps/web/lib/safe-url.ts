// URLs read from the database end up in href attributes: anything but https (javascript:, data:, http) is dropped.
export function httpsUrlOrNull(url: string | null | undefined): string | null {
	if (!url) return null;
	try {
		return new URL(url).protocol === "https:" ? url : null;
	} catch {
		return null;
	}
}
