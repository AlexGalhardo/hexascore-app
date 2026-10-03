import { type DocPageProps, docMetadata, enterDocPage } from "@/components/docs/metadata";
import { C, Code, DocTitle, H2, Note, P } from "@/components/docs/prose";
import { Link } from "@/i18n/navigation";
import { API_BASE_URL } from "@/lib/api/openapi";

export const generateMetadata = docMetadata("authentication");

export default async function AuthenticationPage(props: DocPageProps) {
	await enterDocPage(props);
	return (
		<>
			<DocTitle lead="Every request is authenticated with a secret API key sent as a Bearer token.">
				Authentication
			</DocTitle>

			<H2 id="bearer">Bearer token</H2>
			<P>
				Send the key in the <C>Authorization</C> header. Keys look like <C>hx_live_</C> followed by 32 letters
				and digits. Query-string keys are not accepted, so keys never end up in access logs or browser history.
			</P>
			<Code title="HTTP">{`GET /v1/movies?sort=-score HTTP/1.1
Host: ${new URL(API_BASE_URL).host}
Authorization: Bearer hx_live_0123456789ABCDEFGHIJabcdefghij`}</Code>
			<P>
				A missing, malformed, unknown or revoked key returns <C>401</C> with{" "}
				<Link href="/docs/errors#unauthorized">
					<C>code: "unauthorized"</C>
				</Link>{" "}
				and a <C>WWW-Authenticate: Bearer</C> header.
			</P>

			<H2 id="managing-keys">Managing keys</H2>
			<ul className="my-3 list-disc space-y-2 pl-6">
				<li>
					Create and revoke keys at{" "}
					<Link href="/profile/api" className="underline underline-offset-4">
						/profile/api
					</Link>{" "}
					(verified email required, up to 5 active keys).
				</li>
				<li>The full key is shown once. Only a SHA-256 hash is stored; we cannot recover it for you.</li>
				<li>Revocation is immediate. Rotate by creating a new key, deploying it, then revoking the old one.</li>
				<li>All keys of an account share the same plan, quota and per-second rate limit.</li>
			</ul>

			<H2 id="security">Keep keys secret</H2>
			<Note>
				Call the API from your server. A key embedded in a web page or mobile app can be extracted and used up
				your quota. If a key leaks, revoke it right away.
			</Note>
		</>
	);
}
