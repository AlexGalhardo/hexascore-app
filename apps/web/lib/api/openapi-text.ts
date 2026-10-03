import type { Locale } from "@/i18n/routing";
import en from "@/messages/en/openapi.json";
import es from "@/messages/es/openapi.json";
import ptBR from "@/messages/pt-BR/openapi.json";

// Human-readable OpenAPI text lives in messages/<locale>/openapi.json. It is not a next-intl namespace: only the
// generator (lib/api/openapi.ts) reads it, so it never ships in page payloads.

type Leaves<T> = {
	[K in keyof T & string]: T[K] extends string ? K : `${K}.${Leaves<T[K]>}`;
}[keyof T & string];

export type OpenApiKey = Leaves<typeof en>;
export type TextParams = Record<string, string | number>;
export type OperationId = keyof (typeof en)["operations"];

// Typing every locale as `typeof en` also makes tsc reject a pt-BR/es file whose keys drift from en.
const MESSAGES: Record<Locale, typeof en> = { en, "pt-BR": ptBR, es };

/** `{name}` placeholders are filled from `params`; braces without a matching param stay literal. */
export function openApiText(locale: Locale, key: OpenApiKey, params: TextParams = {}): string {
	const value = key
		.split(".")
		.reduce<unknown>(
			(node, part) =>
				typeof node === "object" && node !== null ? (node as Record<string, unknown>)[part] : undefined,
			MESSAGES[locale],
		);
	if (typeof value !== "string") throw new Error(`Missing OpenAPI text "${key}" for ${locale}`);
	return value.replace(/\{(\w+)\}/g, (match, name: string) => (name in params ? String(params[name]) : match));
}

// zod `.meta({ description })` is fixed at module load, so schemas carry the English text and this map remembers
// which key produced it; `localizeDescriptions` swaps it per locale after `z.toJSONSchema`.
const described = new Map<string, { key: OpenApiKey; params: TextParams }>();

/** English description for a zod `.meta()`, translatable later by `localizeDescriptions`. */
export function apiDoc(key: OpenApiKey, params: TextParams = {}): string {
	const text = openApiText("en", key, params);
	const seen = described.get(text);
	if (seen && seen.key !== key) throw new Error(`OpenAPI text "${text}" comes from both ${seen.key} and ${key}`);
	described.set(text, { key, params });
	return text;
}

/** Rewrites every `description` in a JSON Schema tree (in place); one not made by `apiDoc` is an error. */
export function localizeDescriptions<T>(json: T, locale: Locale): T {
	const visit = (node: unknown): void => {
		if (Array.isArray(node)) {
			for (const child of node) visit(child);
			return;
		}
		if (typeof node !== "object" || node === null) return;
		const record = node as Record<string, unknown>;
		for (const [key, value] of Object.entries(record)) {
			if (key === "description" && typeof value === "string") {
				const entry = described.get(value);
				if (!entry) throw new Error(`OpenAPI description without a message key: "${value}" (use apiDoc)`);
				record[key] = openApiText(locale, entry.key, entry.params);
			} else {
				visit(value);
			}
		}
	};
	visit(json);
	return json;
}
