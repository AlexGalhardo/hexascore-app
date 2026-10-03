"use client";

import { Dialog } from "@base-ui/react/dialog";
import { Radio } from "@base-ui/react/radio";
import { RadioGroup } from "@base-ui/react/radio-group";
import { useQuery } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useLocale, useTranslations } from "next-intl";
import { type ReactNode, useEffect, useId, useState, useTransition } from "react";
import { z } from "zod";
import { createPixCharge, type PixChargeView, simulatePixPaymentAction, startStripeCheckout } from "@/actions/billing";
import { buttonClass } from "@/components/ui/form";
import { secondsLeft } from "@/lib/account-deletion";
import {
	CREDIT_PACK_PRICE_CENTS,
	CURRENCIES,
	type Currency,
	formatPrice,
	MONTHLY_PRICE_CENTS,
	type PixPurpose,
	pixOffer,
} from "@/lib/billing/catalog";

type Product = "pro" | "business" | "credits";
type Method = "card" | "pix";
type PixStatus = "pending" | "paid" | "expired" | "failed";

const pixStatusSchema = z.object({ status: z.enum(["pending", "paid", "expired", "failed"]) });
const POLL_MS = 4000;
const SANDBOX_COOLDOWN_SECONDS = 10;
const PAID_LOCK_SECONDS = 10;

// Seconds left of a `total`-second window that started at `startedAt` (0 when not started).
function useSecondsLeft(startedAt: number | null, total: number): number {
	const [now, setNow] = useState(() => Date.now());
	useEffect(() => {
		if (startedAt === null) return;
		setNow(Date.now());
		const timer = setInterval(() => {
			const current = Date.now();
			setNow(current);
			if (secondsLeft(startedAt, current, total) === 0) clearInterval(timer);
		}, 250);
		return () => clearInterval(timer);
	}, [startedAt, total]);
	// `now` can lag one render behind a fresh start, which would read as more than `total`.
	return startedAt === null ? 0 : Math.min(total, secondsLeft(startedAt, now, total));
}

type Props = {
	stripe: boolean;
	pix: boolean;
	/** Server-side ABACATEPAY_SANDBOX: shows the simulated payment button. */
	pixSandbox: boolean;
	defaultCurrency: Currency;
	defaultProduct: Product;
	hasPaidPlan: boolean;
};

function Choices<T extends string>(props: {
	label: string;
	value: T;
	onChange: (value: T) => void;
	options: { value: T; label: ReactNode; disabled?: boolean }[];
}) {
	const id = useId();
	return (
		<fieldset className="space-y-2">
			<legend id={id} className="text-sm font-medium">
				{props.label}
			</legend>
			<RadioGroup
				aria-labelledby={id}
				value={props.value}
				onValueChange={(v) => props.onChange(v as T)}
				className="flex flex-wrap gap-2"
			>
				{props.options.map((o) => (
					// biome-ignore lint/a11y/noLabelWithoutControl: Base UI Radio.Root renders the control inside the label
					<label
						key={o.value}
						className="flex min-h-11 cursor-pointer items-center gap-2 rounded-md border border-border px-3 text-sm has-[[data-checked]]:border-fg has-[[data-disabled]]:cursor-not-allowed has-[[data-disabled]]:opacity-50"
					>
						<Radio.Root
							value={o.value}
							disabled={o.disabled}
							className="flex size-4 items-center justify-center rounded-full border border-fg"
						>
							<Radio.Indicator className="size-2 rounded-full bg-fg" />
						</Radio.Root>
						{o.label}
					</label>
				))}
			</RadioGroup>
		</fieldset>
	);
}

