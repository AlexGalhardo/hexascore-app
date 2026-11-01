/**
 * Body of POST /app/v1/search/events (app/requests/search_event.go). The server already recorded
 * search.performed when it answered GET /search with this searchId, so the client only reports clicks,
 * and never the query text (SDD 07 §7.8: the server keeps a hash).
 */
export type SearchClick = {
	searchId: string;
	refId: string;
	/** 1-based position within its group. */
	position: number;
	/** Singular item type, e.g. "game". */
	group: string;
};

/** Fire and forget: a lost click must never block opening the item. */
export function searchClickSender(
	post: (body: SearchClick) => Promise<unknown>,
): (click: SearchClick) => void {
	return (click) => {
		post(click).catch(() => undefined);
	};
}
