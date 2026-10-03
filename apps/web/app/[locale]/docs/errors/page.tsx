import type { ReactNode } from "react";
import { type DocPageProps, docMetadata, enterDocPage } from "@/components/docs/metadata";
import { C, Code, DocTitle, H2, P } from "@/components/docs/prose";
import { Link } from "@/i18n/navigation";
import { PROBLEM_STATUS, type ProblemCode } from "@/lib/api/problem";

export const generateMetadata = docMetadata("errors");

const EXAMPLE = `HTTP/1.1 400 Bad Request
Content-Type: application/problem+json
X-Request-Id: 5f0c8d3e-4b7a-4f43-9a51-2a8c4f3f9b10

{
  "type": "https://hexascore.com/docs/errors#bad_request",
  "title": "Invalid request",
  "status": 400,
  "detail": "One or more query parameters are invalid.",
  "code": "bad_request",
  "request_id": "5f0c8d3e-4b7a-4f43-9a51-2a8c4f3f9b10",
  "errors": [{ "path": "limit", "message": "Your plan allows at most 10 results per page." }]
}`;

const GUIDE: Record<ProblemCode, { when: ReactNode; fix: ReactNode }> = {
	bad_request: {
		when: (
			<>
				A query parameter is unknown or out of range, the <C>cursor</C> was tampered with or belongs to another
				query, or <C>limit</C> is above your plan's page size (10 on Free). <C>errors</C> lists each problem
				with its <C>path</C>.
			</>
		),
		fix: "Fix the listed parameters. Not counted against your quota when the parameters fail validation.",
	},
	unauthorized: {
		when: (
			<>
				No <C>Authorization: Bearer</C> header, a malformed key, or a key that does not exist or was revoked.
			</>
		),
		fix: (
			<>
				Check the header and the key at <Link href="/profile/api">/profile/api</Link>. Never retried
				automatically.
			</>
		),
	},
	payment_required: {
		when: (
			<>
				The quota is exhausted: the 100 lifetime requests on Free, or the monthly quota plus prepaid credits on
				paid plans. The body includes <C>plan</C> and <C>limit</C>.
			</>
		),
		fix: (
			<>
				Upgrade or buy credits at <Link href="/pricing">/pricing</Link>. Paid quotas reset on the 1st of each
				month (UTC).
			</>
		),
	},
	forbidden: {
		when: "The account that owns the key is suspended.",
		fix: (
			<>
				Contact us through <Link href="/contact">/contact</Link> with the <C>request_id</C>.
			</>
		),
	},
	not_found: {
		when: (
			<>
				Unknown item id, an id that belongs to another type (e.g. a game id under <C>/movies</C>), an unknown
				type path, or no match for <C>/lookup</C>.
			</>
		),
		fix: "Check the id and the type. Lookups require an exact id (e.g. tt0111161).",
	},
	rate_limited: {
		when: (
			<>
				More requests per second than your plan allows. The <C>Retry-After</C> header says how many seconds to
				wait. Not counted against your quota.
			</>
		),
		fix: "Wait Retry-After seconds and retry; space out bursts or upgrade for a higher rate.",
	},
	internal_error: {
		when: "Something failed on our side. The request is refunded to your quota.",
		fix: (
			<>
				Retry with exponential backoff. If it persists, send us the <C>request_id</C>.
			</>
		),
	},
	service_unavailable: {
		when: (
			<>
				A dependency is temporarily down. Comes with <C>Retry-After</C>.
			</>
		),
		fix: "Retry after the indicated delay.",
	},
};

export default async function ErrorsPage(props: DocPageProps) {
	await enterDocPage(props);
	return (
		<>
			<DocTitle lead="Errors use RFC 9457 problem details with a stable machine-readable code.">Errors</DocTitle>

			<H2 id="format">Format</H2>
			<P>
				Every error has <C>Content-Type: application/problem+json</C>. Branch on <C>code</C> (stable) rather
				than on <C>title</C> or <C>detail</C> (human text that may change). <C>type</C> links to the matching
				section below, and <C>request_id</C> equals the <C>X-Request-Id</C> header — include it when you contact
				support.
			</P>
			<Code title="Example">{EXAMPLE}</Code>

			<H2 id="codes">Codes</H2>
			{(Object.keys(GUIDE) as ProblemCode[]).map((code) => {
				const [status, title] = PROBLEM_STATUS[code];
				return (
					<section key={code} id={code} className="scroll-mt-20 border-t border-border py-5">
						<h3 className="font-mono text-lg font-semibold">
							<a href={`#${code}`} className="hover:underline">
								{status} {code}
							</a>
						</h3>
						<p className="text-sm text-muted">{title}</p>
						<dl className="mt-3 grid gap-2 text-sm leading-6 sm:grid-cols-[6rem_1fr]">
							<dt className="font-semibold">When</dt>
							<dd>{GUIDE[code].when}</dd>
							<dt className="font-semibold">What to do</dt>
							<dd>{GUIDE[code].fix}</dd>
						</dl>
					</section>
				);
			})}

			<H2 id="retries">Retries</H2>
			<P>
				Retry only <C>429</C>, <C>500</C> and <C>503</C>, honouring <C>Retry-After</C> when present and backing
				off exponentially otherwise. Every endpoint is a read-only <C>GET</C>, so retries are always safe.
			</P>
		</>
	);
}
