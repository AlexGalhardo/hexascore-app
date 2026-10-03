// RFC 9457 problem details for route handlers.
export function problem(status: number, title: string, detail?: string, headers?: HeadersInit): Response {
	return Response.json(
		{ type: "about:blank", title, status, ...(detail ? { detail } : {}) },
		{ status, headers: { "content-type": "application/problem+json", ...headers } },
	);
}
