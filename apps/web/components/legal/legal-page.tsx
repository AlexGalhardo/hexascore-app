import { getFormatter } from "next-intl/server";
import { LEGAL_UPDATED_AT, type LegalDocument } from "@/lib/legal";

export async function LegalPage({ doc }: { doc: LegalDocument }) {
	const format = await getFormatter();

	return (
		<article className="mx-auto flex max-w-3xl flex-col gap-6">
			<header>
				<h1 className="font-serif text-3xl font-bold sm:text-4xl">{doc.title}</h1>
				<p className="mt-2 text-sm text-muted">
					{doc.updatedLabel}:{" "}
					<time dateTime={LEGAL_UPDATED_AT} className="font-mono">
						{format.dateTime(new Date(`${LEGAL_UPDATED_AT}T12:00:00Z`), { dateStyle: "long" })}
					</time>
				</p>
			</header>
			<p className="text-lg">{doc.intro}</p>
			{doc.sections.map((section) => (
				<section key={section.heading} className="flex flex-col gap-3">
					<h2 className="font-serif text-xl font-bold">{section.heading}</h2>
					{section.body.map((paragraph) => (
						<p key={paragraph.slice(0, 48)} className="leading-relaxed">
							{paragraph}
						</p>
					))}
				</section>
			))}
		</article>
	);
}
