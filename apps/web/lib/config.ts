export const APP_NAME = process.env.NEXT_PUBLIC_APP_NAME || "HexaScore";
export const APP_URL =
	process.env.NEXT_PUBLIC_APP_URL ||
	(process.env.VERCEL_PROJECT_PRODUCTION_URL
		? `https://${process.env.VERCEL_PROJECT_PRODUCTION_URL}`
		: "http://localhost:3000");
