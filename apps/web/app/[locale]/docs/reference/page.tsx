import { getTranslations } from "next-intl/server";
import { ApiReference } from "@/components/api/api-reference";
import { type DocPageProps, docMetadata, enterDocPage } from "@/components/docs/metadata";

export const generateMetadata = docMetadata("reference");

export default async function ReferencePage(props: DocPageProps) {
	await enterDocPage(props);
	const t = await getTranslations("api.docs");
	return (
		<>
			<h1 className="sr-only">{t("nav.reference")}</h1>
			<ApiReference loadingLabel={t("loading")} />
		</>
	);
}
