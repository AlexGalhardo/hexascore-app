// Server Actions never throw to the client: they return a code the UI translates (messages: actions.<code>).
export type ActionError =
	| "unauthorized"
	| "invalid"
	| "rateLimited"
	| "bot"
	| "passwordAlreadySet"
	| "weakPassword"
	| "notScheduled"
	| "emailNotVerified"
	| "nicknameRequired"
	| "nicknameTaken"
	| "forbidden"
	| "notFound"
	| "limit"
	| "server";

export type ActionFailure = { ok: false; error: ActionError; fields?: string[] };
export type ActionResult<T extends object = Record<never, never>> = ({ ok: true } & T) | ActionFailure;

export const fail = (error: ActionError, fields?: string[]): ActionFailure => ({ ok: false, error, fields });
export const ok: ActionResult = { ok: true };
