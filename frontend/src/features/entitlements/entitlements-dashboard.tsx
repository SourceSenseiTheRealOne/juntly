"use client";

import { FormEvent, useEffect, useState } from "react";
import type {
  EntitlementCatalog,
  MyEntitlements,
} from "@/shared/api/generated";
import {
  validCatalog,
  validEntitlements,
} from "@/features/entitlements/entitlement-bff";
import { AvailableListingSelect } from "@/features/listings/available-listing-select";

export type EntitlementsCopy = {
  title: string;
  description: string;
  access: string;
  activeListings: string;
  photos: string;
  analytics: string;
  enabled: string;
  disabled: string;
  plans: string;
  choosePlan: string;
  promotion: string;
  listingId: string;
  selectListing: string;
  loadingListings: string;
  emptyListings: string;
  period: string;
  promote: string;
  current: string;
  none: string;
  loading: string;
  error: string;
  requested: string;
  pending: string;
  active: string;
  cancelled: string;
  expired: string;
};
export function EntitlementsDashboard({
  copy,
  locale,
}: {
  copy: EntitlementsCopy;
  locale: "pt-PT" | "en" | "es";
}) {
  const [catalog, setCatalog] = useState<EntitlementCatalog | null>(null),
    [mine, setMine] = useState<MyEntitlements | null>(null),
    [loading, setLoading] = useState(true),
    [failed, setFailed] = useState(false),
    [notice, setNotice] = useState("");
  useEffect(() => {
    let active = true;
    void Promise.all([
      fetch("/api/v1/entitlements/catalog"),
      fetch("/api/v1/me/entitlements"),
    ])
      .then(async ([catalogResponse, mineResponse]) => {
        const nextCatalog: unknown = await catalogResponse.json(),
          nextMine: unknown = await mineResponse.json();
        if (
          !catalogResponse.ok ||
          !mineResponse.ok ||
          !validCatalog(nextCatalog) ||
          !validEntitlements(nextMine)
        )
          throw new Error();
        if (active) {
          setCatalog(nextCatalog);
          setMine(nextMine);
        }
      })
      .catch(() => {
        if (active) setFailed(true);
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);
  async function subscribe(planId: string) {
    setFailed(false);
    setNotice("");
    try {
      const response = await fetch("/api/v1/me/subscriptions", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ planId }),
      });
      if (!response.ok) throw new Error();
      setNotice(copy.requested);
    } catch {
      setFailed(true);
    }
  }
  async function promote(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!catalog) return;
    const data = new FormData(event.currentTarget);
    setFailed(false);
    setNotice("");
    try {
      const response = await fetch("/api/v1/me/promotions", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          listingId: data.get("listingId"),
          periodId: data.get("periodId"),
        }),
      });
      if (!response.ok) throw new Error();
      event.currentTarget.reset();
      setNotice(copy.requested);
    } catch {
      setFailed(true);
    }
  }
  const status = (value: string) =>
    value === "pending"
      ? copy.pending
      : value === "active"
        ? copy.active
        : value === "cancelled"
          ? copy.cancelled
          : copy.expired;
  if (loading) return <p className="market-empty">{copy.loading}</p>;
  return (
    <section aria-labelledby="entitlements-title">
      <div className="market-page-header">
        <h1
          id="entitlements-title"
          className="text-4xl font-bold tracking-[-0.055em] sm:text-5xl"
        >
          {copy.title}
        </h1>
        <p>{copy.description}</p>
      </div>
      {failed ? (
        <p className="market-alert mt-6" role="alert">
          {copy.error}
        </p>
      ) : null}
      {notice ? (
        <p className="market-success mt-6 font-semibold" aria-live="polite">
          {notice}
        </p>
      ) : null}
      {mine ? (
        <div className="market-card market-data-grid mt-8 p-5 sm:p-6">
          <h2 className="text-xl font-bold sm:col-span-full">{copy.access}</h2>
          <Metric
            label={copy.activeListings}
            value={String(mine.access.maxActiveListings)}
          />
          <Metric
            label={copy.photos}
            value={String(mine.access.maxPhotosPerListing)}
          />
          <Metric
            label={copy.analytics}
            value={mine.access.analyticsEnabled ? copy.enabled : copy.disabled}
          />
          <p className="border-t border-line pt-4 text-sm text-muted sm:col-span-full">
            {copy.current}:{" "}
            {mine.subscription ? status(mine.subscription.status) : copy.none}
          </p>
        </div>
      ) : null}
      <h2 className="mt-10 text-2xl font-bold">{copy.plans}</h2>
      <div className="mt-4 grid gap-4 md:grid-cols-2">
        {catalog?.plans.map((plan) => (
          <article className="market-card p-5" key={plan.id}>
            <h3 className="text-xl font-bold">{plan.name}</h3>
            <p className="mt-2 text-muted">
              {(plan.priceMinor / 100).toLocaleString(undefined, {
                style: "currency",
                currency: plan.currency,
              })}{" "}
              / {plan.billingDays}
            </p>
            <p className="mt-3 text-sm text-muted">
              {copy.activeListings}: {plan.maxActiveListings}. {copy.photos}:{" "}
              {plan.maxPhotosPerListing}.
            </p>
            <button
              className="market-button mt-5"
              type="button"
              onClick={() => void subscribe(plan.id)}
            >
              {copy.choosePlan}
            </button>
          </article>
        ))}
      </div>
      <form className="market-form-section mt-10" onSubmit={promote}>
        <h2 className="text-2xl font-bold">{copy.promotion}</h2>
        <AvailableListingSelect
          emptyLabel={copy.emptyListings}
          label={copy.listingId}
          loadingLabel={copy.loadingListings}
          locale={locale}
          name="listingId"
          placeholder={copy.selectListing}
          scope="mine"
        />
        <label className="grid gap-2 font-semibold">
          {copy.period}
          <select className="market-control" name="periodId" required>
            {catalog?.promotionPeriods.map((period) => (
              <option value={period.id} key={period.id}>
                {period.name} -{" "}
                {(period.priceMinor / 100).toLocaleString(undefined, {
                  style: "currency",
                  currency: period.currency,
                })}
              </option>
            ))}
          </select>
        </label>
        <button className="market-button justify-self-start" type="submit">
          {copy.promote}
        </button>
      </form>
    </section>
  );
}
function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="text-sm text-muted">{label}</p>
      <p className="mt-1 text-2xl font-bold">{value}</p>
    </div>
  );
}
