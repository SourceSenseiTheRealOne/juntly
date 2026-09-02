"use client";

import { useEffect, useRef, useState } from "react";

type Listing = {
  id: string;
  title: string;
  description: string;
  state: string;
  revision: number;
};

export type ModerationQueueCopy = {
  title: string;
  loading: string;
  empty: string;
  error: string;
  retry: string;
  approve: string;
  reject: string;
};

export function ModerationQueue({ copy }: { copy: ModerationQueueCopy }) {
  const [listings, setListings] = useState<Listing[] | null>(null);
  const [failed, setFailed] = useState(false);
  const [saving, setSaving] = useState(false);
  const [reason, setReason] = useState("");
  const generation = useRef(0);

  async function load() {
    const current = ++generation.current;
    setFailed(false);
    try {
      const response = await fetch("/api/v1/moderation/listings");
      const value: unknown = await response.json();
      if (!response.ok || !validList(value)) throw new Error();
      if (current === generation.current) setListings(value.listings);
    } catch {
      if (current === generation.current) setFailed(true);
    }
  }

  useEffect(() => {
    const current = ++generation.current;
    async function loadInitialQueue() {
      try {
        const response = await fetch("/api/v1/moderation/listings");
        const value: unknown = await response.json();
        if (!response.ok || !validList(value)) throw new Error();
        if (current === generation.current) setListings(value.listings);
      } catch {
        if (current === generation.current) setFailed(true);
      }
    }
    void loadInitialQueue();
    return () => {
      generation.current += 1;
    };
  }, []);

  async function approve(item: Listing) {
    if (saving) return;
    const current = ++generation.current;
    setSaving(true);
    setFailed(false);
    try {
      const response = await fetch(
        `/api/v1/moderation/listings/${item.id}/approve`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ revision: item.revision }),
        },
      );
      const value: unknown = await response.json();
      if (!response.ok || !validListing(value)) throw new Error();
      if (current === generation.current)
        setListings(
          (previous) =>
            previous?.filter((listing) => listing.id !== item.id) ?? null,
        );
    } catch {
      if (current === generation.current) setFailed(true);
    } finally {
      if (current === generation.current) setSaving(false);
    }
  }

  async function reject(item: Listing) {
    const normalized = reason.trim();
    if (saving || !normalized || normalized.length > 500) return;
    const current = ++generation.current;
    setSaving(true);
    setFailed(false);
    try {
      const response = await fetch(
        `/api/v1/moderation/listings/${item.id}/reject`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ revision: item.revision, reason: normalized }),
        },
      );
      const value: unknown = await response.json();
      if (!response.ok || !validListing(value)) throw new Error();
      if (current === generation.current) {
        setListings(
          (previous) =>
            previous?.filter((listing) => listing.id !== item.id) ?? null,
        );
        setReason("");
      }
    } catch {
      if (current === generation.current) setFailed(true);
    } finally {
      if (current === generation.current) setSaving(false);
    }
  }

  if (listings === null && !failed)
    return (
      <p className="market-empty" aria-live="polite">
        {copy.loading}
      </p>
    );
  if (failed && listings === null)
    return (
      <div className="market-alert grid justify-items-start gap-3" role="alert">
        <p>{copy.error}</p>
        <button
          className="market-button-secondary"
          type="button"
          onClick={() => void load()}
        >
          {copy.retry}
        </button>
      </div>
    );

  return (
    <section aria-labelledby="moderation-title">
      <h1
        id="moderation-title"
        className="text-4xl font-bold tracking-[-0.055em] sm:text-5xl"
      >
        {copy.title}
      </h1>
      {failed ? (
        <p className="market-alert mt-5" role="alert">
          {copy.error}
        </p>
      ) : null}
      <div className="mt-6 grid gap-4">
        {listings?.length ? (
          listings.map((item) => (
            <article key={item.id} className="market-card p-5">
              <h2 className="font-semibold">{item.title}</h2>
              <p className="mt-2 text-muted">{item.description}</p>
              <span className="market-chip mt-3">{item.state}</span>
              <button
                disabled={saving}
                type="button"
                className="market-button mt-3"
                onClick={() => void approve(item)}
              >
                {copy.approve}
              </button>
              <label className="mt-3 grid gap-2 text-sm">
                {copy.reject}
                <input
                  value={reason}
                  maxLength={500}
                  onChange={(event) => setReason(event.target.value)}
                  className="market-control"
                />
              </label>
              <button
                disabled={saving || !reason.trim()}
                type="button"
                className="market-button-secondary mt-3"
                onClick={() => void reject(item)}
              >
                {copy.reject}
              </button>
            </article>
          ))
        ) : (
          <p className="market-empty">{copy.empty}</p>
        )}
      </div>
    </section>
  );
}

function exact(
  value: unknown,
  keys: string[],
): value is Record<string, unknown> {
  if (value === null || typeof value !== "object" || Array.isArray(value))
    return false;
  const actual = Object.keys(value).sort();
  const expected = [...keys].sort();
  return (
    actual.length === expected.length &&
    actual.every((key, index) => key === expected[index])
  );
}

function validList(value: unknown): value is { listings: Listing[] } {
  return (
    exact(value, ["listings"]) &&
    Array.isArray(value.listings) &&
    value.listings.every(validListing)
  );
}

function validListing(value: unknown): value is Listing {
  return (
    exact(value, [
      "id",
      "categoryId",
      "primaryLocalityId",
      "title",
      "description",
      "priceType",
      "priceMinor",
      "currency",
      "travelsToCustomer",
      "receivesCustomer",
      "remoteServices",
      "state",
      "revision",
      "createdAt",
      "updatedAt",
    ]) &&
    typeof value.id === "string" &&
    typeof value.title === "string" &&
    typeof value.description === "string" &&
    typeof value.state === "string" &&
    Number.isInteger(value.revision)
  );
}
