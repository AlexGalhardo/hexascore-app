// The server's password policy, mirrored for live feedback in the browser.
//
// SOURCE OF TRUTH: app/requests/password.go. RegisterSchema and ResetPasswordSchema
// both validate the password with PasswordSchema (8 to 128 runes, upper, lower,
// digit, non-alphanumeric) and compare it with `password_confirmation` in their `Refine`. The constants and the checks
// below are that policy and nothing else: a form that promises rules the server
// does not enforce (or hides rules it does) is worse than no hint at all.
//
// When you change the policy, change it in app/requests/password.go and here in the
// same commit. Suggestions that are not rules (mixing character classes, say)
// belong in `passwordAdvice`, never in `checkPassword`.
//
// The length is counted in Unicode code points because `z.String().Min/.Max`
// count runes (`utf8.RuneCountInString`). JavaScript's `String.length` counts
// UTF-16 code units, which is larger for anything outside the BMP, so it would
// call a password long enough that the server then rejects. For the same reason
// the inputs carry no `minLength` attribute: the browser would count UTF-16
// units and disagree with both this file and the server.

/** Minimum password length in runes: `PasswordMin` in app/requests/password.go. */
export const PASSWORD_MIN_LENGTH = 8;

/** Maximum password length in runes: `PasswordMax` in app/requests/password.go. */
export const PASSWORD_MAX_LENGTH = 128;

// PasswordCheck is one rule of the server policy and whether the value meets it.
export interface PasswordCheck {
	id: string;
	label: string;
	met: boolean;
}

// PasswordAdvice rates a password that already passes every rule. It is a hint,
// not a requirement: the server accepts any password `checkPassword` accepts.
export interface PasswordAdvice {
	// 0 (weakest) to 4 (strongest).
	score: number;
	label: string;
	hint: string;
}

// runeLength counts Unicode code points, the unit the server's z.String()
// length rules use.
export function runeLength(value: string): number {
	return Array.from(value).length;
}

// checkPassword returns the server's password rules against value. Every rule
// here is one the server rejects on, and every rule it rejects on is here.
export function checkPassword(value: string): readonly PasswordCheck[] {
	const length = runeLength(value);

	return [
		{
			id: "min",
			label: `At least ${PASSWORD_MIN_LENGTH} characters`,
			met: length >= PASSWORD_MIN_LENGTH,
		},
		{
			id: "max",
			label: `At most ${PASSWORD_MAX_LENGTH} characters`,
			met: length <= PASSWORD_MAX_LENGTH,
		},
		{ id: "upper", label: "An uppercase letter", met: /\p{Lu}/u.test(value) },
		{ id: "lower", label: "A lowercase letter", met: /\p{Ll}/u.test(value) },
		{ id: "digit", label: "A digit", met: /\p{Nd}/u.test(value) },
		{ id: "special", label: "A symbol or space", met: /[^A-Za-z0-9]/.test(value) },
	];
}

// checkConfirmation returns the confirmation rule the schemas' Refine applies:
// password_confirmation must equal password.
export function checkConfirmation(
	password: string,
	confirmation: string,
): readonly PasswordCheck[] {
	return [
		{
			id: "match",
			label: "Matches the password above",
			met: confirmation.length > 0 && confirmation === password,
		},
	];
}

// passwordAdvice rates length and character variety. It never gates the form;
// it is shown apart from the rules so a password the API would take is never
// presented as rejected.
export function passwordAdvice(value: string): PasswordAdvice {
	const length = runeLength(value);
	const classes = [/\p{Ll}/u, /\p{Lu}/u, /\p{Nd}/u, /[^\p{L}\p{Nd}]/u].filter((re) =>
		re.test(value),
	).length;

	if (length === 0) {
		return {
			score: 0,
			label: "—",
			hint: "A few unrelated words are easy to remember and hard to guess.",
		};
	}

	const score = Math.max(0, Math.min(4, Math.floor(length / 8) + classes - 1));

	if (score <= 1) {
		return { score, label: "Weak", hint: "Longer is stronger: try a few unrelated words." };
	}

	if (score === 2) {
		return { score, label: "Fair", hint: "Add length, or another kind of character." };
	}

	if (score === 3) {
		return { score, label: "Good", hint: "A little more length would make it stronger." };
	}

	return { score, label: "Strong", hint: "Store it in a password manager." };
}
