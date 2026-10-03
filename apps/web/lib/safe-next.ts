// Only same-origin relative paths: "//evil.com" and "/\evil.com" are protocol-relative in browsers.
export function safeNextPath(value: string | null | undefined, fallback = "/"): string {
	if (!value || !value.startsWith("/") || value.startsWith("//") || value.startsWith("/\\")) return fallback;
	return value;
}
