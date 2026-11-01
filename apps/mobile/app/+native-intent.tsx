import { resolveDeepLink } from "../src/linking";

// Expo Router hook for incoming links (docs.expo.dev/router/advanced/native-intent): web App Links
// carry the locale prefix and web-only paths, which have no route here.
// shortcut: the `locale` part is dropped until the session-only locale store lands (P5.B.4.2 merge,
// row FIX-mobile-deeplink-locale).
export function redirectSystemPath({ path }: { path: string; initial: boolean }): string {
	return resolveDeepLink(path).path;
}
