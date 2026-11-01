// The server's sign-up rules, mirrored for the form. SOURCE OF TRUTH:
// `nicknameRe` in app/accounts/nickname.go and `EmailMin`/`EmailMax` in
// app/requests/auth.go. Change both in the same commit.

export const NICKNAME_MIN = 4;
export const NICKNAME_MAX = 16;
/** HTML `pattern` (anchored by the browser) for `^[a-z]{4,16}$`. */
export const NICKNAME_PATTERN = "[a-z]{4,16}";

export const EMAIL_MIN = 12;
export const EMAIL_MAX = 48;

/** Server error codes on the `nickname` field mapped to catalogue keys. */
export const NICKNAME_ERRORS: Readonly<Record<string, string>> = {
	"nickname-taken": "auth.errors.usernameTaken",
	"invalid-nickname": "auth.errors.invalidNickname",
};

// normalizeNickname applies what the server does before checking (lowercase)
// and drops whitespace as it is typed, since a nickname never has any.
export function normalizeNickname(value: string): string {
	return value.toLowerCase().replace(/\s+/g, "");
}

// isValidNickname is the server rule.
export function isValidNickname(value: string): boolean {
	return /^[a-z]{4,16}$/.test(value);
}
