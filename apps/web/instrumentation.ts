export async function register(): Promise<void> {
	const { reportDisabledFeatures } = await import("./lib/env");
	reportDisabledFeatures();
}
