// Only same-origin relative paths: "//evil.com" and "/\evil.com" are protocol-relative in browsers.
export function safeNextPath(value: string | null | undefined, fallback = "/"): string {
	return value && /^\/(?![/\\])/.test(value) ? value : fallback;
}
