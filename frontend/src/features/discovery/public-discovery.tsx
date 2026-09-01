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
  promoted: boolean;
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
  marketplaceLabel: string;
  locationContextLabel: string;
  filtersLabel: string;
  promoted: string;
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
    <section aria-labelledby="discovery-title" className="pb-12">
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-end">
        <div>
          <p className="text-xs font-bold tracking-[0.14em] text-accent uppercase">
            {copy.marketplaceLabel}
          </p>
          <h1
            id="discovery-title"
            className="mt-2 text-4xl font-bold tracking-[-0.05em] sm:text-5xl"
          >
            {copy.title}
          </h1>
          <p className="mt-3 max-w-2xl text-lg leading-8 text-muted">
            {copy.description}
          </p>
        </div>
        <span className="market-chip shrink-0">
          {copy.locationContextLabel}
        </span>
      </div>
      <form
        className="market-panel mt-8 flex flex-col gap-3 p-3 sm:flex-row sm:items-end"
        onSubmit={(event) => {
          event.preventDefault();
          void load();
        }}
      >
        <label className="grid flex-1 gap-1">
          <span className="px-1 text-sm font-semibold">{copy.searchLabel}</span>
          <input
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            maxLength={80}
            placeholder={copy.searchLabel}
            className="market-control w-full"
          />
        </label>
        <button
          type="submit"
          className="market-button w-full self-end sm:w-auto"
        >
          {copy.searchButton}
        </button>
      </form>
      {failed ? (
        <p role="alert" className="mt-4 text-earth">
          {copy.error}
        </p>
      ) : null}
      <div className="mt-6 flex flex-wrap gap-2" aria-label={copy.filtersLabel}>
        <span className="market-chip">{copy.categoryLabel}</span>
        <span className="market-chip">{copy.localityLabel}</span>
        <span className="market-chip">{copy.radiusLabel}</span>
        <span className="market-chip">{copy.priceLabel}</span>
        <span className="market-chip">{copy.modeLabel}</span>
      </div>
      <div className="mt-8 grid gap-4 md:grid-cols-2">
        {listings?.length ? (
          listings.map((listing) => (
            <article
              key={listing.id}
              className="market-card group overflow-hidden p-0"
            >
              <div className="flex min-h-24 items-end justify-between gap-3 bg-control p-5">
                <span className="market-chip bg-surface text-ink">
                  {listing.categoryName}
                </span>
                {listing.promoted ? (
                  <span className="market-chip">{copy.promoted}</span>
                ) : null}
              </div>
              <div className="p-5">
                <p className="text-sm font-medium text-muted">
                  {listing.localityName}
                </p>
                <h2 className="mt-2 text-xl font-bold tracking-[-0.03em]">
                  {listing.title}
                </h2>
                <p className="mt-2 line-clamp-2 text-sm leading-6 text-muted">
                  {listing.description}
                </p>
                <div className="mt-5 flex items-center justify-between gap-4 border-t border-line pt-4">
                  <p className="min-w-0 truncate text-sm font-semibold">
                    {listing.providerDisplayName}
                  </p>
                  <a
                    className="market-button-secondary shrink-0"
                    href={`/${locale}/listings/${listing.id}`}
                  >
                    {copy.details}
                  </a>
                </div>
              </div>
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
      "promoted",
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
    typeof item.promoted === "boolean" &&
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
