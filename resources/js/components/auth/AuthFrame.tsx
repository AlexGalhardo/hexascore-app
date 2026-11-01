import type { ReactNode } from "react";

import { useT } from "../../lib/useT";

interface Props {
	/** Large serif line on the left column. */
	heading: string;
	lede?: string;
	/** Form card title; it is the page h1 so focus lands on it after a visit. */
	title: string;
	children: ReactNode;
}

// The v1 members layout: intro column on the left, the form in a bordered card on the right; stacked on mobile.
export default function AuthFrame({ heading, lede, title, children }: Props): React.ReactElement {
	const t = useT();

	return (
		<div className="grid gap-6 py-4 md:grid-cols-[1fr_minmax(0,448px)] md:gap-16 md:py-12">
			<div>
				<p className="text-xs font-bold tracking-[0.1em] text-primary uppercase">
					{t("auth.members")}
				</p>
				<p className="mt-3 font-display text-[32px] leading-[1.05] font-bold tracking-[-0.02em] text-balance md:mt-6 md:text-[72px]">
					{heading}
				</p>
				{lede && <p className="mt-4 max-w-xl text-muted md:mt-6 md:text-lg">{lede}</p>}
			</div>
			{/* Mobile drops the card for a top rule; the title stays for screen readers and focus. */}
			<section className="border-t border-fg pt-4 md:border md:px-8 md:py-9">
				<h1 className="sr-only font-display text-[28px] leading-tight font-bold md:not-sr-only md:mb-5">
					{title}
				</h1>
				{children}
			</section>
		</div>
	);
}

/** The "or" rule between the Google button and the form. */
export function OrRule(): React.ReactElement {
	const t = useT();

	return (
		<p className="my-5 flex items-center gap-4 text-sm text-muted before:h-px before:flex-1 before:bg-border after:h-px after:flex-1 after:bg-border">
			{t("auth.or")}
		</p>
	);
}

// Google sign-in branding colours are fixed by Google, not HexaScore tokens: the one third-party exception
// documented in docs/openpencil/v1-editorial/README.md (round 2, item 15). Inter instead of Roboto (two-family budget).
const google =
	"inline-flex min-h-11 w-full items-center justify-center gap-3 rounded-xs border border-[#747775] bg-[#FFFFFF] px-4 text-[15px] font-medium text-[#1F1F1F] shadow-[0_2px_0_#747775] transition-[box-shadow,translate] duration-100 hover:bg-[#F8F8F8] focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fg active:translate-y-px active:shadow-[0_1px_0_#747775] motion-reduce:transition-none motion-reduce:active:translate-y-0 dark:border-[#8E918F] dark:bg-[#131314] dark:text-[#E3E3E3] dark:shadow-[0_2px_0_#8E918F] dark:hover:bg-[#1E1E20] dark:active:shadow-[0_1px_0_#8E918F]";

/** "Sign in with Google" (login) or "Sign up with Google" (signup); a full navigation to the OAuth redirect. */
export function GoogleButton({ mode }: { mode: "signIn" | "signUp" }): React.ReactElement {
	const t = useT();

	return (
		<a href="/auth/google/redirect" className={google}>
			<svg aria-hidden="true" viewBox="0 0 48 48" className="size-5 shrink-0">
				<path
					fill="#EA4335"
					d="M24 9.5c3.54 0 6.71 1.22 9.21 3.6l6.85-6.85C35.9 2.38 30.47 0 24 0 14.62 0 6.51 5.38 2.56 13.22l7.98 6.19C12.43 13.72 17.74 9.5 24 9.5z"
				/>
				<path
					fill="#4285F4"
					d="M46.98 24.55c0-1.57-.15-3.09-.38-4.55H24v9.02h12.94c-.58 2.96-2.26 5.48-4.78 7.18l7.73 6c4.51-4.18 7.09-10.36 7.09-17.65z"
				/>
				<path
					fill="#FBBC05"
					d="M10.53 28.59c-.48-1.45-.76-2.99-.76-4.59s.27-3.14.76-4.59l-7.98-6.19C.92 16.46 0 20.12 0 24c0 3.88.92 7.54 2.56 10.78l7.97-6.19z"
				/>
				<path
					fill="#34A853"
					d="M24 48c6.48 0 11.93-2.13 15.89-5.81l-7.73-6c-2.15 1.45-4.92 2.3-8.16 2.3-6.26 0-11.57-4.22-13.47-9.91l-7.98 6.19C6.51 42.62 14.62 48 24 48z"
				/>
			</svg>
			{t(mode === "signIn" ? "auth.googleSignIn" : "auth.googleSignUp")}
		</a>
	);
}

export const authLink = "font-semibold text-primary hover:underline";
