import { flag } from "flags/next";

// Kill switches, evaluated on the server only. Default on; turn one off by setting its env var to `off`
// (Vercel project settings → redeploy) or, with FLAGS_SECRET set, per browser through the Vercel Toolbar overrides.
// See docs/process.md → "Feature flags".

const OFF = new Set(["off", "false", "0"]);

export function envFlag(name: string): boolean {
	return !OFF.has((process.env[name] ?? "").trim().toLowerCase());
}

const options = [
	{ value: true, label: "On" },
	{ value: false, label: "Off" },
];

export const pixCheckoutFlag = flag<boolean>({
	key: "pix-checkout",
	description: "PIX (AbacatePay) on /profile/api/checkout and in createPixCharge. Env: FLAG_PIX_CHECKOUT",
	defaultValue: true,
	options,
	decide: () => envFlag("FLAG_PIX_CHECKOUT"),
});

// Unlike the kill switches above, off by default: the owner sets FLAG_SEED_DATA=on while synthetic seed users live
// in production, and /admin shows a reminder to remove them (`seed.ts --purge`, docs/testing.md).
export const seedDataFlag = flag<boolean>({
	key: "seed-data",
	description: "Synthetic seed users are in this database; /admin reminds to purge them. Env: FLAG_SEED_DATA=on",
	defaultValue: false,
	options,
	decide: () => (process.env.FLAG_SEED_DATA ?? "").trim().toLowerCase() === "on",
});

export const publicApiSignupFlag = flag<boolean>({
	key: "public-api-signup",
	description: "Creating new public API keys on /profile/api. Env: FLAG_PUBLIC_API_SIGNUP",
	defaultValue: true,
	options,
	decide: () => envFlag("FLAG_PUBLIC_API_SIGNUP"),
});
