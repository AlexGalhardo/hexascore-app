export type Translate = (key: string, params?: Record<string, string | number>) => string;

// Relation and date fields need their own editors; the scalar text ones are editable here.
export const editable = (f: string): boolean =>
	f === "title" ||
	f === "cover" ||
	f.startsWith("translations.") ||
	f.startsWith("synopsis.") ||
	f.startsWith("external_ids.");

export function fieldLabel(t: Translate, f: string): string {
	if (f.startsWith("translations."))
		return t("community.fields.translationTitle", { locale: f.split(".")[1] });
	if (f.startsWith("synopsis."))
		return t("community.fields.synopsis", { locale: f.split(".")[1] });
	if (f.startsWith("external_ids."))
		return t("community.fields.externalId", { provider: f.split(".")[1] });
	return t(`community.fields.${f}`);
}

export const asText = (v: unknown): string => (typeof v === "string" ? v : "");
