import { initBotId } from "botid/client/core";

// Must list every request the server bot-checks (lib/bot.ts). Server Actions POST to the page URL, per locale.
initBotId({
	protect: [
		{ path: "/api/auth/sign-up/email", method: "POST" },
		{ path: "/api/auth/sign-in/email", method: "POST" },
		{ path: "/api/auth/sign-in/username", method: "POST" },
		{ path: "/contact", method: "POST" },
		{ path: "/pt-br/contact", method: "POST" },
		{ path: "/es/contact", method: "POST" },
	],
});
