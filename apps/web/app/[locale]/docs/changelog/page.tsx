import { type DocPageProps, docMetadata, enterDocPage } from "@/components/docs/metadata";
import { C, Code, DocTitle, H2, P } from "@/components/docs/prose";

export const generateMetadata = docMetadata("changelog");

const DEPRECATION = `HTTP/1.1 200 OK
Deprecation: @1798761600
Sunset: Tue, 01 Jun 2027 00:00:00 GMT
Link: <https://hexascore.com/docs/changelog>; rel="deprecation"`;

export default async function ChangelogPage(props: DocPageProps) {
	await enterDocPage(props);
	return (
		<>
			<DocTitle lead="The version is part of the path. v1 only changes in backwards-compatible ways.">
				Changelog & versioning
			</DocTitle>

			<H2 id="policy">Versioning policy</H2>
			<P>
				Within <C>/v1</C> we only make additive changes: new endpoints, new optional parameters, new response
				fields, new enum values in <C>categories</C> keys for new types, and new error <C>detail</C> texts.
				Ignore unknown fields and do not rely on field order.
			</P>
			<P>
				Breaking changes — removing or renaming a field, changing a type, making a parameter required — ship as
				a new path (<C>/v2</C>). Error <C>code</C> values, category keys and sort names are part of the
				contract.
			</P>

			<H2 id="deprecation">Deprecation and sunset</H2>
			<ul className="my-3 list-disc space-y-2 pl-6">
				<li>A deprecated version keeps working for at least 6 months after its successor is released.</li>
				<li>
					Its responses carry <C>Deprecation</C> (RFC 9745) and <C>Sunset</C> (RFC 8594) headers with the date
					it stops working, plus a <C>Link</C> to this page.
				</li>
				<li>
					Account owners with recent traffic on the old version are emailed at deprecation and 30 days before
					sunset.
				</li>
			</ul>
			<Code title="Headers on a deprecated version">{DEPRECATION}</Code>

			<H2 id="changes">Changes</H2>
			<section className="border-t border-border py-4">
				<h3 className="font-semibold">v1 — first release</h3>
				<ul className="mt-2 list-disc space-y-1 pl-6 text-sm leading-6">
					<li>
						Lists, details and genres for <C>games</C>, <C>movies</C>, <C>series</C> and <C>books</C>.
					</li>
					<li>
						<C>/search</C>, <C>/lookup</C> by external id and <C>/me/usage</C>.
					</li>
					<li>
						OpenAPI 3.1 document at <a href="/openapi.json">/openapi.json</a>.
					</li>
				</ul>
			</section>
		</>
	);
}
