import { createInertiaApp } from "@inertiajs/react";

import AppLayout from "./layouts/AppLayout";

// Server entry: the @inertiajs/vite plugin rewrites this top-level call into
// createServer + renderToString for `vite build --ssr`. No browser globals here.
createInertiaApp({
	layout: () => AppLayout,
});
