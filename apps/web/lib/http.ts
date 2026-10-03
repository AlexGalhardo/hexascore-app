export class UpstreamError extends Error {
	constructor(
		message: string,
		readonly status: number | null,
		readonly url: string,
	) {
		super(message);
		this.name = "UpstreamError";
	}
}

// All outbound calls go through here so none can hang a request: timeout + explicit error with the URL.
export async function fetchWithTimeout(url: string, init: RequestInit = {}, timeoutMs = 8000): Promise<Response> {
	try {
		const response = await fetch(url, { ...init, signal: init.signal ?? AbortSignal.timeout(timeoutMs) });
		if (!response.ok) throw new UpstreamError(`HTTP ${response.status} from ${url}`, response.status, url);
		return response;
	} catch (error) {
		if (error instanceof UpstreamError) throw error;
		const reason =
			error instanceof Error && error.name === "TimeoutError" ? `timed out after ${timeoutMs} ms` : String(error);
		throw new UpstreamError(`Request to ${url} failed: ${reason}`, null, url);
	}
}
