import { getTranslations } from "next-intl/server";
import type { ReactNode } from "react";
import { type DocPageProps, docMetadata, enterDocPage } from "@/components/docs/metadata";
import { Code, DocTitle, H2, P, richCode } from "@/components/docs/prose";

export const generateMetadata = docMetadata("changelog");

const DEPRECATION = `HTTP/1.1 200 OK
Deprecation: @1798761600
Sunset: Tue, 01 Jun 2027 00:00:00 GMT
Link: <https://hexascore.app/docs/changelog>; rel="deprecation"`;

// Static files (not localized routes), so plain anchors instead of the i18n Link.
const file = (href: string) => (chunks: ReactNode) => <a href={href}>{chunks}</a>;

export default async function ChangelogPage(props: DocPageProps) {
	await enterDocPage(props);
	const [t, nav] = await Promise.all([getTranslations("docs.changelog"), getTranslations("api.docs.nav")]);
	return (
		<>
			<DocTitle lead={t("lead")}>{nav("changelog")}</DocTitle>

			<H2 id="policy">{t("policyHeading")}</H2>
			<P>{t.rich("policyAdditive", richCode)}</P>
			<P>{t.rich("policyBreaking", richCode)}</P>

			<H2 id="deprecation">{t("deprecationHeading")}</H2>
			<ul className="my-3 list-disc space-y-2 pl-6">
				<li>{t("deprecationWindow")}</li>
				<li>{t.rich("deprecationHeaders", richCode)}</li>
				<li>{t("deprecationEmail")}</li>
			</ul>
			<Code title={t("deprecationTitle")}>{DEPRECATION}</Code>

			<H2 id="changes">{t("changesHeading")}</H2>
			<section className="border-t border-border py-4">
				<h3 className="font-semibold">{t("openapiI18nTitle")}</h3>
				<ul className="mt-2 list-disc space-y-1 pl-6 text-sm leading-6">
					<li>
						{t.rich("openapiI18nBody", {
							...richCode,
							ptBr: file("/openapi.pt-BR.json"),
							es: file("/openapi.es.json"),
						})}
					</li>
				</ul>
			</section>
			<section className="border-t border-border py-4">
				<h3 className="font-semibold">{t("timeToBeatTitle")}</h3>
				<ul className="mt-2 list-disc space-y-1 pl-6 text-sm leading-6">
					<li>{t.rich("timeToBeatBody", richCode)}</li>
				</ul>
			</section>
			<section className="border-t border-border py-4">
				<h3 className="font-semibold">{t("firstTitle")}</h3>
				<ul className="mt-2 list-disc space-y-1 pl-6 text-sm leading-6">
					<li>{t.rich("firstLists", richCode)}</li>
					<li>{t.rich("firstEndpoints", richCode)}</li>
					<li>{t.rich("firstOpenapi", { openapi: file("/openapi.json") })}</li>
				</ul>
			</section>
		</>
	);
}
