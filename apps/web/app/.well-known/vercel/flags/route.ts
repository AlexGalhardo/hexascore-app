import { createFlagsDiscoveryEndpoint, getProviderData } from "flags/next";
import { pixCheckoutFlag, publicApiSignupFlag, seedDataFlag } from "@/lib/flags";

// Vercel Toolbar flag discovery; the helper rejects requests whose `Authorization` is not signed with FLAGS_SECRET.
export const GET = createFlagsDiscoveryEndpoint(() =>
	getProviderData({ pixCheckoutFlag, publicApiSignupFlag, seedDataFlag }),
);
