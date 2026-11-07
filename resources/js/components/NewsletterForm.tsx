import { type FormEvent, useEffect, useId, useState } from "react";

import { useT } from "../lib/useT";

type State = "idle" | "loading" | "done" | "invalid" | "limited" | "error";

const byStatus: Record<number, State> = { 202: "done", 422: "invalid", 429: "limited" };

/**
 * The footer newsletter form. It posts straight to vukem-email's public subscribe endpoint (no cookies, so no
 * credentials option); the server answers 202 for every acceptable address, so "done" only asks to check the inbox.
 */
export default function NewsletterForm({ endpoint }: { endpoint: string }) {
	const t = useT();
	const id = useId();
	const [state, setState] = useState<State>("idle");
	// The form only works with JS (it posts with fetch), and hydration clears a value typed into the server-rendered
	// input, so it stays disabled until React has hydrated it.
	const [ready, setReady] = useState(false);
	useEffect(() => setReady(true), []);

	const submit = async (e: FormEvent<HTMLFormElement>) => {
		e.preventDefault();
		const data = new FormData(e.currentTarget);
		setState("loading");

		try {
			const res = await fetch(endpoint, {
				method: "POST",
				headers: { "Content-Type": "application/json", Accept: "application/json" },
				body: JSON.stringify({
					email: String(data.get("email") ?? ""),
					website: String(data.get("website") ?? ""),
				}),
			});
			setState(byStatus[res.status] ?? "error");
		} catch {
			setState("error");
		}
	};

	const message: Record<State, string> = {
		idle: "",
		loading: t("footer.newsletterSending"),
		done: t("footer.newsletterDone"),
		invalid: t("footer.newsletterInvalid"),
		limited: t("footer.newsletterRateLimited"),
		error: t("footer.newsletterError"),
	};
	const failed = state === "invalid" || state === "limited" || state === "error";

	return (
		<form
			onSubmit={(e) => void submit(e)}
			className="mx-auto w-full lg:mx-0 lg:max-w-xl"
			noValidate
		>
			{/* Honeypot: invisible to people and assistive tech, bots fill it. */}
			<input
				type="text"
				name="website"
				tabIndex={-1}
				autoComplete="off"
				aria-hidden="true"
				className="absolute -left-[9999px]"
			/>
			<div className="flex gap-2">
				<input
					type="email"
					name="email"
					required
					maxLength={254}
					autoComplete="email"
					disabled={!ready}
					aria-label={t("footer.newsletterPlaceholder")}
					aria-describedby={`${id}-status`}
					aria-invalid={failed ? true : undefined}
					placeholder={t("footer.newsletterPlaceholder")}
					className="min-h-11 min-w-0 flex-1 rounded-sm border border-footer-border bg-footer-border/30 px-3 text-sm text-footer-fg placeholder:text-footer-muted focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-footer-fg"
				/>
				<button
					type="submit"
					disabled={!ready || state === "loading"}
					className="min-h-11 shrink-0 rounded-sm bg-footer-fg px-5 text-sm font-bold text-footer-bg focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-footer-fg disabled:cursor-not-allowed disabled:opacity-60"
				>
					{t("footer.newsletterSubmit")}
				</button>
			</div>
			<p
				id={`${id}-status`}
				role={failed ? "alert" : "status"}
				className="mt-2 min-h-5 text-sm text-footer-muted"
			>
				{message[state]}
			</p>
		</form>
	);
}
