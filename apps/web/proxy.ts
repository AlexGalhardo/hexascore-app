import createMiddleware from "next-intl/middleware";
import { routing } from "./i18n/routing";

export default createMiddleware(routing);

export const config = {
	// The dot must be escaped: an unescaped `.*..*` matches every path, so no page was localized.
	matcher: ["/((?!api|_next|_vercel|.*\\..*).*)"],
};
