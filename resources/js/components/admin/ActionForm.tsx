import { type ReactElement, type ReactNode, useState } from "react";

import { useAction } from "../../lib/adminAction";
import { cn } from "../../lib/cn";
import { ActionError } from "./ActionError";
import { Button, type ButtonProps } from "./ui/button";

/**
 * A form that posts to an admin action. It sends the fields form-encoded, shows the server's refusal under the
 * fields, and on success reloads the page props and clears itself.
 */
export function ActionForm({
	action,
	submit,
	variant = "default",
	className,
	transform,
	children,
}: {
	action: string;
	submit: ReactNode;
	variant?: ButtonProps["variant"];
	className?: string;
	/** Reshapes the collected fields before they are posted (a date input becoming the time the handler reads). */
	transform?: (data: Record<string, string>) => Record<string, string>;
	children: ReactNode;
}): ReactElement {
	const act = useAction();
	const [key, setKey] = useState(0);
	return (
		<form
			key={key}
			className={cn("grid gap-3", className)}
			onSubmit={async (e) => {
				e.preventDefault();
				const data = Object.fromEntries(new FormData(e.currentTarget)) as Record<
					string,
					string
				>;
				if (await act.run(action, transform ? transform(data) : data)) setKey((k) => k + 1);
			}}
		>
			{children}
			{act.error && (
				<div className="col-span-full">
					<ActionError message={act.error} />
				</div>
			)}
			<div className="col-span-full">
				<Button type="submit" variant={variant} disabled={act.pending}>
					{submit}
				</Button>
			</div>
		</form>
	);
}

/** A single button that posts to an admin action (unlock, delete, toggle); a refusal shows under it. */
export function ActionButton({
	action,
	fields = {},
	variant = "outline",
	size = "sm",
	children,
}: {
	action: string;
	fields?: Record<string, string | boolean | number | undefined>;
	variant?: ButtonProps["variant"];
	size?: ButtonProps["size"];
	children: ReactNode;
}): ReactElement {
	const act = useAction();
	return (
		<span className="inline-grid gap-1">
			<Button
				variant={variant}
				size={size}
				disabled={act.pending}
				onClick={() => void act.run(action, fields)}
			>
				{children}
			</Button>
			{act.error && (
				<span role="alert" className="max-w-56 text-xs text-danger">
					{act.error}
				</span>
			)}
		</span>
	);
}
