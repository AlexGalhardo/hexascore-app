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

export type DateValue = { value: string; precision: string };

export type RelationOp = { id: string; role?: string };

export type Change =
	| { before: DateValue | string | null; after: DateValue | string }
	| { add: RelationOp[]; remove: RelationOp[] };

export const isDate = (f: string): boolean => f === "release_date";

// Roles per relation field, from the SDD 01 §5.2 table; fields without roles carry plain ids.
export const relationRoles: Record<string, string[]> = {
	genres: [],
	franchise: [],
	platforms: [],
	creators: ["director", "writer", "cast", "creator", "author"],
	organizations: ["developer", "publisher", "production_company", "network"],
};

export const isRelation = (f: string): boolean => f in relationRoles;

export const asDate = (v: unknown): DateValue | null => {
	const d = v as Partial<DateValue> | null;
	return d && typeof d.value === "string"
		? { value: d.value, precision: d.precision ?? "day" }
		: null;
};

export const asOps = (v: unknown): RelationOp[] => (Array.isArray(v) ? (v as RelationOp[]) : []);

// One id per line.
export const parseIds = (text: string, role?: string): RelationOp[] =>
	text
		.split("\n")
		.map((s) => s.trim())
		.filter(Boolean)
		.map((id) => (role ? { id, role } : { id }));

export const idsText = (ops: RelationOp[]): string => ops.map((o) => o.id).join("\n");
