import type { ReactNode } from "react";

export function AuthCard({ title, children, footer }: { title: string; children: ReactNode; footer?: ReactNode }) {
	return (
		<section className="mx-auto flex w-full max-w-sm flex-col gap-6">
			<h1 className="font-serif text-3xl font-bold">{title}</h1>
			{children}
			{footer && <p className="border-t border-border pt-4 text-sm text-muted">{footer}</p>}
		</section>
	);
}
