"use client";

import dynamic from "next/dynamic";
import { useEffect, useState } from "react";
import type { Theme } from "@/lib/preferences";

// Scalar is large and browser-only: load it on demand, never during SSR.
const ApiReferenceReact = dynamic(() => import("@scalar/api-reference-react").then((m) => m.ApiReferenceReact), {
	ssr: false,
});

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
	if (!theme) {
		return (
			<p role="status" className="text-sm text-muted">
				{loadingLabel}
			</p>
		);
	}
	return (
		<ApiReferenceReact
			// Remount on theme/token change: Scalar reads these options once.
			key={`${theme}:${token ?? ""}`}
			configuration={{
				url: "/openapi.json",
				forceDarkModeState: theme,
				hideDarkModeToggle: true,
				withDefaultFonts: false,
				telemetry: false,
				persistAuth: false,
				showDeveloperTools: "never",
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
	);
}
