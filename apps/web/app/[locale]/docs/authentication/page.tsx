import { getTranslations } from "next-intl/server";
import type { ReactNode } from "react";
import { type DocPageProps, docMetadata, enterDocPage } from "@/components/docs/metadata";
import { C, Code, DocTitle, H2, Note, P, richCode, richLink } from "@/components/docs/prose";
import { Link } from "@/i18n/navigation";
import { API_BASE_URL } from "@/lib/api/openapi";

export const generateMetadata = docMetadata("authentication");

export default async function AuthenticationPage(props: DocPageProps) {
	await enterDocPage(props);
	const [t, nav] = await Promise.all([getTranslations("docs.authentication"), getTranslations("api.docs.nav")]);
	return (
		<>
			<DocTitle lead={t("lead")}>{nav("authentication")}</DocTitle>

			<H2 id="bearer">{t("bearerHeading")}</H2>
			<P>{t.rich("bearerBody", richCode)}</P>
			<Code title="HTTP">{`GET /v1/movies?sort=-score HTTP/1.1
Host: ${new URL(API_BASE_URL).host}
Authorization: Bearer hx_live_0123456789ABCDEFGHIJabcdefghij`}</Code>
			<P>
				{t.rich("unauthorized", {
					...richCode,
					problemCode: 'code: "unauthorized"',
					link: (chunks: ReactNode) => (
						<Link href="/docs/errors#unauthorized">
							<C>{chunks}</C>
						</Link>
					),
				})}
			</P>

			<H2 id="managing-keys">{t("managingHeading")}</H2>
			<ul className="my-3 list-disc space-y-2 pl-6">
				<li>{t.rich("manageCreate", { link: richLink("/profile/api", "underline underline-offset-4") })}</li>
				<li>{t("manageOnce")}</li>
				<li>{t("manageRevoke")}</li>
				<li>{t("manageShared")}</li>
			</ul>

			<H2 id="security">{t("securityHeading")}</H2>
			<Note>{t("securityNote")}</Note>
		</>
	);
}
