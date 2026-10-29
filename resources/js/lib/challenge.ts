interface Challenge {
	required: boolean;
	c: string;
	bits: number;
}

// challengeHeaders asks whether `route` needs proof of work right now and, if
// so, solves it in a worker. The result goes into the visit's `headers`; it is
// empty when no challenge is due or the challenge endpoint is unreachable (the
// server then answers 428 itself).
export async function challengeHeaders(route: string): Promise<Record<string, string>> {
	try {
		const res = await fetch(`/challenge?route=${encodeURIComponent(route)}`, {
			headers: { Accept: "application/json" },
		});
		if (!res.ok) return {};
		const ch = (await res.json()) as Challenge;
		if (!ch.required) return {};

		const worker = new Worker(new URL("./pow-worker.ts", import.meta.url), { type: "module" });
		const n = await new Promise<string>((resolve, reject) => {
			worker.onmessage = (e: MessageEvent<string>) => resolve(e.data);
			worker.onerror = (e) => reject(e);
			worker.postMessage({ c: ch.c, bits: ch.bits });
		});
		worker.terminate();
		return { "X-Challenge": `${ch.c}:${n}` };
	} catch {
		return {};
	}
}
