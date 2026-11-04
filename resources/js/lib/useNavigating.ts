import { router } from "@inertiajs/react";
import { useEffect, useState } from "react";

/**
 * True while an Inertia visit to another page is loading, after a short delay so a fast page does not flash the
 * indicator. Reloads that keep the page (preserveState, such as a refresh after an action) do not count.
 */
export function useNavigating(delayMs = 120): boolean {
	const [loading, setLoading] = useState(false);

	useEffect(() => {
		let timer: ReturnType<typeof setTimeout> | undefined;
		const stop = () => {
			if (timer) clearTimeout(timer);
			timer = undefined;
			setLoading(false);
		};
		const offStart = router.on("start", (event) => {
			if (event.detail.visit.preserveState) return;
			if (timer) clearTimeout(timer);
			timer = setTimeout(() => setLoading(true), delayMs);
		});
		const offFinish = router.on("finish", stop);
		return () => {
			offStart();
			offFinish();
			stop();
		};
	}, [delayMs]);

	return loading;
}
