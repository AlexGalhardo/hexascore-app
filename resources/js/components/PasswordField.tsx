import { useId, useState } from "react";

import type { PasswordAdvice, PasswordCheck } from "../lib/password-policy";
import { useT } from "../lib/useT";
import { fieldClass } from "./Field";

interface Props {
	label: string;
	// The server's field name ("password", "password_confirmation"): it is what
	// a password manager reads. The DOM id comes from useId, so two fields with
	// the same name on one page still get distinct ids.
	name: string;
	value: string;
	onChange: (value: string) => void;
	autoComplete: "new-password" | "current-password";
	error?: string;
	// The server rules this value must satisfy, from lib/password-policy.
	// Omit them on a login form: it must not advertise the policy.
	checks?: readonly PasswordCheck[];
	// Optional strength rating. Kept visually apart from `checks` because it is
	// advice: the server accepts anything `checks` accepts, whatever this says.
	advice?: PasswordAdvice;
}

const meter = ["bg-border", "bg-red-500", "bg-amber-500", "bg-lime-500", "bg-green-600"];

// PasswordField is a password input with a show/hide toggle inside the field
// and, when `checks` are given, the server's rules checked as you type.
//
// The label is bound with htmlFor instead of wrapping the input: a <label> that
// encloses the toggle button folds the button's text into the field's
// accessible name, and the field announces as "Password Show password".
//
// Visibility lives in component state only, so a revealed password is hidden
// again on the next mount; revealing it is a deliberate act every time.
//
// The checklist carries role="list" explicitly: Tailwind's reset sets
// list-style: none, which drops the list role in some screen readers. Each
// rule's state is in its text too, never in its colour alone.
export default function PasswordField({
	label,
	name,
	value,
	onChange,
	autoComplete,
	error,
	checks,
	advice,
}: Props) {
	const t = useT();
	const [visible, setVisible] = useState(false);
	const id = useId();
	const errorId = `${id}-error`;
	const checksId = `${id}-checks`;
	const adviceId = `${id}-advice`;

	const describedBy = [error && errorId, checks && checksId, advice && adviceId]
		.filter(Boolean)
		.join(" ");

	return (
		<div>
			<label htmlFor={id} className="text-sm font-semibold">
				{label}
			</label>
			<div className="relative">
				<input
					id={id}
					name={name}
					type={visible ? "text" : "password"}
					autoComplete={autoComplete}
					required
					value={value}
					onChange={(e) => onChange(e.target.value)}
					aria-invalid={error ? true : undefined}
					aria-describedby={describedBy || undefined}
					className={`${fieldClass} pr-24`}
				/>
				<button
					type="button"
					onClick={() => setVisible((shown) => !shown)}
					aria-label={t(visible ? "auth.fields.hide" : "auth.fields.show")}
					aria-pressed={visible}
					aria-controls={id}
					className="absolute top-2 right-0 bottom-0 flex min-w-11 items-center justify-center px-3.5 text-[15px] font-semibold text-primary hover:underline focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-primary"
				>
					{t(visible ? "auth.fields.hideShort" : "auth.fields.showShort")}
				</button>
			</div>
			{error && (
				<p id={errorId} className="mt-1.5 text-sm text-danger">
					{error}
				</p>
			)}
			{checks && (
				// biome-ignore lint/a11y/noRedundantRoles: list-style none drops list semantics in Safari VoiceOver (see comment above)
				<ul id={checksId} role="list" className="mt-3 space-y-1 text-sm">
					{checks.map((check) => (
						<li key={check.id} className={check.met ? "text-fg" : "text-muted"}>
							<span
								aria-hidden="true"
								className={`mr-2 inline-block w-3 text-center ${check.met ? "text-score-high" : ""}`}
							>
								{check.met ? "✓" : "×"}
							</span>
							{check.label}
							<span className="sr-only">
								{check.met ? " (met)" : " (not met yet)"}
							</span>
						</li>
					))}
				</ul>
			)}
			{advice && (
				<div id={adviceId} className="mt-2 text-sm text-muted">
					<div aria-hidden="true" className="mb-1 flex gap-1">
						{[1, 2, 3, 4].map((step) => (
							<span
								key={step}
								className={`h-1 flex-1 rounded-full ${step <= advice.score ? meter[advice.score] : meter[0]}`}
							/>
						))}
					</div>
					Strength: {advice.label} (advice, not a requirement). {advice.hint}
				</div>
			)}
		</div>
	);
}
