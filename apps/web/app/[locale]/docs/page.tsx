import { getTranslations } from "next-intl/server";
import type { ReactNode } from "react";
import { type DocPageProps, docMetadata, enterDocPage } from "@/components/docs/metadata";
import { Code, DocTitle, H2, Note, P, richCode, richLink } from "@/components/docs/prose";
import { API_BASE_URL } from "@/lib/api/openapi";
import { PLAN_LIMITS } from "@/lib/api/plans";
import { APP_URL } from "@/lib/config";

export const generateMetadata = docMetadata("quickstart");

const curl = `curl "${API_BASE_URL}/games?sort=-score&limit=5" \\
  -H "Authorization: Bearer $HEXASCORE_API_KEY"`;

const js = `const res = await fetch("${API_BASE_URL}/games?sort=-score&limit=5", {
  headers: { Authorization: \`Bearer \${process.env.HEXASCORE_API_KEY}\` },
});
if (!res.ok) {
  const problem = await res.json(); // RFC 9457: { code, detail, request_id, ... }
  throw new Error(\`\${problem.code}: \${problem.detail}\`);
}
const { data, next_cursor } = await res.json();
for (const game of data) console.log(game.score, game.title);`;

const python = `import os
import requests

res = requests.get(
    "${API_BASE_URL}/games",
    params={"sort": "-score", "limit": 5},
    headers={"Authorization": f"Bearer {os.environ['HEXASCORE_API_KEY']}"},
    timeout=10,
)
if not res.ok:
    problem = res.json()  # RFC 9457: code, detail, request_id
    raise RuntimeError(f"{problem['code']}: {problem['detail']}")
for game in res.json()["data"]:
    print(game["score"], game["title"])`;

const response = `{
  "data": [
    {
      "id": "0190a1b2-0000-7000-8000-000000000001",
      "type": "game",
      "slug": "the-legend-of-zelda",
      "title": "The Legend of Zelda",
      "release_date": "1986-02-21",
      "year": 1986,
      "score": 90,
      "categories": { "gameplay": 92, "fun": 91, "sound": 88, "technical": 85, "visual": 89, "storytelling": 95 },
      "rating_count": 1234,
      "genres": ["Adventure"],
      "platforms": ["NES"],
      "cover_url": null,
      "url": "https://hexascore.app/games/the-legend-of-zelda"
    }
  ],
  "next_cursor": "eyJzIjoiLXNjb3JlIiwidiI6IjkwIi..."
}`;

const sdk = `import { createHexascoreClient, listGames } from "@hexascore/sdk";

const client = createHexascoreClient({ apiKey: process.env.HEXASCORE_API_KEY });
const { data, error } = await listGames({ client, query: { sort: "-score", limit: 5 } });
if (error) throw new Error(\`\${error.code}: \${error.detail}\`);`;

export default async function QuickstartPage(props: DocPageProps) {
	await enterDocPage(props);
	const [t, nav] = await Promise.all([getTranslations("docs.quickstart"), getTranslations("api.docs.nav")]);
	return (
		<>
			<DocTitle lead={t("lead")}>{nav("quickstart")}</DocTitle>

			<H2 id="key">{t("keyHeading")}</H2>
			<P>
				{t.rich("keyBody", {
					...richCode,
					dashboard: richLink("/profile/api", "underline underline-offset-4"),
				})}
			</P>
			<Code title="shell">{'export HEXASCORE_API_KEY="hx_live_..."'}</Code>
			<P>{t("freePlan", { quota: PLAN_LIMITS.free.quota })}</P>

			<H2 id="first-request">{t("requestHeading")}</H2>
			<P>{t.rich("requestBody", { ...richCode, baseUrl: API_BASE_URL })}</P>
			<Code title="cURL">{curl}</Code>
			<Note>{t.rich("hostNote", { ...richCode, siteUrl: `${APP_URL}/api/v1` })}</Note>
			<Code title="JavaScript (Node 18+, Bun, Deno)">{js}</Code>
			<Code title="Python (requests)">{python}</Code>

			<H2 id="response">{t("responseHeading")}</H2>
			<P>{t.rich("responseBody", richCode)}</P>
			<Code title="200 OK">{response}</Code>

			<H2 id="next">{t("nextHeading")}</H2>
			<ul className="my-3 list-disc space-y-2 pl-6">
				<li>{t.rich("nextDetail", { ...richCode, path: "GET /games/{id}" })}</li>
				<li>{t.rich("nextLookup", { ...richCode, path: "GET /lookup?source=imdb&id=tt0111161" })}</li>
				<li>
					{t.rich("nextGuides", {
						errors: richLink("/docs/errors"),
						limits: richLink("/docs/rate-limits"),
						reference: richLink("/docs/reference"),
					})}
				</li>
			</ul>

			<H2 id="sdks">{t("sdkHeading")}</H2>
			<P>
				{t.rich("sdkBody", {
					...richCode,
					openapi: (chunks: ReactNode) => <a href="/openapi.json">{chunks}</a>,
					tsPackage: "@hexascore/sdk",
					pyPackage: "hexascore",
				})}
			</P>
			<Code title={t("sdkTitle")}>{sdk}</Code>
		</>
	);
}
