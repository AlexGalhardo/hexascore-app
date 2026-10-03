import { type DocPageProps, docMetadata, enterDocPage } from "@/components/docs/metadata";
import { C, Code, DocTitle, H2, Note, P, Table } from "@/components/docs/prose";
import { Link } from "@/i18n/navigation";
import { CREDIT_PACK_REQUESTS, PLAN_IDS, PLAN_LIMITS } from "@/lib/api/plans";

export const generateMetadata = docMetadata("limits");

const fmt = (n: number) => new Intl.NumberFormat("en").format(n);

const HEADERS = `HTTP/1.1 200 OK
RateLimit-Limit: 50000
RateLimit-Remaining: 49231
RateLimit-Reset: 1036800
X-Quota-Remaining: 49231
X-Request-Id: 5f0c8d3e-4b7a-4f43-9a51-2a8c4f3f9b10`;

export default async function RateLimitsPage(props: DocPageProps) {
	await enterDocPage(props);
	return (
		<>
			<DocTitle lead="Each plan has a request quota, a per-second rate limit and a maximum page size.">
				Rate limits & quotas
			</DocTitle>

			<H2 id="plans">Plans</H2>
			<Table
				head={["Plan", "Quota", "Requests / second", "Max page size", "Credits"]}
				rows={PLAN_IDS.map((id) => {
					const p = PLAN_LIMITS[id];
					return [
						<span key="n" className="font-semibold capitalize">
							{id}
						</span>,
						p.period === "lifetime" ? `${fmt(p.quota)} in total` : `${fmt(p.quota)} / month`,
						fmt(p.requestsPerSecond),
						fmt(p.maxPageSize),
						id === "free" ? "—" : `Packs of ${fmt(CREDIT_PACK_REQUESTS)} requests`,
					];
				})}
			/>
			<P>
				Prices and upgrades: <Link href="/pricing">/pricing</Link>. Credits are prepaid requests spent only
				after the monthly quota runs out; they do not expire with the month.
			</P>

			<H2 id="counting">What counts</H2>
			<ul className="my-3 list-disc space-y-2 pl-6">
				<li>Every authenticated request that reaches an endpoint counts as one, including 404s.</li>
				<li>
					Not counted: <C>401</C>, <C>400</C> for invalid parameters, <C>429</C> and any <C>5xx</C>{" "}
					(refunded).
				</li>
				<li>
					<C>GET /me/usage</C> counts too — read the response headers of your regular calls when you can.
				</li>
				<li>Paid quotas reset at 00:00 UTC on the 1st of each month. The Free quota never resets.</li>
			</ul>

			<H2 id="headers">Response headers</H2>
			<P>Successful responses tell you where you stand (IETF RateLimit header fields):</P>
			<Code title="Headers">{HEADERS}</Code>
			<Table
				head={["Header", "Meaning"]}
				rows={[
					[<C key="h">RateLimit-Limit</C>, "Requests in the current period (your plan quota)."],
					[<C key="h">RateLimit-Remaining</C>, "Requests left before 402 (credits not included)."],
					[<C key="h">RateLimit-Reset</C>, "Seconds until the quota resets; 0 for the lifetime Free quota."],
					[<C key="h">X-Quota-Remaining</C>, "Same as RateLimit-Remaining."],
					[<C key="h">Retry-After</C>, "On 429 and 503: seconds to wait before retrying."],
				]}
			/>

			<H2 id="exceeding">When you hit a limit</H2>
			<Note>
				Too many requests per second → <Link href="/docs/errors#rate_limited">429 rate_limited</Link>: wait{" "}
				<C>Retry-After</C> seconds. Quota (and credits) exhausted →{" "}
				<Link href="/docs/errors#payment_required">402 payment_required</Link>: upgrade or buy credits; retrying
				will not help until the reset.
			</Note>
		</>
	);
}
