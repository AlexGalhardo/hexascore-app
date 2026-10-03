import { getTranslations } from "next-intl/server";
import { type DocPageProps, docMetadata, enterDocPage } from "@/components/docs/metadata";
import { Code, DocTitle, H2, P, richCode } from "@/components/docs/prose";
import { API_BASE_URL } from "@/lib/api/openapi";
import { PLAN_LIMITS } from "@/lib/api/plans";

export const generateMetadata = docMetadata("pagination");

const loop = `async function* allGames(params = {}) {
  let cursor = null;
  do {
    const query = new URLSearchParams({ ...params, limit: "100", ...(cursor ? { cursor } : {}) });
    const res = await fetch(\`${API_BASE_URL}/games?\${query}\`, {
      headers: { Authorization: \`Bearer \${process.env.HEXASCORE_API_KEY}\` },
    });
    if (!res.ok) throw new Error((await res.json()).detail);
    const page = await res.json();
    yield* page.data;
    cursor = page.next_cursor;
  } while (cursor);
}

for await (const game of allGames({ genre: "Adventure", sort: "-score" })) console.log(game.title);`;

export default async function PaginationPage(props: DocPageProps) {
	await enterDocPage(props);
	const [t, nav] = await Promise.all([getTranslations("docs.pagination"), getTranslations("api.docs.nav")]);
	return (
		<>
			<DocTitle lead={t("lead")}>{nav("pagination")}</DocTitle>

			<H2 id="cursor">{t("howHeading")}</H2>
			<P>{t.rich("howBody", { ...richCode, shape: "{ data, next_cursor }", param: "cursor=<next_cursor>" })}</P>
			<ul className="my-3 list-disc space-y-2 pl-6">
				<li>{t.rich("sameParams", richCode)}</li>
				<li>{t("opaque")}</li>
				<li>{t("stable")}</li>
			</ul>

			<H2 id="limit">{t("limitHeading")}</H2>
			<P>
				{t.rich("limitBody", {
					...richCode,
					free: PLAN_LIMITS.free.maxPageSize,
					paid: PLAN_LIMITS.pro.maxPageSize,
				})}
			</P>

			<H2 id="example">{t("exampleHeading")}</H2>
			<Code title="JavaScript">{loop}</Code>

			<H2 id="search">{t("searchHeading")}</H2>
			<P>{t.rich("searchBody", richCode)}</P>
		</>
	);
}
