import { type DocPageProps, docMetadata, enterDocPage } from "@/components/docs/metadata";
import { C, Code, DocTitle, H2, Note, P } from "@/components/docs/prose";
import { Link } from "@/i18n/navigation";
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
      "url": "https://hexascore.com/games/the-legend-of-zelda"
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
	return (
		<>
			<DocTitle lead="Community scores for games, movies, TV series and books — 0 to 100, in six categories per item — over a read-only JSON API.">
				Quickstart
			</DocTitle>

			<H2 id="key">1. Create an API key</H2>
			<P>
				<Link href="/profile/api" className="underline underline-offset-4">
					Open your API dashboard
				</Link>{" "}
				and create a key. It starts with <C>hx_live_</C> and is shown only once — store it in an environment
				variable, never in client-side code:
			</P>
			<Code title="shell">{'export HEXASCORE_API_KEY="hx_live_..."'}</Code>
			<P>
				The Free plan includes {PLAN_LIMITS.free.quota} requests in total — enough to build and test an
				integration.
			</P>

			<H2 id="first-request">2. Make your first request</H2>
			<P>
				The base URL is <C>{API_BASE_URL}</C>. Every endpoint is a <C>GET</C>. Here are the five best-rated
				games:
			</P>
			<Code title="cURL">{curl}</Code>
			<Note>
				Until the <C>api.</C> domain is live, use the same API on the website host: <C>{`${APP_URL}/api/v1`}</C>
				. Paths, parameters and responses are identical.
			</Note>
			<Code title="JavaScript (Node 18+, Bun, Deno)">{js}</Code>
			<Code title="Python (requests)">{python}</Code>

			<H2 id="response">3. Read the response</H2>
			<P>
				Lists return <C>data</C> plus a <C>next_cursor</C> for the next page. Field names are snake_case and
				dates are ISO 8601. <C>score</C> and <C>categories</C> are <C>null</C> until an item has 5 ratings.
			</P>
			<Code title="200 OK">{response}</Code>

			<H2 id="next">Next steps</H2>
			<ul className="my-3 list-disc space-y-2 pl-6">
				<li>
					Fetch one item with <C>GET /games/{"{id}"}</C> for external ids, trailers and the licensed synopsis.
				</li>
				<li>
					Map ids you already have with <C>GET /lookup?source=imdb&id=tt0111161</C> (also <C>steam</C>,{" "}
					<C>igdb</C>, <C>tmdb</C>, <C>isbn13</C>, <C>openlibrary</C>, <C>wikidata</C>).
				</li>
				<li>
					Handle <Link href="/docs/errors">errors</Link> and <Link href="/docs/rate-limits">rate limits</Link>
					, then browse the <Link href="/docs/reference">API reference</Link>.
				</li>
			</ul>

			<H2 id="sdks">SDKs</H2>
			<P>
				Typed clients are generated from the <a href="/openapi.json">OpenAPI document</a>: TypeScript (
				<C>@hexascore/sdk</C>) and Python (<C>hexascore</C>). They are not on npm/PyPI yet — build them from{" "}
				<C>sdks/</C> in the repository until the first release.
			</P>
			<Code title="TypeScript SDK">{sdk}</Code>
		</>
	);
}
