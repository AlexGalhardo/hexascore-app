import { router } from "@inertiajs/react";
import { useState } from "react";

export type ActionResult = { ok: true } | { ok: false; message: string };

type Fields = Record<string, string | boolean | number | undefined>;

async function errorMessage(res: Response): Promise<string> {
	try {
		const data: unknown = await res.json();
		if (data && typeof data === "object") {
			const o = data as Record<string, unknown>;
			for (const key of ["detail", "message", "title", "error"]) {
				if (typeof o[key] === "string" && o[key]) return o[key] as string;
			}
		}
	} catch {
		// Not JSON: fall through to the status line.
	}
	return `${res.status} ${res.statusText}`.trim();
}

/**
 * Posts a form-encoded body to an admin action. The handlers answer a success with a redirect (the fetch follows it
 * and ends 200) and a refusal with a 4xx whose JSON problem carries the message. The browser sends
 * Sec-Fetch-Site, which is what the panel's CSRF check reads for same-origin requests.
 */
export async function postAction(url: string, fields: Fields): Promise<ActionResult> {
	const body = new URLSearchParams();
	for (const [key, value] of Object.entries(fields)) {
		if (value === undefined || value === false) continue;
		body.set(key, value === true ? "on" : String(value));
	}
	try {
		const res = await fetch(url, {
			method: "POST",
			body,
			headers: { Accept: "application/json" },
			credentials: "same-origin",
		});
		return res.ok ? { ok: true } : { ok: false, message: await errorMessage(res) };
	} catch {
		return { ok: false, message: "Network error" };
	}
}

/** An action with pending and error state; a success reloads the page props. */
export function useAction(): {
	pending: boolean;
	error: string | null;
	clear: () => void;
	run: (url: string, fields: Fields) => Promise<boolean>;
} {
	const [pending, setPending] = useState(false);
	const [error, setError] = useState<string | null>(null);
	return {
		pending,
		error,
		clear: () => setError(null),
		run: async (url, fields) => {
			setPending(true);
			setError(null);
			const result = await postAction(url, fields);
			setPending(false);
			if (result.ok) {
				router.reload();
				return true;
			}
			setError(result.message);
			return false;
		},
	};
}
