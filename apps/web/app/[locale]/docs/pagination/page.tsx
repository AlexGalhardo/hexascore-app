import { type DocPageProps, docMetadata, enterDocPage } from "@/components/docs/metadata";
import { C, Code, DocTitle, H2, P } from "@/components/docs/prose";
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
	return (
		<>
			<DocTitle lead="List endpoints use opaque cursors: stable under inserts, fast at any depth.">
				Pagination
			</DocTitle>

			<H2 id="cursor">How it works</H2>
			<P>
				<C>GET /games</C>, <C>/movies</C>, <C>/series</C> and <C>/books</C> return{" "}
				<C>{"{ data, next_cursor }"}</C>. To get the next page, repeat the same request with{" "}
				<C>cursor=&lt;next_cursor&gt;</C>. When <C>next_cursor</C> is <C>null</C> you reached the end.
			</P>
			<ul className="my-3 list-disc space-y-2 pl-6">
				<li>
					Keep every other parameter identical. A cursor from another <C>sort</C> returns <C>400</C>.
				</li>
				<li>Cursors are opaque: do not parse or build them, their format may change.</li>
				<li>Ordering is stable: ties are broken by item id, and unscored or undated items always come last.</li>
			</ul>

			<H2 id="limit">Page size</H2>
			<P>
				<C>limit</C> defaults to 10. The maximum depends on your plan: {PLAN_LIMITS.free.maxPageSize} on Free,{" "}
				{PLAN_LIMITS.pro.maxPageSize} on Pro and Business. A larger value returns <C>400 bad_request</C> — the
				API never silently truncates. Each page counts as one request.
			</P>

			<H2 id="example">Example: iterate over every page</H2>
			<Code title="JavaScript">{loop}</Code>

			<H2 id="search">Search</H2>
			<P>
				<C>GET /search</C> is not paginated: it returns the best matches (up to 50) for a query, best first.
			</P>
		</>
	);
}
