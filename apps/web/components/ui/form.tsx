import type { ComponentProps, ReactNode } from "react";

export const inputClass =
	"min-h-11 w-full rounded-md border border-border bg-bg px-3 text-base disabled:cursor-not-allowed disabled:bg-surface disabled:text-muted aria-invalid:border-danger";

const buttonVariants = {
	primary: "bg-fg text-bg hover:opacity-90",
	secondary: "border border-border hover:bg-surface",
	danger: "bg-danger text-bg hover:opacity-90",
} as const;

export function buttonClass(variant: keyof typeof buttonVariants = "primary", extra = ""): string {
	return `inline-flex min-h-11 items-center justify-center gap-2 rounded-md px-4 font-medium disabled:cursor-not-allowed disabled:opacity-50 ${buttonVariants[variant]} ${extra}`;
}

type FieldProps = {
	id: string;
	label: string;
	hint?: ReactNode;
	error?: string | undefined;
	children: ReactNode;
};

export function Field({ id, label, hint, error, children }: FieldProps) {
	return (
		<div className="flex flex-col gap-1">
			<label htmlFor={id} className="text-sm font-medium">
				{label}
			</label>
			{children}
			{hint && (
				<p id={`${id}-hint`} className="text-sm text-muted">
					{hint}
				</p>
			)}
			{error && (
				<p id={`${id}-error`} className="text-sm text-danger">
					{error}
				</p>
			)}
		</div>
	);
}

export function Input(props: ComponentProps<"input">) {
	return <input {...props} className={`${inputClass} ${props.className ?? ""}`} />;
}

type NoticeProps = { tone: "error" | "success" | "info"; children: ReactNode };

const tones = {
	error: "border-danger text-danger",
	success: "border-score-high",
	info: "border-border",
} as const;

// role=alert for errors interrupts screen readers; success/info are polite status updates.
export function Notice({ tone, children }: NoticeProps) {
	return (
		<div
			role={tone === "error" ? "alert" : "status"}
			className={`rounded-md border-l-4 bg-surface px-3 py-2 text-sm ${tones[tone]}`}
		>
			{children}
		</div>
	);
}

export function Divider({ label }: { label: string }) {
	return (
		<div className="flex items-center gap-3 text-sm text-muted">
			<span className="h-px flex-1 bg-border" />
			{label}
			<span className="h-px flex-1 bg-border" />
		</div>
	);
}
