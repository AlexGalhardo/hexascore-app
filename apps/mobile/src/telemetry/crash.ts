import type { CrashConsent } from "./consent";

export type CrashContext = {
	appVersion: string;
	versionCode: number | null;
	androidApi: number | null;
};

/** Body of `POST /app/v1/client-errors` (SDD 05 §6.4): no user id, token or e-mail. */
export type CrashReport = CrashContext & { screen: string; message: string; stack: string };

export const STACK_MAX_BYTES = 8192;
// Server limits in app/appapi/telemetry.go (runes).
const MESSAGE_MAX = 1000;
const SCREEN_MAX = 200;

const EMAIL = /[^\s@"'<>]+@[^\s@"'<>]+\.[^\s@"'<>]+/g;
const TOKEN = /\b(?:Bearer\s+)?hx_[A-Za-z0-9_]+|\bBearer\s+\S+/g;

export function redact(text: string): string {
	return text.replace(EMAIL, "[email]").replace(TOKEN, "[token]");
}

const encoder = new TextEncoder();

function truncateBytes(text: string, maxBytes: number): string {
	if (encoder.encode(text).length <= maxBytes) return text;
	let out = "";
	let size = 0;
	for (const char of text) {
		const width = encoder.encode(char).length;
		if (size + width > maxBytes) break;
		out += char;
		size += width;
	}
	return out;
}

export function buildCrashReport(
	error: unknown,
	screen: string,
	context: CrashContext,
): CrashReport {
	const message = error instanceof Error ? error.message : String(error);
	const stack = error instanceof Error ? (error.stack ?? "") : "";
	// The server rejects screens with a query string; route patterns never need one.
	const route = screen.split("?")[0] ?? "/";
	return {
		...context,
		screen: Array.from(redact(route)).slice(0, SCREEN_MAX).join(""),
		message: Array.from(redact(message) || "error")
			.slice(0, MESSAGE_MAX)
			.join(""),
		stack: truncateBytes(redact(stack), STACK_MAX_BYTES),
	};
}

export type CrashReporterDeps = {
	consent: () => CrashConsent;
	send: (report: CrashReport) => Promise<unknown>;
	context: CrashContext;
	screen: () => string;
};

/** Never throws: a report must not crash the crash handler. */
export function createCrashReporter(deps: CrashReporterDeps): (error: unknown) => void {
	return (error) => {
		if (deps.consent() !== "granted") return;
		try {
			deps.send(buildCrashReport(error, deps.screen(), deps.context)).catch(() => {});
		} catch {
			// Building or sending failed synchronously; there is nowhere left to report it.
		}
	};
}

type GlobalHandler = (error: unknown, isFatal?: boolean) => void;
export type ErrorUtilsLike = {
	getGlobalHandler(): GlobalHandler;
	setGlobalHandler(handler: GlobalHandler): void;
};

/** Chains React Native's global JS handler, so the red box and the native crash still happen. */
export function installGlobalHandler(
	report: (error: unknown) => void,
	errorUtils: ErrorUtilsLike,
): () => void {
	const previous = errorUtils.getGlobalHandler();
	errorUtils.setGlobalHandler((error, isFatal) => {
		report(error);
		previous(error, isFatal);
	});
	return () => errorUtils.setGlobalHandler(previous);
}
