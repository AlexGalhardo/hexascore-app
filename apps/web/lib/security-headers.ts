// Security headers shared by `proxy.ts` (per-request CSP with nonce for pages) and `next.config.ts` (static headers).
// Kept free of Next imports so it runs in the proxy, the config and `bun test`.

/** Pages that mount Scalar (`components/api/api-reference.tsx`); only they get its extra allowances. */
const SCALAR_PAGE = /^(?:\/(?:pt-br|es))?\/(?:docs\/reference|profile\/api)\/?$/;

/** Origin of the production API host used by the spec `servers` (Scalar "Try it" on /docs/reference). */
const PUBLIC_API_ORIGIN = "https://api.hexascore.com";

export function createNonce(): string {
	const bytes = new Uint8Array(16);
	crypto.getRandomValues(bytes);
	return btoa(String.fromCharCode(...bytes));
}

export function isScalarPage(pathname: string): boolean {
	return SCALAR_PAGE.test(pathname);
}

export function pageCsp(nonce: string, pathname: string, isDev: boolean): string {
	const scalar = isScalarPage(pathname);
	const directives: Record<string, string[]> = {
		"default-src": ["'self'"],
		// 'strict-dynamic': scripts loaded by nonced scripts (Next chunks, BotID challenge, Vercel Analytics and
		// Speed Insights) are trusted; host allowlists and 'self' are ignored by CSP3 browsers.
		"script-src": ["'self'", `'nonce-${nonce}'`, "'strict-dynamic'", ...(isDev ? ["'unsafe-eval'"] : [])],
		// Inline style attributes (React `style={{}}`, Base UI positioning) and Scalar's runtime <style> tags cannot
		// carry a nonce; style injection without script execution is an accepted risk (docs/security.md).
		"style-src": ["'self'", "'unsafe-inline'"],
		// Covers come from several hosts (Open Library, Steam, GOG, Commons, Vercel Blob); QR codes are data: URIs.
		"img-src": ["'self'", "data:", "blob:", "https:"],
		"font-src": ["'self'", "data:"],
		"connect-src": ["'self'", ...(scalar ? [PUBLIC_API_ORIGIN] : []), ...(isDev ? ["ws:"] : [])],
		// 'self' for the BotID challenge iframe; trailers only from the privacy-enhanced YouTube host.
		"frame-src": ["'self'", "https://www.youtube-nocookie.com"],
		"worker-src": ["'self'", "blob:"],
		"object-src": ["'none'"],
		"base-uri": ["'self'"],
		// Stripe Checkout/Portal and Google OAuth are reached by navigation after a Server Action/fetch; listed so a
		// no-JS form post that is redirected there is not blocked.
		"form-action": [
			"'self'",
			"https://checkout.stripe.com",
			"https://billing.stripe.com",
			"https://accounts.google.com",
		],
		"frame-ancestors": ["'none'"],
		"manifest-src": ["'self'"],
	};
	const parts = Object.entries(directives).map(([key, values]) => `${key} ${values.join(" ")}`);
	if (!isDev) parts.push("upgrade-insecure-requests");
	return parts.join("; ");
}

/** JSON-only responses (`/api/*`): nothing may be loaded, rendered or framed from them. */
export const API_CSP = "default-src 'none'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'";

export const PERMISSIONS_POLICY = [
	"accelerometer=()",
	"autoplay=()",
	"camera=()",
	"display-capture=()",
	"geolocation=()",
	"gyroscope=()",
	"magnetometer=()",
	"microphone=()",
	"midi=()",
	"payment=()",
	"usb=()",
	"browsing-topics=()",
	// Passkeys (WebAuthn) run in the top-level document only.
	"publickey-credentials-get=(self)",
	"publickey-credentials-create=(self)",
	'fullscreen=(self "https://www.youtube-nocookie.com")',
	'picture-in-picture=(self "https://www.youtube-nocookie.com")',
	'encrypted-media=(self "https://www.youtube-nocookie.com")',
].join(", ");

/** Headers for every response. CSP and framing are per route (proxy for pages, `/api/*` below, BotID sets its own). */
export const BASE_HEADERS: { key: string; value: string }[] = [
	{ key: "Strict-Transport-Security", value: "max-age=63072000; includeSubDomains; preload" },
	{ key: "X-Content-Type-Options", value: "nosniff" },
	{ key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
	{ key: "Permissions-Policy", value: PERMISSIONS_POLICY },
	{ key: "Cross-Origin-Opener-Policy", value: "same-origin" },
	{ key: "X-DNS-Prefetch-Control", value: "off" },
];

export const API_HEADERS: { key: string; value: string }[] = [
	{ key: "Content-Security-Policy", value: API_CSP },
	{ key: "X-Frame-Options", value: "DENY" },
	{ key: "Cross-Origin-Resource-Policy", value: "same-site" },
];
