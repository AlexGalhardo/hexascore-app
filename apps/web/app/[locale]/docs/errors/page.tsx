import { getTranslations } from "next-intl/server";
import { type DocPageProps, docMetadata, enterDocPage } from "@/components/docs/metadata";
import { Code, DocTitle, H2, P, richCode, richLink } from "@/components/docs/prose";
import { PROBLEM_CODES, PROBLEM_STATUS, type ProblemCode } from "@/lib/api/problem";

export const generateMetadata = docMetadata("errors");

const EXAMPLE = `HTTP/1.1 400 Bad Request
Content-Type: application/problem+json
X-Request-Id: 5f0c8d3e-4b7a-4f43-9a51-2a8c4f3f9b10

{
  "type": "https://hexascore.app/docs/errors#bad_request",
  "title": "Invalid request",
  "status": 400,
  "detail": "One or more query parameters are invalid.",
  "code": "bad_request",
  "request_id": "5f0c8d3e-4b7a-4f43-9a51-2a8c4f3f9b10",
  "errors": [{ "path": "limit", "message": "Your plan allows at most 10 results per page." }]
}`;

// Target of the `<link>` tag in `docs.errors.codes.<code>.fix`.
const FIX_LINK: Partial<Record<ProblemCode, string>> = {
	unauthorized: "/profile/api",
	payment_required: "/pricing",
	forbidden: "/contact",
};

export default async function ErrorsPage(props: DocPageProps) {
	await enterDocPage(props);
	const [t, nav] = await Promise.all([getTranslations("docs.errors"), getTranslations("api.docs.nav")]);
	return (
		<>
			<DocTitle lead={t("lead")}>{nav("errors")}</DocTitle>

			<H2 id="format">{t("formatHeading")}</H2>
			<P>{t.rich("formatBody", richCode)}</P>
			<Code title={t("exampleTitle")}>{EXAMPLE}</Code>

			<H2 id="codes">{t("codesHeading")}</H2>
			{PROBLEM_CODES.map((code) => {
				const [status] = PROBLEM_STATUS[code];
				const href = FIX_LINK[code];
				return (
					<section key={code} id={code} className="scroll-mt-20 border-t border-border py-5">
						<h3 className="font-mono text-lg font-semibold">
							<a href={`#${code}`}>
								{status} {code}
							</a>
						</h3>
						<p className="text-sm text-muted">{t(`codes.${code}.title`)}</p>
						<dl className="mt-3 grid gap-2 text-sm leading-6 sm:grid-cols-[6rem_1fr]">
							<dt className="font-semibold">{t("when")}</dt>
							<dd>{t.rich(`codes.${code}.when`, richCode)}</dd>
							<dt className="font-semibold">{t("fix")}</dt>
							<dd>
								{t.rich(`codes.${code}.fix`, href ? { ...richCode, link: richLink(href) } : richCode)}
							</dd>
						</dl>
					</section>
				);
			})}

			<H2 id="retries">{t("retriesHeading")}</H2>
			<P>{t.rich("retriesBody", richCode)}</P>
		</>
	);
}
