/** A plain preference cookie, one year, read server-side. */
export function setCookie(name: string, value: string): void {
	// biome-ignore lint/suspicious/noDocumentCookie: a plain preference cookie, read server-side
	document.cookie = `${name}=${value}; path=/; max-age=31536000; samesite=lax`;
}

/** True while the `dark` class is on <html> (the theme is set before paint by the app bootstrap). */
export function isDark(): boolean {
	return document.documentElement.classList.contains("dark");
}

/** Flips the theme and remembers it in the cookie (read server-side) and in localStorage. */
export function toggleTheme(): void {
	const dark = !isDark();
	document.documentElement.classList.toggle("dark", dark);
	const value = dark ? "dark" : "light";
	setCookie("theme", value);
	try {
		localStorage.setItem("theme", value);
	} catch {
		// Storage can be blocked: the cookie already holds the choice.
	}
}
