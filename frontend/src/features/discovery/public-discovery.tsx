"use client";

import { useEffect, useRef, useState } from "react";

type Listing = {
  id: string;
  title: string;
  description: string;
  categoryId: string;
  categorySlug: string;
  categoryName: string;
  primaryLocalityId: string;
  localitySlug: string;
  localityName: string;
  priceType: string;
  priceMinor: number | null;
  currency: "EUR";
  travelsToCustomer: boolean;
  receivesCustomer: boolean;
  remoteServices: boolean;
  providerDisplayName: string;
  providerType: string;
  updatedAt: string;
};

export type PublicDiscoveryCopy = {
  title: string;
  description: string;
  loading: string;
  empty: string;
  error: string;
  retry: string;
  searchLabel: string;
  searchButton: string;
  categoryLabel: string;
  localityLabel: string;
  radiusLabel: string;
  priceLabel: string;
  modeLabel: string;
  details: string;
};

export function PublicDiscovery({
  copy,
  locale,
}: {
  copy: PublicDiscoveryCopy;
  locale: "pt-PT" | "en" | "es";
}) {
  const [listings, setListings] = useState<Listing[] | null>(null);
  const [failed, setFailed] = useState(false);
  const [query, setQuery] = useState("");
  const generation = useRef(0);

  async function load(nextQuery = query) {
    const current = ++generation.current;
    setFailed(false);
    try {
      const values = await fetchListings(locale, nextQuery);
      if (current === generation.current) setListings(values);
    } catch {
      if (current === generation.current) setFailed(true);
    }
  }

  useEffect(() => {
    const current = ++generation.current;
    async function loadInitial() {
      try {
        const values = await fetchListings(locale, "");
        if (current === generation.current) setListings(values);
      } catch {
        if (current === generation.current) setFailed(true);
      }
    }
    void loadInitial();
    return () => {
      generation.current += 1;
    };
  }, [locale]);

  if (listings === null && !failed)
    return <p aria-live="polite">{copy.loading}</p>;
  if (failed && listings === null)
    return (
      <div role="alert">
        <p>{copy.error}</p>
        <button type="button" onClick={() => void load()}>
          {copy.retry}
        </button>
      </div>
    );

  return (
    <section aria-labelledby="discovery-title">
      <h1 id="discovery-title" className="text-3xl font-semibold">
        {copy.title}
      </h1>
      <p className="mt-3 text-muted">{copy.description}</p>
      <form
        className="mt-6 flex gap-3"
        onSubmit={(event) => {
          event.preventDefault();
          void load();
        }}
      >
        <label className="grid gap-1">
          <span className="text-sm">{copy.searchLabel}</span>
          <input
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            maxLength={80}
          />
        </label>
        <button
          type="submit"
          className="min-h-11 self-end rounded-full bg-ink px-5 text-canvas"
        >
          {copy.searchButton}
        </button>
      </form>
      {failed ? (
        <p role="alert" className="mt-4 text-earth">
          {copy.error}
        </p>
      ) : null}
      <div className="mt-8 grid gap-4">
        {listings?.length ? (
          listings.map((listing) => (
            <article
              key={listing.id}
              className="rounded-2xl border border-line p-5"
            >
              <p className="text-sm text-muted">
                {listing.categoryName} · {listing.localityName}
              </p>
              <h2 className="mt-2 text-xl font-semibold">{listing.title}</h2>
              <p className="mt-2 text-muted">{listing.description}</p>
              <p className="mt-3 text-sm">{listing.providerDisplayName}</p>
              <a
                className="mt-4 inline-flex min-h-11 items-center rounded-full border border-line px-4"
                href={`/${locale}/listings/${listing.id}`}
              >
                {copy.details}
              </a>
            </article>
          ))
        ) : (
          <p>{copy.empty}</p>
        )}
      </div>
    </section>
  );
}

async function fetchListings(
  locale: "pt-PT" | "en" | "es",
  query: string,
): Promise<Listing[]> {
  const params = new URLSearchParams({ locale });
  const normalized = query.trim().replace(/\s+/g, " ");
  if (normalized.length >= 2) params.set("q", normalized);
  const response = await fetch(`/api/v1/discovery/listings?${params}`);
  const value: unknown = await response.json();
  if (!response.ok || !validListings(value)) throw new Error();
  return value.listings;
}

function validListings(value: unknown): value is { listings: Listing[] } {
  return (
    exact(value, ["listings"]) &&
    Array.isArray(value.listings) &&
    value.listings.every(validListing)
  );
}

function validListing(value: unknown): value is Listing {
  if (
    !exact(value, [
      "id",
      "title",
      "description",
      "categoryId",
      "categorySlug",
      "categoryName",
      "primaryLocalityId",
      "localitySlug",
      "localityName",
      "priceType",
      "priceMinor",
      "currency",
      "travelsToCustomer",
      "receivesCustomer",
      "remoteServices",
      "providerDisplayName",
      "providerType",
      "updatedAt",
    ])
  )
    return false;
  const item = value as Record<string, unknown>;
  return (
    uuid(item.id) &&
    uuid(item.categoryId) &&
    uuid(item.primaryLocalityId) &&
    typeof item.title === "string" &&
    typeof item.description === "string" &&
    typeof item.categorySlug === "string" &&
    typeof item.categoryName === "string" &&
    typeof item.localitySlug === "string" &&
    typeof item.localityName === "string" &&
    ["fixed", "hourly", "daily", "quote", "negotiable"].includes(
      String(item.priceType),
    ) &&
    (item.priceMinor === null ||
      (Number.isInteger(item.priceMinor) && Number(item.priceMinor) > 0)) &&
    item.currency === "EUR" &&
    typeof item.travelsToCustomer === "boolean" &&
    typeof item.receivesCustomer === "boolean" &&
    typeof item.remoteServices === "boolean" &&
    typeof item.providerDisplayName === "string" &&
    ["individual", "professional", "business"].includes(
      String(item.providerType),
    ) &&
    typeof item.updatedAt === "string" &&
    !Number.isNaN(Date.parse(item.updatedAt))
  );
}
function exact(
  value: unknown,
  expected: string[],
): value is Record<string, unknown> {
  if (value === null || typeof value !== "object" || Array.isArray(value))
    return false;
  const actual = Object.keys(value).sort(),
    wanted = [...expected].sort();
  return (
    actual.length === wanted.length &&
    actual.every((key, index) => key === wanted[index])
  );
}
function uuid(value: unknown): value is string {
  return (
    typeof value === "string" &&
    /^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$/i.test(value)
  );
}
