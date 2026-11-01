import { usePage } from "@inertiajs/react";
import { type ReactNode, useCallback, useEffect, useRef, useState } from "react";
import type { SharedProps } from "../types";

declare global {
	interface Window {
		turnstile?: {
			render(el: HTMLElement, opts: Record<string, unknown>): string;
			remove(id: string): void;
		};
	}
}

// The same URL as vukem's challenge.TurnstileScriptURL, which app.html renders on full loads of the protected pages.
export const TURNSTILE_SCRIPT = "https://challenges.cloudflare.com/turnstile/v0/api.js";

// The request header vukem's challenge middleware reads (challenge.Header).
export const CHALLENGE_TOKEN_HEADER = "X-Challenge-Token";

let loading: Promise<void> | null = null;

// loadTurnstile resolves once window.turnstile exists. After a client-side visit the head script is absent, so it
// is added here; 'strict-dynamic' trusts a script added by the nonce'd app bundle.
export function loadTurnstile(): Promise<void> {
	if (window.turnstile) return Promise.resolve();
	loading ??= new Promise<void>((resolve, reject) => {
		const existing = document.querySelector<HTMLScriptElement>(
			`script[src^="${TURNSTILE_SCRIPT}"]`,
		);
		const script = existing ?? document.createElement("script");
		script.addEventListener("load", () => resolve());
		script.addEventListener("error", () => {
			loading = null;
			reject(new Error("turnstile script failed"));
		});
		if (!existing) {
			script.src = TURNSTILE_SCRIPT;
			script.async = true;
			const nonce = document.querySelector<HTMLMetaElement>(
				'meta[property="csp-nonce"]',
			)?.nonce;
			if (nonce) script.nonce = nonce;
			document.head.appendChild(script);
		}
	});
	return loading;
}

interface Turnstile {
	/** The widget and the challenge error, or nothing while Turnstile is off. */
	render: (error?: string) => ReactNode;
	/** Headers to merge into the visit: the token, once solved. */
	headers: () => Record<string, string>;
	/** Shows a fresh widget; a token is single use, so call it after every submit. */
	reset: () => void;
}

// useTurnstile renders the Cloudflare Turnstile widget for action ("login", "register", "contact") when the server
// shares a site key (https://vukem.dev/docs/security/bot-challenge, "Inertia and React").
export function useTurnstile(action: string): Turnstile {
	const siteKey = usePage<SharedProps>().props.turnstileSiteKey ?? "";
	const [token, setToken] = useState("");
	const [round, setRound] = useState(0);
	const ref = useRef<HTMLDivElement>(null);

	// biome-ignore lint/correctness/useExhaustiveDependencies: round re-renders a fresh widget after reset
	useEffect(() => {
		if (!siteKey || !ref.current) return;
		const el = ref.current;
		let id: string | undefined;
		let gone = false;
		loadTurnstile()
			.then(() => {
				if (gone || !window.turnstile) return;
				id = window.turnstile.render(el, {
					sitekey: siteKey,
					action,
					callback: (t: string) => setToken(t),
					"expired-callback": () => setToken(""),
					"error-callback": () => setToken(""),
				});
			})
			.catch(() => setToken(""));
		return () => {
			gone = true;
			if (id) window.turnstile?.remove(id);
		};
	}, [siteKey, action, round]);

	const reset = useCallback(() => {
		setToken("");
		setRound((n) => n + 1);
	}, []);

	return {
		render: (error) =>
			siteKey ? (
				<div className="space-y-1">
					<div ref={ref} className="min-h-[65px]" data-testid="turnstile" />
					{error && (
						<p role="alert" className="text-sm text-danger">
							{error}
						</p>
					)}
				</div>
			) : null,
		headers: () => (token ? { [CHALLENGE_TOKEN_HEADER]: token } : {}),
		reset,
	};
}
