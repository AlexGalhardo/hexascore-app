import { createInertiaApp } from "@inertiajs/react";

import AdminLayout from "./layouts/AdminLayout";
import AppLayout from "./layouts/AppLayout";

// Server entry: the @inertiajs/vite plugin rewrites this top-level call into
// createServer + renderToString for `vite build --ssr`. No browser globals here.
createInertiaApp({
	layout: (name) => (name.startsWith("Admin/") ? AdminLayout : AppLayout),
});
