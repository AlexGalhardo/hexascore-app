import { z } from "zod";
import type { Locale } from "@/i18n/routing";
import { APP_NAME } from "./config";

export const LEGAL_UPDATED_AT = "2026-10-02";

const documentSchema = z.object({
	title: z.string(),
	description: z.string(),
	intro: z.string(),
	sections: z.array(z.object({ heading: z.string(), body: z.array(z.string()).min(1) })).min(1),
});

const legalSchema = z.object({ updatedLabel: z.string(), policy: documentSchema, terms: documentSchema });

export type LegalDocument = z.infer<typeof documentSchema> & { updatedLabel: string };

// Kept out of the i18n NAMESPACES on purpose: the long legal text would otherwise ship to the browser on every page.
export async function loadLegal(locale: Locale, doc: "policy" | "terms"): Promise<LegalDocument> {
	const raw: unknown = (await import(`../messages/${locale}/legal.json`)).default;
	const legal = legalSchema.parse(raw);
	// The brand is not final, so the text says {appName} and gets it from config.
	const brand = (text: string): string => text.replaceAll("{appName}", APP_NAME);
	const { title, description, intro, sections } = legal[doc];
	return {
		title,
		description,
		intro: brand(intro),
		sections: sections.map((s) => ({ heading: brand(s.heading), body: s.body.map(brand) })),
		updatedLabel: legal.updatedLabel,
	};
}
