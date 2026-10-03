import { z } from "zod";
import type { Locale } from "@/i18n/routing";
import { CATEGORY_KEYS } from "@/lib/categories";
import type { ItemType } from "@/lib/db/schema/catalog";
import { SECTION_PATHS } from "@/lib/sections";
import { localizeDescriptions, type OpenApiKey, type OperationId, openApiText, type TextParams } from "./openapi-text";
import { PLAN_LIMITS } from "./plans";
import { PROBLEM_CODES, PROBLEM_STATUS, type ProblemCode } from "./problem";
import {
	detailQuerySchema,
	genreListSchema,
	itemDetailSchema,
	itemListSchema,
	itemSummarySchema,
	listQuerySchema,
	lookupQuerySchema,
	problemSchema,
	SEARCH_MAX,
	searchQuerySchema,
	searchResultSchema,
	synopsisSchema,
	timeToBeatSchema,
	usageSchema,
} from "./schemas";

// OpenAPI 3.1 built from the same zod schemas the route handlers parse their responses with, once per locale:
// only description/summary text differs (messages/<locale>/openapi.json). `bun run openapi` writes
// public/openapi.json (en) + openapi.<locale>.json; `bun run openapi:check` fails CI when any is stale.

type Json = Record<string, unknown>;
type T = (key: OpenApiKey, params?: TextParams) => string;

// The public docs show this host; until the custom domain exists, `/api/v1` on the deployment is the same API.
export const API_BASE_URL = "https://api.hexascore.app/v1";
const DOCS_URL = "https://hexascore.app/docs";

const components = z.registry<{ id: string }>();
components.add(itemSummarySchema, { id: "ItemSummary" });
components.add(itemDetailSchema, { id: "ItemDetail" });
components.add(synopsisSchema, { id: "Synopsis" });
components.add(timeToBeatSchema, { id: "TimeToBeat" });
components.add(itemListSchema, { id: "ItemList" });
components.add(searchResultSchema, { id: "SearchResult" });
components.add(genreListSchema, { id: "GenreList" });
components.add(usageSchema, { id: "Usage" });
components.add(problemSchema, { id: "Problem" });

const ref = (id: string): Json => ({ $ref: `#/components/schemas/${id}` });

function clean(schema: Json): Json {
	const { $schema: _s, $id: _i, ...rest } = schema;
	return rest;
}

function queryParameters(schema: z.ZodObject, locale: Locale): Json[] {
	return Object.entries(schema.shape).map(([name, field]) => {
		const json = localizeDescriptions(clean(z.toJSONSchema(field as z.ZodType, { io: "input" }) as Json), locale);
		const { description, examples, ...rest } = json;
		return {
			name,
			in: "query",
			required: !(field as z.ZodType).safeParse(undefined).success,
			...(description ? { description } : {}),
			...(Array.isArray(examples) && examples.length ? { example: examples[0] } : {}),
			schema: rest,
		};
	});
}

function rateLimitHeaders(t: T): Json {
	return {
		"RateLimit-Limit": { description: t("headers.rateLimitLimit"), schema: { type: "integer" } },
		"RateLimit-Remaining": { description: t("headers.rateLimitRemaining"), schema: { type: "integer" } },
		"RateLimit-Reset": { description: t("headers.rateLimitReset"), schema: { type: "integer" } },
		"X-Quota-Remaining": { description: t("headers.quotaRemaining"), schema: { type: "integer" } },
		"X-Request-Id": { description: t("headers.requestId"), schema: { type: "string" } },
	};
}

const responseName = (code: ProblemCode): string =>
	code.replace(/(^|_)([a-z])/g, (_m, _s, c: string) => c.toUpperCase());

function problemResponse(code: ProblemCode, t: T): Json {
	const [status, title] = PROBLEM_STATUS[code];
	return {
		description: t(`errors.${code}`),
		headers: {
			"X-Request-Id": rateLimitHeaders(t)["X-Request-Id"],
			...(status === 429 || status === 503
				? { "Retry-After": { description: t("headers.retryAfter"), schema: { type: "integer" } } }
				: {}),
		},
		content: {
			"application/problem+json": {
				schema: ref("Problem"),
				example: {
					type: `${DOCS_URL}/errors#${code}`,
					title,
					status,
					// The example mirrors a real response body, and the API answers in English.
					detail: openApiText("en", `errors.${code}`),
					code,
					request_id: "5f0c8d3e-4b7a-4f43-9a51-2a8c4f3f9b10",
				},
			},
		},
	};
}

function operation(
	t: T,
	locale: Locale,
	options: {
		id: OperationId;
		tag: string;
		params?: TextParams;
		query: z.ZodObject;
		pathParams?: Json[];
		schema: string;
		errors: ProblemCode[];
	},
): Json {
	return {
		operationId: options.id,
		tags: [options.tag],
		summary: t(`operations.${options.id}.summary`),
		description: t(`operations.${options.id}.description`, options.params),
		parameters: [...(options.pathParams ?? []), ...queryParameters(options.query, locale)],
		responses: {
			"200": {
				description: "OK",
				headers: rateLimitHeaders(t),
				content: { "application/json": { schema: ref(options.schema) } },
			},
			...Object.fromEntries(
				options.errors.map((code) => [
					String(PROBLEM_STATUS[code][0]),
					{ $ref: `#/components/responses/${responseName(code)}` },
				]),
			),
		},
	};
}

