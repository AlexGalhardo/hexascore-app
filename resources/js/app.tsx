import { createInertiaApp } from "@inertiajs/react";

import AppLayout from "./layouts/AppLayout";
import { withQueryClient } from "./lib/queryClient";

// Dark mode follows the "dark" class on <html>: the saved choice
// (localStorage.theme = "dark" | "light") or the system preference. A theme
// toggle flips the class and saves it.
try {
	// The server sets data-theme from the theme cookie (explicit choice wins).
	const theme = document.documentElement.dataset.theme ?? localStorage.getItem("theme");
	const dark = theme ? theme === "dark" : matchMedia("(prefers-color-scheme: dark)").matches;
	document.documentElement.classList.toggle("dark", dark);
} catch {
	// Storage can be blocked (private mode): keep the default theme.
}

// The @inertiajs/vite plugin resolves pages from ./pages and mounts the app.
// The nonce lets the progress bar styles through the CSP.
// It mounts on #app, the default inertia.Config.RootID; when you change
// RootID, pass the same value here as `id`.
// The HTML loader sets <html inert> until this promise settles. Script onload
// is too early: createInertiaApp hydrates asynchronously, and a click in that
// gap submits the server form.
try {
	void Promise.resolve(
		createInertiaApp({
			layout: () => AppLayout,
			withApp: withQueryClient,
			nonce: document.querySelector<HTMLMetaElement>('meta[property="csp-nonce"]')?.nonce,
			progress: { color: "#4f46e5" },
		}),
	).finally(() => {
		document.documentElement.inert = false;
	});
} catch {
	document.documentElement.inert = false;
}
