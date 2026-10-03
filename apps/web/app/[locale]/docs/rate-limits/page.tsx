import { getTranslations } from "next-intl/server";
import { type DocPageProps, docMetadata, enterDocPage } from "@/components/docs/metadata";
import { C, Code, DocTitle, H2, Note, P, richCode, richLink, Table } from "@/components/docs/prose";
import { CREDIT_PACK_REQUESTS, PLAN_IDS, PLAN_LIMITS } from "@/lib/api/plans";

export const generateMetadata = docMetadata("limits");

const HEADERS = `HTTP/1.1 200 OK
RateLimit-Limit: 50000
RateLimit-Remaining: 49231
RateLimit-Reset: 1036800
X-Quota-Remaining: 49231
X-Request-Id: 5f0c8d3e-4b7a-4f43-9a51-2a8c4f3f9b10`;

export default async function RateLimitsPage(props: DocPageProps) {
	const locale = await enterDocPage(props);
	const [t, nav] = await Promise.all([getTranslations("docs.limits"), getTranslations("api.docs.nav")]);
	const fmt = (n: number) => new Intl.NumberFormat(locale).format(n);
	return (
		<>
			<DocTitle lead={t("lead")}>{nav("limits")}</DocTitle>

			<H2 id="plans">{t("plansHeading")}</H2>
			<Table
				head={[t("colPlan"), t("colQuota"), t("colRate"), t("colPageSize"), t("colCredits")]}
				rows={PLAN_IDS.map((id) => {
					const p = PLAN_LIMITS[id];
					const quota = fmt(p.quota);
					return [
						<span key="n" className="font-semibold capitalize">
							{id}
						</span>,
						p.period === "lifetime" ? t("lifetimeQuota", { quota }) : t("monthlyQuota", { quota }),
						fmt(p.requestsPerSecond),
						fmt(p.maxPageSize),
						id === "free" ? "—" : t("creditPacks", { requests: fmt(CREDIT_PACK_REQUESTS) }),
					];
				})}
			/>
			<P>{t.rich("pricing", { link: richLink("/pricing") })}</P>

			<H2 id="counting">{t("countingHeading")}</H2>
			<ul className="my-3 list-disc space-y-2 pl-6">
				<li>{t("countEvery")}</li>
				<li>{t.rich("countNot", richCode)}</li>
				<li>{t.rich("countUsage", richCode)}</li>
				<li>{t("countReset")}</li>
			</ul>

			<H2 id="headers">{t("headersHeading")}</H2>
			<P>{t("headersBody")}</P>
			<Code title={t("headersTitle")}>{HEADERS}</Code>
			<Table
				head={[t("colHeader"), t("colMeaning")]}
				rows={[
					[<C key="h">RateLimit-Limit</C>, t("headerLimit")],
					[<C key="h">RateLimit-Remaining</C>, t("headerRemaining")],
					[<C key="h">RateLimit-Reset</C>, t("headerReset")],
					[<C key="h">X-Quota-Remaining</C>, t("headerQuota")],
					[<C key="h">Retry-After</C>, t("headerRetry")],
				]}
			/>

			<H2 id="exceeding">{t("exceedingHeading")}</H2>
			<Note>
				{t.rich("exceedingNote", {
					...richCode,
					rateLimited: richLink("/docs/errors#rate_limited"),
					paymentRequired: richLink("/docs/errors#payment_required"),
				})}
			</Note>
		</>
	);
}
