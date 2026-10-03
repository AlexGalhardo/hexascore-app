import { z } from "zod";
import { CATEGORY_KEYS } from "@/lib/categories";
import type { ItemType } from "@/lib/db/schema/catalog";
import { SECTION_PATHS } from "@/lib/sections";
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
	usageSchema,
} from "./schemas";

// OpenAPI 3.1 built from the same zod schemas the route handlers parse their responses with.
// `bun run openapi` writes public/openapi.json; `bun run openapi:check` fails CI when it is stale.

type Json = Record<string, unknown>;

// The public docs show this host; until the custom domain exists, `/api/v1` on the deployment is the same API.
export const API_BASE_URL = "https://api.hexascore.com/v1";
const DOCS_URL = "https://hexascore.com/docs";

const components = z.registry<{ id: string }>();
components.add(itemSummarySchema, { id: "ItemSummary" });
components.add(itemDetailSchema, { id: "ItemDetail" });
components.add(synopsisSchema, { id: "Synopsis" });
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

function queryParameters(schema: z.ZodObject): Json[] {
	return Object.entries(schema.shape).map(([name, field]) => {
		const json = clean(z.toJSONSchema(field as z.ZodType, { io: "input" }) as Json);
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

const rateLimitHeaders: Json = {
	"RateLimit-Limit": { description: "Requests in the current quota period.", schema: { type: "integer" } },
	"RateLimit-Remaining": { description: "Requests left in the period.", schema: { type: "integer" } },
	"RateLimit-Reset": {
		description: "Seconds until the quota resets (`0` for the lifetime Free quota).",
		schema: { type: "integer" },
	},
	"X-Quota-Remaining": { description: "Same as `RateLimit-Remaining`.", schema: { type: "integer" } },
	"X-Request-Id": { description: "Id of this request; quote it to support.", schema: { type: "string" } },
};

const ERROR_RESPONSES: Record<ProblemCode, string> = {
	bad_request: "Invalid query parameter, cursor, or a `limit` above your plan's page size.",
	unauthorized: "Missing, malformed, unknown or revoked API key.",
	payment_required: "Quota (and credits, on paid plans) exhausted. Upgrade or buy credits.",
	forbidden: "The account that owns the key is suspended.",
	not_found: "No such item, type or route.",
	rate_limited: "Too many requests per second. Wait `Retry-After` seconds.",
	internal_error: "Unexpected error. The request was not counted against your quota.",
	service_unavailable: "Temporarily unavailable. Retry after `Retry-After` seconds.",
};

const responseName = (code: ProblemCode): string =>
	code.replace(/(^|_)([a-z])/g, (_m, _s, c: string) => c.toUpperCase());

function problemResponse(code: ProblemCode): Json {
	const [status, title] = PROBLEM_STATUS[code];
	return {
		description: ERROR_RESPONSES[code],
		headers: {
			"X-Request-Id": rateLimitHeaders["X-Request-Id"],
			...(status === 429 || status === 503
				? { "Retry-After": { description: "Seconds to wait.", schema: { type: "integer" } } }
				: {}),
		},
		content: {
			"application/problem+json": {
				schema: ref("Problem"),
				example: {
					type: `${DOCS_URL}/errors#${code}`,
					title,
					status,
					detail: ERROR_RESPONSES[code],
					code,
					request_id: "5f0c8d3e-4b7a-4f43-9a51-2a8c4f3f9b10",
				},
			},
		},
	};
}

function operation(options: {
	id: string;
	tag: string;
	summary: string;
	description: string;
	query: z.ZodObject;
	pathParams?: Json[];
	schema: string;
	errors: ProblemCode[];
}): Json {
	return {
		operationId: options.id,
		tags: [options.tag],
		summary: options.summary,
		description: options.description,
		parameters: [...(options.pathParams ?? []), ...queryParameters(options.query)],
		responses: {
			"200": {
				description: "OK",
				headers: rateLimitHeaders,
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

const NAMES: Record<ItemType, { one: string; many: string; tag: string }> = {
	game: { one: "Game", many: "Games", tag: "Games" },
	movie: { one: "Movie", many: "Movies", tag: "Movies" },
	series: { one: "Series", many: "Series", tag: "Series" },
	book: { one: "Book", many: "Books", tag: "Books" },
};

function typePaths(type: ItemType): Record<string, Json> {
	const section = SECTION_PATHS[type];
	const { one, many, tag } = NAMES[type];
	const label = section === "series" ? "TV series" : section;
	const categories = CATEGORY_KEYS[type].map((k) => `\`${k}\``).join(", ");
	return {
		[`/${section}`]: {
			get: operation({
				id: `list${many}`,
				tag,
				summary: `List ${label}`,
				description: `Pages through ${label} with filters. Categories for this type: ${categories} — filter with \`min_<category>\`. Pass \`next_cursor\` back as \`cursor\` until it is \`null\`.`,
				query: listQuerySchema(type, PLAN_LIMITS.business.maxPageSize),
				schema: "ItemList",
				errors: ["bad_request", ...COMMON_ERRORS],
			}),
		},
		[`/${section}/{id}`]: {
			get: operation({
				id: `get${one}`,
				tag,
				summary: `Get a ${label === "TV series" ? "TV series" : one.toLowerCase()}`,
				description: `Full record: metadata, community score and per-category averages (${categories}), external ids, trailers and the licensed synopsis. An id of another type returns 404.`,
				query: detailQuerySchema,
				pathParams: [
					{
						name: "id",
						in: "path",
						required: true,
						description: "Item id (UUID).",
						schema: { type: "string", format: "uuid" },
					},
				],
				schema: "ItemDetail",
				errors: ["bad_request", "not_found", ...COMMON_ERRORS],
			}),
		},
		[`/${section}/genres`]: {
			get: operation({
				id: `list${one}Genres`,
				tag,
				summary: `List ${label} genres`,
				description: "Every genre used by this type with its item count. Use `slug` as the `genre` filter.",
				query: z.strictObject({}),
				schema: "GenreList",
				errors: COMMON_ERRORS,
			}),
		},
	};
}

export function buildOpenApiDocument(): Json {
	const { schemas } = z.toJSONSchema(components, { uri: (id) => `#/components/schemas/${id}` }) as {
		schemas: Record<string, Json>;
	};
	const types = Object.keys(SECTION_PATHS) as ItemType[];

	return {
		openapi: "3.1.0",
		info: {
			title: "Hexascore API",
			version: "1.0.0",
			summary: "Community scores for games, movies, TV series and books.",
			description: [
				"Read-only REST API over the Hexascore catalog: every item carries a 0–100 community score and six per-category averages.",
				"",
				"- **Auth**: `Authorization: Bearer hx_live_…` — create a key at `/profile/api`.",
				"- **Format**: JSON, `snake_case` fields, ISO 8601 dates. Errors are RFC 9457 `application/problem+json` with a stable `code`.",
				"- **Pagination**: cursor based — pass `next_cursor` back as `cursor`.",
				`- **Plans**: Free = ${PLAN_LIMITS.free.quota} requests in total, ${PLAN_LIMITS.free.maxPageSize} items per page, ${PLAN_LIMITS.free.requestsPerSecond} req/s. Pro and Business have monthly quotas, ${PLAN_LIMITS.pro.maxPageSize} items per page and prepaid credits.`,
				"",
				`Guides: ${DOCS_URL}`,
			].join("\n"),
			contact: { name: "Hexascore", url: DOCS_URL },
			license: { name: "Proprietary", url: "https://hexascore.com/terms" },
		},
		servers: [
			{ url: API_BASE_URL, description: "Production" },
			{ url: "/api/v1", description: "This deployment (same API under /api/v1)" },
		],
		security: [{ bearerAuth: [] }],
		tags: [
			...types.map((t) => ({ name: NAMES[t].tag, description: `${NAMES[t].many}: list, detail and genres.` })),
			{ name: "Search", description: "Find items by text or by an id from another database." },
			{ name: "Account", description: "Your quota, credits and rate limit." },
		],
		externalDocs: { description: "Guides", url: DOCS_URL },
		paths: {
			...Object.assign({}, ...types.map(typePaths)),
			"/search": {
				get: operation({
					id: "searchItems",
					tag: "Search",
					summary: "Search items",
					description: `Typo-tolerant search over titles and synopses in every locale, best match first (up to ${SEARCH_MAX} results).`,
					query: searchQuerySchema(PLAN_LIMITS.business.maxPageSize),
					schema: "SearchResult",
					errors: ["bad_request", ...COMMON_ERRORS],
				}),
			},
			"/lookup": {
				get: operation({
					id: "lookupItem",
					tag: "Search",
					summary: "Look up by external id",
					description:
						"Finds the item for an IMDb, Steam, IGDB, TMDB, ISBN-13, Open Library or Wikidata id. Ids must match exactly (e.g. `tt0111161`, `Q42`).",
					query: lookupQuerySchema,
					schema: "ItemDetail",
					errors: ["bad_request", "not_found", ...COMMON_ERRORS],
				}),
			},
			"/me/usage": {
				get: operation({
					id: "getUsage",
					tag: "Account",
					summary: "Get usage",
					description:
						"Plan, quota used in the current period, prepaid credits and rate limit. Counts as one request.",
					query: z.strictObject({}),
					schema: "Usage",
					errors: COMMON_ERRORS,
				}),
			},
		},
		components: {
			schemas: Object.fromEntries(Object.entries(schemas).map(([id, s]) => [id, clean(s)])),
			responses: Object.fromEntries(PROBLEM_CODES.map((code) => [responseName(code), problemResponse(code)])),
			securitySchemes: {
				bearerAuth: {
					type: "http",
					scheme: "bearer",
					bearerFormat: "hx_live_<32 chars>",
					description: "API key from /profile/api, sent as `Authorization: Bearer hx_live_…`.",
				},
			},
		},
	};
}

export function serializeOpenApi(): string {
	return `${JSON.stringify(buildOpenApiDocument(), null, 2)}\n`;
}
