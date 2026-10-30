interface Challenge {
	required: boolean;
	c: string;
	bits: number;
}

// After this long a slow device asks once for an easier challenge (SDD 13 §7.2); the server grants
// it at most once an hour per IP and otherwise returns the same difficulty.
export const FALLBACK_AFTER_MS = 10_000;

async function fetchChallenge(route: string, fallback: boolean): Promise<Challenge | null> {
	const query = `route=${encodeURIComponent(route)}${fallback ? "&fallback=1" : ""}`;
	const res = await fetch(`/challenge?${query}`, { headers: { Accept: "application/json" } });
	return res.ok ? ((await res.json()) as Challenge) : null;
}

function solveInWorker(ch: Challenge, workers: Worker[]): Promise<string> {
	const worker = new Worker(new URL("../challenge/worker.ts", import.meta.url), {
		type: "module",
	});
	workers.push(worker);
	return new Promise<string>((resolve, reject) => {
		worker.onmessage = (e: MessageEvent<string>) => resolve(`${ch.c}:${e.data}`);
		worker.onerror = (e) => reject(e);
		worker.postMessage({ c: ch.c, bits: ch.bits });
	});
}

// challengeHeaders asks whether `route` needs proof of work right now and, if
// so, solves it in a worker. The result goes into the visit's `headers`; it is
// empty when no challenge is due or the challenge endpoint is unreachable (the
// server then answers 428 itself). onChecking drives the page's aria-live status.
export async function challengeHeaders(
	route: string,
	onChecking?: (checking: boolean) => void,
): Promise<Record<string, string>> {
	const workers: Worker[] = [];
	let timer: ReturnType<typeof setTimeout> | undefined;
	try {
		const ch = await fetchChallenge(route, false);
		if (!ch?.required) return {};
		onChecking?.(true);

		const easier = new Promise<string>((resolve) => {
			timer = setTimeout(async () => {
				const fb = await fetchChallenge(route, true).catch(() => null);
				// Only a lower difficulty is worth a second worker; otherwise keep waiting for the first.
				if (fb?.required && fb.bits < ch.bits)
					solveInWorker(fb, workers).then(resolve, () => {});
			}, FALLBACK_AFTER_MS);
		});
		const proof = await Promise.race([solveInWorker(ch, workers), easier]);
		return { "X-Challenge": proof };
	} catch {
		return {};
	} finally {
		clearTimeout(timer);
		for (const w of workers) w.terminate();
		onChecking?.(false);
	}
}