export function CheckoutForm({ stripe, pix, pixSandbox, defaultCurrency, defaultProduct, hasPaidPlan }: Props) {
	const t = useTranslations("billing.checkout");
	const tp = useTranslations("billing.plans");
	const locale = useLocale();
	const router = useRouter();
	const [product, setProduct] = useState<Product>(
		defaultProduct === "credits" && !hasPaidPlan ? "pro" : defaultProduct,
	);
	const [method, setMethod] = useState<Method>(stripe ? "card" : "pix");
	const [months, setMonths] = useState<"1" | "12">("1");
	const [currency, setCurrency] = useState<Currency>(defaultCurrency);
	const [pending, start] = useTransition();
	const [error, setError] = useState<string | null>(null);
	const [charge, setCharge] = useState<PixChargeView | null>(null);
	// Keyed by charge id, so a new charge starts again at "pending"; polling stops once the status is final.
	const pixQuery = useQuery({
		queryKey: ["pix-charge", charge?.id ?? null],
		queryFn: async ({ signal }): Promise<PixStatus> => {
			const response = await fetch(`/api/billing/pix/${charge?.id}`, {
				cache: "no-store",
				signal: AbortSignal.any([signal, AbortSignal.timeout(POLL_MS)]),
			});
			if (!response.ok) throw new Error(`HTTP ${response.status}`);
			return pixStatusSchema.parse(await response.json()).status;
		},
		enabled: charge !== null,
		refetchInterval: (query) => ((query.state.data ?? "pending") === "pending" ? POLL_MS : false),
		// Network blips are retried on the next tick.
		retry: false,
	});
	const pixStatus: PixStatus = pixQuery.data ?? "pending";
	const [copied, setCopied] = useState(false);
	const [simulating, startSimulate] = useTransition();
	const [simulatedAt, setSimulatedAt] = useState<number | null>(null);
	const [simulateError, setSimulateError] = useState<string | null>(null);
	const [paidAt, setPaidAt] = useState<number | null>(null);
	const cooldownLeft = useSecondsLeft(simulatedAt, SANDBOX_COOLDOWN_SECONDS);
	const lockLeft = useSecondsLeft(paidAt, PAID_LOCK_SECONDS);
	const locked = pixStatus === "paid" && lockLeft > 0;

	const purpose: PixPurpose = product === "credits" ? "credits" : `${product}_${months}m`;
	// Display only: the server recomputes every amount from the same catalog.
	const total =
		method === "pix"
			? formatPrice(pixOffer(purpose).amountCents, "brl", locale)
			: formatPrice(
					product === "credits" ? CREDIT_PACK_PRICE_CENTS[currency] : MONTHLY_PRICE_CENTS[product][currency],
					currency,
					locale,
				);

	useEffect(() => {
		if (pixStatus !== "paid") return;
		setPaidAt(Date.now());
		router.refresh();
	}, [pixStatus, router]);

	function submit(): void {
		setError(null);
		start(async () => {
			if (method === "card") {
				const result = await startStripeCheckout({ product, currency });
				if (result.ok) window.location.assign(result.url);
				else setError(t(`errors.${result.error}`));
				return;
			}
			const result = await createPixCharge({ purpose });
			if (result.ok) {
				setCharge(result.charge);
				setCopied(false);
				setPaidAt(null);
				setSimulatedAt(null);
				setSimulateError(null);
			} else setError(t(`errors.${result.error}`));
		});
	}

	function simulatePayment(): void {
		if (!charge || simulating || cooldownLeft > 0) return;
		setSimulateError(null);
		setSimulatedAt(Date.now());
		startSimulate(async () => {
			const result = await simulatePixPaymentAction(charge.id);
			if (!result.ok) setSimulateError(t(`errors.${result.error}`));
		});
	}

	function closeCharge(): void {
		if (!locked) setCharge(null);
	}

	async function copy(): Promise<void> {
		if (!charge) return;
		try {
			await navigator.clipboard.writeText(charge.brCode);
			setCopied(true);
		} catch {
			setCopied(false);
		}
	}

	const pixDialog = (
		<Dialog.Root
			open={charge !== null}
			onOpenChange={(open) => {
				if (!open) closeCharge();
			}}
			disablePointerDismissal={locked}
		>
			<Dialog.Portal>
				<Dialog.Backdrop className="fixed inset-0 bg-black/40" />
				<Dialog.Popup className="fixed inset-x-2 top-[5dvh] mx-auto grid max-h-[90dvh] max-w-md gap-4 overflow-y-auto rounded-lg border border-border bg-bg p-5 text-fg shadow-xl">
					{charge && renderCharge(charge)}
					<Dialog.Close disabled={locked} className={buttonClass("secondary")}>
						{t("close")}
					</Dialog.Close>
				</Dialog.Popup>
			</Dialog.Portal>
		</Dialog.Root>
	);

	function renderCharge(charge: PixChargeView): ReactNode {
		const expires = new Intl.DateTimeFormat(locale, { timeStyle: "short" }).format(new Date(charge.expiresAt));
		return (
			<>
				<Dialog.Title className="font-serif text-2xl font-bold">{t("pixTitle")}</Dialog.Title>
				<p className="font-mono text-xl tabular-nums">{formatPrice(charge.amountCents, "brl", locale)}</p>
				<p role="status" aria-live="polite" className="text-sm font-medium">
					{t(`pixStatus.${pixStatus}`)}
					{locked && (
						<span className="mt-1 block font-normal text-muted">{t("closeIn", { seconds: lockLeft })}</span>
					)}
				</p>
				{pixStatus === "pending" && (
					<>
						<p className="text-sm text-muted">{t("pixScan")}</p>
						{/* biome-ignore lint/performance/noImgElement: data: URI from AbacatePay, nothing for next/image to optimize */}
						<img
							src={charge.qrImage}
							alt={t("pixCode")}
							width={240}
							height={240}
							className="size-60 rounded-md bg-white p-2"
						/>
						<label className="block space-y-1 text-sm">
							<span>{t("pixCode")}</span>
							<textarea
								readOnly
								value={charge.brCode}
								rows={3}
								className="w-full rounded-md border border-border bg-bg p-2 font-mono text-xs break-all"
								onFocus={(e) => e.currentTarget.select()}
							/>
						</label>
						<div className="flex flex-wrap items-center gap-3">
							<button
								type="button"
								onClick={copy}
								className="min-h-11 rounded-md bg-fg px-4 text-sm text-bg"
							>
								{copied ? t("copied") : t("copy")}
							</button>
							<span className="text-sm text-muted">{t("expires", { time: expires })}</span>
						</div>
						{pixSandbox && (
							<div className="space-y-2 rounded-md border border-dashed border-border p-3">
								<p className="text-xs text-muted">{t("sandboxNote")}</p>
								<button
									type="button"
									onClick={simulatePayment}
									disabled={simulating || cooldownLeft > 0}
									aria-describedby={simulateError ? "pix-sandbox-error" : undefined}
									className={buttonClass("primary", "w-full font-mono tracking-wide")}
								>
									{cooldownLeft > 0 ? t("sandboxWait", { seconds: cooldownLeft }) : t("sandboxPay")}
								</button>
								{simulateError && (
									<p id="pix-sandbox-error" role="alert" className="text-sm text-danger">
										{simulateError}
									</p>
								)}
							</div>
						)}
					</>
				)}
				{(pixStatus === "expired" || pixStatus === "failed") && (
					<button
						type="button"
						onClick={() => setCharge(null)}
						className="min-h-11 rounded-md border border-border px-4 text-sm"
					>
						{t("newCode")}
					</button>
				)}
			</>
		);
	}

	return (
		<>
			{pixDialog}
			<form
				className="space-y-5 rounded-lg border border-border p-5"
				onSubmit={(e) => {
					e.preventDefault();
					submit();
				}}
			>
				<h2 className="font-serif text-2xl font-bold">{t("buy")}</h2>
				<Choices<Product>
					label={t("product")}
					value={product}
					onChange={setProduct}
					options={[
						{ value: "pro", label: tp("pro") },
						{ value: "business", label: tp("business") },
						{ value: "credits", label: tp("credits"), disabled: !hasPaidPlan },
					]}
				/>
				<Choices<Method>
					label={t("method")}
					value={method}
					onChange={setMethod}
					options={[
						{ value: "card", label: t("card"), disabled: !stripe },
						{ value: "pix", label: t("pix"), disabled: !pix },
					]}
				/>
				{method === "pix" && product !== "credits" && (
					<Choices<"1" | "12">
						label={t("period")}
						value={months}
						onChange={setMonths}
						options={[
							{ value: "1", label: t("oneMonth") },
							{ value: "12", label: t("twelveMonths") },
						]}
					/>
				)}
				{method === "card" && (
					<Choices<Currency>
						label={t("currency")}
						value={currency}
						onChange={setCurrency}
						options={CURRENCIES.map((c) => ({
							value: c,
							label: <span className="font-mono">{c.toUpperCase()}</span>,
						}))}
					/>
				)}
				<p className="flex items-baseline justify-between border-t border-border pt-4">
					<span>{t("total")}</span>
					<span className="font-mono text-2xl tabular-nums">{total}</span>
				</p>
				<p className="text-sm text-muted">
					{method === "card" && product !== "credits" ? t("recurring") : t("oneTime")}
				</p>
				{error && (
					<p role="alert" className="text-sm text-danger">
						{error}
					</p>
				)}
				<button
					type="submit"
					disabled={pending}
					className="min-h-11 w-full rounded-md bg-fg px-4 text-bg disabled:opacity-50 sm:w-auto"
				>
					{pending ? t("working") : method === "pix" ? t("payPix") : t("pay")}
				</button>
			</form>
		</>
	);
}
