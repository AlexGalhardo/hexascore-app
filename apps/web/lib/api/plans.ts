export const PLAN_IDS = ["free", "pro", "business"] as const;
export type PlanId = (typeof PLAN_IDS)[number];

export type PlanLimits = {
	/** Requests per period; Free counts over the whole account lifetime. */
	quota: number;
	period: "lifetime" | "month";
	/** Max `limit` accepted on list endpoints. */
	maxPageSize: number;
	requestsPerSecond: number;
};

export const PLAN_LIMITS: Record<PlanId, PlanLimits> = {
	free: { quota: 100, period: "lifetime", maxPageSize: 10, requestsPerSecond: 1 },
	pro: { quota: 50_000, period: "month", maxPageSize: 100, requestsPerSecond: 10 },
	business: { quota: 500_000, period: "month", maxPageSize: 100, requestsPerSecond: 50 },
};

/** Extra requests per credit pack; only paid plans can buy and spend credits. */
export const CREDIT_PACK_REQUESTS = 25_000;
