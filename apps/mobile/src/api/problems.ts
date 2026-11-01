import { ApiError } from "../../../../packages/app-client/src/api";
import { catalogues } from "../i18n";

export type DescribedError = { key: string; requestId: string | undefined };

const UNEXPECTED = "errors.unexpected";

const hasEnglishKey = (key: string): boolean => key in catalogues.en;

// RFC 9457 `code` (kebab-case, SDD 10 §6.1) maps to `errors.<code>`; anything the
// catalogue does not know shows `errors.unexpected` plus the request_id.
export function describeError(
	error: unknown,
	hasKey: (key: string) => boolean = hasEnglishKey,
): DescribedError {
	if (!(error instanceof ApiError)) return { key: UNEXPECTED, requestId: undefined };
	// The generated ApiProblem type omits `code`; read it as unknown.
	const code: unknown = (error.problem as { code?: unknown } | undefined)?.code;
	const key = typeof code === "string" ? `errors.${code}` : UNEXPECTED;
	return { key: hasKey(key) ? key : UNEXPECTED, requestId: error.problem?.request_id };
}
