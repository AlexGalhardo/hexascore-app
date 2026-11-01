import { appLocale } from "../src/auth/runtime";
import { resolveDeepLink } from "../src/linking";

// Expo Router hook for incoming links (docs.expo.dev/router/advanced/native-intent): web App Links
// carry the locale prefix and web-only paths, which have no route here. The prefix sets the
// session locale only (ROADMAP P5.B.5.1): the user's stored choice survives the next launch.
export function redirectSystemPath({ path }: { path: string; initial: boolean }): string {
	const link = resolveDeepLink(path);
	if (link.locale) appLocale.setSession(link.locale);
	return link.path;
}
