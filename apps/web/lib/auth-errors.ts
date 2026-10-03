// Maps Better Auth client errors to keys under `auth.errors` in the messages.
export type AuthErrorKey =
	| "invalidCredentials"
	| "emailNotVerified"
	| "rateLimited"
	| "network"
	| "generic"
	| "emailTaken"
	| "usernameTaken"
	| "invalidCode"
	| "locked"
	| "weakPassword"
	| "bot"
	| "invalidToken"
	| "passkeyFailed";

const BY_CODE: Record<string, AuthErrorKey> = {
	INVALID_EMAIL_OR_PASSWORD: "invalidCredentials",
	INVALID_USERNAME_OR_PASSWORD: "invalidCredentials",
	INVALID_PASSWORD: "invalidCredentials",
	CREDENTIAL_ACCOUNT_NOT_FOUND: "invalidCredentials",
	EMAIL_NOT_VERIFIED: "emailNotVerified",
	USER_ALREADY_EXISTS: "emailTaken",
	USER_ALREADY_EXISTS_USE_ANOTHER_EMAIL: "emailTaken",
	USERNAME_IS_ALREADY_TAKEN: "usernameTaken",
	INVALID_CODE: "invalidCode",
	INVALID_BACKUP_CODE: "invalidCode",
	INVALID_TWO_FACTOR_COOKIE: "invalidCode",
	OTP_HAS_EXPIRED: "invalidCode",
	TOO_MANY_ATTEMPTS_REQUEST_NEW_CODE: "locked",
	ACCOUNT_TEMPORARILY_LOCKED: "locked",
	WEAK_PASSWORD: "weakPassword",
	PASSWORD_TOO_SHORT: "weakPassword",
	PASSWORD_TOO_LONG: "weakPassword",
	BOT_DETECTED: "bot",
	INVALID_TOKEN: "invalidToken",
};

export type AuthClientError = { code?: string | undefined; status?: number | undefined } | null | undefined;

export function authErrorKey(error: AuthClientError): AuthErrorKey {
	if (!error) return "generic";
	if (error.status === 429) return "rateLimited";
	if (error.code && BY_CODE[error.code]) return BY_CODE[error.code] as AuthErrorKey;
	if (!error.status) return "network";
	return "generic";
}
