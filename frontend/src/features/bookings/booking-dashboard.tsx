"use client";
import { FormEvent, useEffect, useRef, useState } from "react";
import type { Booking, BookingState } from "@/shared/api/generated";
import { AvailableListingSelect } from "@/features/listings/available-listing-select";
export type BookingsCopy = {
  title: string;
  description: string;
  newBooking: string;
  sourceType: string;
  proposal: string;
  listing: string;
  direct: string;
  sourceId: string;
  selectListing: string;
  loadingListings: string;
  emptyListings: string;
  providerId: string;
  scheduledAt: string;
  privateLocation: string;
  agreedPrice: string;
  create: string;
  loading: string;
  error: string;
  empty: string;
  price: string;
  confirm: string;
  schedule: string;
  start: string;
  complete: string;
  cancel: string;
  dispute: string;
  refund: string;
};
export function BookingDashboard({
  copy,
  locale,
}: {
  copy: BookingsCopy;
  locale: "pt-PT" | "en" | "es";
}) {
  const [items, setItems] = useState<Booking[]>([]),
    [loading, setLoading] = useState(true),
    [failed, setFailed] = useState(false),
    [creating, setCreating] = useState(false),
    [sourceType, setSourceType] = useState<"proposal" | "listing" | "direct">(
      "proposal",
    ),
    key = useRef(`booking-${crypto.randomUUID()}`);
  useEffect(() => {
    let active = true;
    void fetch("/api/v1/me/bookings")
      .then(async (r) => {
        const v: unknown = await r.json();
        if (!r.ok || !validList(v)) throw new Error();
        if (active) setItems(v.bookings);
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
  async function create(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const d = new FormData(e.currentTarget),
      sourceType = String(d.get("sourceType")),
      sourceId = String(d.get("sourceId") ?? "").trim(),
      providerId = String(d.get("providerId") ?? "").trim(),
      price = String(d.get("agreedPriceMinor") ?? "").trim();
    setFailed(false);
    try {
      const r = await fetch("/api/v1/me/bookings", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            sourceType,
            idempotencyKey: key.current,
            scheduledAt: new Date(String(d.get("scheduledAt"))).toISOString(),
            privateLocation: d.get("privateLocation"),
            ...(sourceId ? { sourceId } : {}),
            ...(providerId ? { providerId } : {}),
            ...(price ? { agreedPriceMinor: Number(price) } : {}),
          }),
        }),
        v: unknown = await r.json();
      if (!r.ok || !validBooking(v)) throw new Error();
      setItems((current) => [v, ...current.filter((i) => i.id !== v.id)]);
      key.current = `booking-${crypto.randomUUID()}`;
      e.currentTarget.reset();
      setSourceType("proposal");
      setCreating(false);
    } catch {
      setFailed(true);
    }
  }
  async function transition(item: Booking, targetState: BookingState) {
    setFailed(false);
    try {
      const r = await fetch(`/api/v1/me/bookings/${item.id}/transitions`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            expectedState: item.state,
            targetState,
            revision: item.revision,
          }),
        }),
        v: unknown = await r.json();
      if (!r.ok || !validBooking(v)) throw new Error();
      setItems((current) => current.map((i) => (i.id === v.id ? v : i)));
    } catch {
      setFailed(true);
    }
  }
  return (
    <section aria-labelledby="bookings-title">
      <div className="market-page-header">
        <h1
          id="bookings-title"
          className="text-4xl font-bold tracking-[-0.055em] sm:text-5xl"
        >
          {copy.title}
        </h1>
        <p>{copy.description}</p>
      </div>
      {loading ? <p className="market-empty mt-8">{copy.loading}</p> : null}
      {failed ? (
        <p role="alert" className="market-alert mt-6">
          {copy.error}
        </p>
      ) : null}
      <button
        className="market-button mt-6"
        type="button"
        onClick={() => setCreating((v) => !v)}
      >
        {copy.newBooking}
      </button>
      {creating ? (
        <form className="market-form-section mt-5" onSubmit={create}>
          <label className="grid gap-2 font-semibold">
            {copy.sourceType}
            <select
              className="market-control"
              name="sourceType"
              value={sourceType}
              onChange={(event) =>
                setSourceType(
                  event.target.value as "proposal" | "listing" | "direct",
                )
              }
            >
              <option value="proposal">{copy.proposal}</option>
              <option value="listing">{copy.listing}</option>
              <option value="direct">{copy.direct}</option>
            </select>
          </label>
          {sourceType === "listing" ? (
            <AvailableListingSelect
              emptyLabel={copy.emptyListings}
              label={copy.selectListing}
              loadingLabel={copy.loadingListings}
              locale={locale}
              name="sourceId"
              placeholder={copy.selectListing}
              scope="public"
            />
          ) : sourceType === "proposal" ? (
            <Field name="sourceId" label={copy.sourceId} required />
          ) : (
            <Field name="providerId" label={copy.providerId} required />
          )}
          <Field
            name="scheduledAt"
            label={copy.scheduledAt}
            type="datetime-local"
            required
          />
          <Field name="privateLocation" label={copy.privateLocation} required />
          <Field
            name="agreedPriceMinor"
            label={copy.agreedPrice}
            type="number"
          />
          <button className="market-button justify-self-start" type="submit">
            {copy.create}
          </button>
        </form>
      ) : null}
      <div className="mt-8 grid gap-4 xl:grid-cols-2">
        {!loading && items.length === 0 ? (
          <p className="market-empty xl:col-span-2">{copy.empty}</p>
        ) : (
          items.map((item) => (
            <article className="market-card p-5 sm:p-6" key={item.id}>
              <div className="flex flex-wrap items-start justify-between gap-4">
                <div>
                  <span className="market-chip text-ink">{item.state}</span>
                  <time
                    className="mt-1 block text-sm text-muted"
                    dateTime={item.scheduledAt}
                  >
                    {new Date(item.scheduledAt).toLocaleString()}
                  </time>
                  <p className="mt-2 text-sm text-muted">
                    {item.privateLocation}
                  </p>
                  <p className="mt-2 font-semibold">
                    {copy.price}: € {(item.agreedPriceMinor / 100).toFixed(2)}
                  </p>
                </div>
                <div className="flex flex-wrap gap-2">
                  {nextStates(item.state).map((state) => (
                    <button
                      className="market-button-secondary"
                      type="button"
                      key={state}
                      onClick={() => void transition(item, state)}
                    >
                      {label(state, copy)}
                    </button>
                  ))}
                </div>
              </div>
            </article>
          ))
        )}
      </div>
    </section>
  );
}
function Field({
  name,
  label,
  type = "text",
  required = false,
}: {
  name: string;
  label: string;
  type?: string;
  required?: boolean;
}) {
  return (
    <label className="grid gap-2 font-semibold">
      {label}
      <input
        className="market-control"
        name={name}
        type={type}
        required={required}
        min={type === "number" ? 1 : undefined}
      />
    </label>
  );
}
function nextStates(state: BookingState): BookingState[] {
  switch (state) {
    case "pending_provider_confirmation":
      return ["confirmed", "cancelled"];
    case "confirmed":
      return ["scheduled", "cancelled", "disputed"];
    case "scheduled":
      return ["in_progress", "cancelled", "disputed"];
    case "in_progress":
      return ["completed", "disputed"];
    case "completed":
      return ["disputed"];
    case "disputed":
      return ["refunded"];
    default:
      return [];
  }
}
function label(s: BookingState, c: BookingsCopy) {
  return s === "confirmed"
    ? c.confirm
    : s === "scheduled"
      ? c.schedule
      : s === "in_progress"
        ? c.start
        : s === "completed"
          ? c.complete
          : s === "cancelled"
            ? c.cancel
            : s === "disputed"
              ? c.dispute
              : c.refund;
}
function record(v: unknown): v is Record<string, unknown> {
  return v !== null && typeof v === "object" && !Array.isArray(v);
}
function validBooking(v: unknown): v is Booking {
  return (
    record(v) &&
    typeof v.id === "string" &&
    typeof v.state === "string" &&
    Number.isInteger(v.revision) &&
    typeof v.scheduledAt === "string" &&
    typeof v.privateLocation === "string" &&
    Number.isInteger(v.agreedPriceMinor)
  );
}
function validList(v: unknown): v is { bookings: Booking[] } {
  return (
    record(v) && Array.isArray(v.bookings) && v.bookings.every(validBooking)
  );
}