const COMMON_ERRORS: ProblemCode[] = [
	"unauthorized",
	"payment_required",
	"forbidden",
	"rate_limited",
	"internal_error",
	"service_unavailable",
];

const NAMES = {
	game: { one: "Game", many: "Games", tag: "Games" },
	movie: { one: "Movie", many: "Movies", tag: "Movies" },
	series: { one: "Series", many: "Series", tag: "Series" },
	book: { one: "Book", many: "Books", tag: "Books" },
} as const satisfies Record<ItemType, { one: string; many: string; tag: string }>;

function typePaths(type: ItemType, t: T, locale: Locale): Record<string, Json> {
	const section = SECTION_PATHS[type];
	const { one, many, tag } = NAMES[type];
	const categories = CATEGORY_KEYS[type].map((k) => `\`${k}\``).join(", ");
	return {
		[`/${section}`]: {
			get: operation(t, locale, {
				id: `list${many}`,
				tag,
				params: { categories },
				query: listQuerySchema(type, PLAN_LIMITS.business.maxPageSize),
				schema: "ItemList",
				errors: ["bad_request", ...COMMON_ERRORS],
			}),
		},
		[`/${section}/{id}`]: {
			get: operation(t, locale, {
				id: `get${one}`,
				tag,
				params: { categories },
				query: detailQuerySchema,
				pathParams: [
					{
						name: "id",
						in: "path",
						required: true,
						description: t("params.itemId"),
						schema: { type: "string", format: "uuid" },
					},
				],
				schema: "ItemDetail",
				errors: ["bad_request", "not_found", ...COMMON_ERRORS],
			}),
		},
		[`/${section}/genres`]: {
			get: operation(t, locale, {
				id: `list${one}Genres`,
				tag,
				query: z.strictObject({}),
				schema: "GenreList",
				errors: COMMON_ERRORS,
			}),
		},
	};
}

export function buildOpenApiDocument(locale: Locale = "en"): Json {
	const t: T = (key, params) => openApiText(locale, key, params);
	const { schemas } = z.toJSONSchema(components, { uri: (id) => `#/components/schemas/${id}` }) as {
		schemas: Record<string, Json>;
	};
	localizeDescriptions(schemas, locale);
	const types = Object.keys(SECTION_PATHS) as ItemType[];

	return {
		openapi: "3.1.0",
		info: {
			title: "HexaScore API",
			version: "1.0.0",
			summary: t("info.summary"),
			description: t("info.description", {
				brand: "HexaScore",
				freeQuota: PLAN_LIMITS.free.quota,
				freePageSize: PLAN_LIMITS.free.maxPageSize,
				freeRps: PLAN_LIMITS.free.requestsPerSecond,
				paidPageSize: PLAN_LIMITS.pro.maxPageSize,
				docsUrl: DOCS_URL,
			}),
			contact: { name: "HexaScore", url: DOCS_URL },
			license: { name: "Proprietary", url: "https://hexascore.app/terms" },
		},
		servers: [
			{ url: API_BASE_URL, description: t("servers.production") },
			{ url: "/api/v1", description: t("servers.deployment") },
		],
		security: [{ bearerAuth: [] }],
		tags: [
			...types.map((type) => ({ name: NAMES[type].tag, description: t(`tags.${type}`) })),
			{ name: "Search", description: t("tags.search") },
			{ name: "Account", description: t("tags.account") },
		],
		externalDocs: { description: t("externalDocs"), url: DOCS_URL },
		paths: {
			...Object.assign({}, ...types.map((type) => typePaths(type, t, locale))),
			"/search": {
				get: operation(t, locale, {
					id: "searchItems",
					tag: "Search",
					params: { max: SEARCH_MAX },
					query: searchQuerySchema(PLAN_LIMITS.business.maxPageSize),
					schema: "SearchResult",
					errors: ["bad_request", ...COMMON_ERRORS],
				}),
			},
			"/lookup": {
				get: operation(t, locale, {
					id: "lookupItem",
					tag: "Search",
					query: lookupQuerySchema,
					schema: "ItemDetail",
					errors: ["bad_request", "not_found", ...COMMON_ERRORS],
				}),
			},
			"/me/usage": {
				get: operation(t, locale, {
					id: "getUsage",
					tag: "Account",
					query: z.strictObject({}),
					schema: "Usage",
					errors: COMMON_ERRORS,
				}),
			},
		},
		components: {
			schemas: Object.fromEntries(Object.entries(schemas).map(([id, s]) => [id, clean(s)])),
			responses: Object.fromEntries(PROBLEM_CODES.map((code) => [responseName(code), problemResponse(code, t)])),
			securitySchemes: {
				bearerAuth: {
					type: "http",
					scheme: "bearer",
					bearerFormat: "hx_live_<32 chars>",
					description: t("security.bearerAuth"),
				},
			},
		},
	};
}

export function serializeOpenApi(locale: Locale = "en"): string {
	return `${JSON.stringify(buildOpenApiDocument(locale), null, 2)}\n`;
}
