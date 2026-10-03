import { APP_URL } from "../config";

export type ProblemCode =
	| "bad_request"
	| "unauthorized"
	| "payment_required"
	| "forbidden"
	| "not_found"
	| "rate_limited"
	| "internal_error"
	| "service_unavailable";

export const PROBLEM_STATUS: Record<ProblemCode, [number, string]> = {
	bad_request: [400, "Invalid request"],
	unauthorized: [401, "Missing or invalid API key"],
	payment_required: [402, "Quota exhausted"],
	forbidden: [403, "Forbidden"],
	not_found: [404, "Not found"],
	rate_limited: [429, "Too many requests"],
	internal_error: [500, "Internal error"],
	service_unavailable: [503, "Service unavailable"],
};

export const PROBLEM_CODES = Object.keys(PROBLEM_STATUS) as [ProblemCode, ...ProblemCode[]];

export type ProblemExtras = { errors?: { path: string; message: string }[]; [key: string]: unknown };

/** RFC 9457 problem details; `type` points at the human docs for the error. */
export function problem(
	code: ProblemCode,
	detail: string,
	requestId: string,
	extras: ProblemExtras = {},
	headers: HeadersInit = {},
): Response {
	const [status, title] = PROBLEM_STATUS[code];
	return Response.json(
		{ type: `${APP_URL}/docs/errors#${code}`, title, status, detail, code, request_id: requestId, ...extras },
		{ status, headers: { "content-type": "application/problem+json", "x-request-id": requestId, ...headers } },
	);
}
