import type { ButtonHTMLAttributes } from "react";

interface Props extends Omit<ButtonHTMLAttributes<HTMLButtonElement>, "type"> {
	pending?: boolean;
}

// SubmitButton shows a spinner and a busy state while a submit is in flight. Before hydration the
// loader sets <html inert>; the `html[inert] [data-submit]` rule in app.css shows the same state,
// so a click in that gap is never a silent no-op. Without JavaScript the button stays usable.
export default function SubmitButton({
	pending = false,
	className = "",
	children,
	...rest
}: Props) {
	return (
		<button
			{...rest}
			type="submit"
			data-submit=""
			data-pending={pending || undefined}
			aria-busy={pending}
			disabled={pending || rest.disabled}
			className={`inline-flex items-center gap-2 rounded-md bg-indigo-600 px-4 py-2 font-medium text-white hover:bg-indigo-500 disabled:opacity-50 ${className}`}
		>
			<span className="submit-spinner" aria-hidden="true" />
			{children}
		</button>
	);
}
