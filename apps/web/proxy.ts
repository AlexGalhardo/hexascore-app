import createMiddleware from "next-intl/middleware";
import { routing } from "./i18n/routing";

export default createMiddleware(routing);

export const config = {
	// Skip api, Next internals and files (paths with a literal dot); `.*..*` would skip every path.
	matcher: ["/((?!api|_next|_vercel|.*\\..*).*)"],
};
