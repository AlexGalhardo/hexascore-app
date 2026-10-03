"use client";

import dynamic from "next/dynamic";
import { hasLocale, useLocale } from "next-intl";
import { useEffect, useState } from "react";
import { routing } from "@/i18n/routing";
import { openApiPath } from "@/lib/api/spec-path";
import type { Theme } from "@/lib/preferences";
// The package JS does not import its stylesheet; without it Scalar renders unstyled.
import "@scalar/api-reference-react/style.css";

// Scalar is large and browser-only: load it on demand, never during SSR.
const ApiReferenceReact = dynamic(() => import("@scalar/api-reference-react").then((m) => m.ApiReferenceReact), {
	ssr: false,
});

// Scalar ships built-in UI packs; our pt-BR maps to its `pt`. The spec text comes from openapi.<locale>.json.
const SCALAR_LOCALE = { en: "en", "pt-BR": "pt", es: "es" } as const;

function currentTheme(): Theme {
	return document.documentElement.dataset.theme === "dark" ? "dark" : "light";
}

/** Follows the app's `data-theme` (toggled by ThemeToggle) instead of Scalar's own switch. */
function useAppTheme(): Theme | null {
	const [theme, setTheme] = useState<Theme | null>(null);
	useEffect(() => {
		setTheme(currentTheme());
		const observer = new MutationObserver(() => setTheme(currentTheme()));
		observer.observe(document.documentElement, { attributes: true, attributeFilter: ["data-theme"] });
		return () => observer.disconnect();
	}, []);
	return theme;
}

type Props = {
	loadingLabel: string;
	/** Prefills "Try it". Lives only in React state: Scalar's auth persistence is off, nothing reaches storage. */
	token?: string;
	/** Overrides the spec servers so "Try it" calls this deployment. */
	serverUrl?: string;
};

export function ApiReference({ loadingLabel, token, serverUrl }: Props) {
	const theme = useAppTheme();
	const requested = useLocale();
	const siteLocale = hasLocale(routing.locales, requested) ? requested : routing.defaultLocale;
	const locale = SCALAR_LOCALE[siteLocale];
	const specUrl = openApiPath(siteLocale);
	if (!theme) {
		return (
			<p role="status" className="text-sm text-muted">
				{loadingLabel}
			</p>
		);
	}
	return (
		// data-plain: Scalar styles its own links; the global button-like link hover must not reach them.
		// lang: the docs <article> is lang="en" (guides), but the reference text follows the site locale.
		<div data-plain lang={siteLocale}>
			<ApiReferenceReact
				// Remount on theme/locale/token change: Scalar reads these options once.
				key={`${theme}:${locale}:${specUrl}:${token ?? ""}`}
				configuration={{
					url: specUrl,
					forceDarkModeState: theme,
					hideDarkModeToggle: true,
					withDefaultFonts: false,
					telemetry: false,
					localization: { locale },
					persistAuth: false,
					showDeveloperTools: "never",
					// No third-party AI/MCP features: requests carry the user's key on /profile/api.
					agent: { disabled: true },
					mcp: { disabled: true },
					defaultHttpClient: { targetKey: "shell", clientKey: "curl" },
					...(serverUrl ? { servers: [{ url: serverUrl }] } : {}),
					...(token
						? {
								authentication: {
									preferredSecurityScheme: "bearerAuth",
									securitySchemes: { bearerAuth: { token } },
								},
							}
						: {}),
				}}
			/>
		</div>
	);
}
