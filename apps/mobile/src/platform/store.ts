export const ANDROID_PACKAGE = "app.hexascore";

export function storeListingUrl(packageName: string = ANDROID_PACKAGE): string {
	return `market://details?id=${packageName}`;
}
