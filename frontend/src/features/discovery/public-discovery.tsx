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

type CategoryOption = { id: string; name: string };
type LocalityOption = { id: string; name: string };
type DiscoveryFilters = {
  categoryId: string;
  nearLocalityId: string;
  radiusKm: string;
  priceType: string;
  serviceMode: string;
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
  allCategories: string;
  allLocalities: string;
  anyPrice: string;
  anyMode: string;
  priceFixed: string;
  priceHourly: string;
  priceDaily: string;
  priceQuote: string;
  priceNegotiable: string;
  modeTravels: string;
  modeReceives: string;
  modeRemote: string;
  applyFilters: string;
};

export function PublicDiscovery({
  copy,
  locale,
}: {
  copy: PublicDiscoveryCopy;
  locale: "pt-PT" | "en" | "es";
}) {
  const [listings, setListings] = useState<Listing[] | null>(null);
  const [categories, setCategories] = useState<CategoryOption[]>([]);
  const [localities, setLocalities] = useState<LocalityOption[]>([]);
  const [failed, setFailed] = useState(false);
  const [query, setQuery] = useState("");
  const [filters, setFilters] = useState<DiscoveryFilters>({
    categoryId: "",
    nearLocalityId: "",
    radiusKm: "25",
    priceType: "",
    serviceMode: "",
  });
  const generation = useRef(0);

  async function load(nextQuery = query, nextFilters = filters) {
    const current = ++generation.current;
    setFailed(false);
    try {
      const values = await fetchListings(locale, nextQuery, nextFilters);
      if (current === generation.current) setListings(values);
    } catch {
      if (current === generation.current) setFailed(true);
    }
  }

  useEffect(() => {
    const current = ++generation.current;
    async function loadInitial() {
      try {
        const [values, references] = await Promise.all([
          fetchListings(locale, "", {
            categoryId: "",
            nearLocalityId: "",
            radiusKm: "25",
            priceType: "",
            serviceMode: "",
          }),
          fetchReferences(locale),
        ]);
        if (current === generation.current) {
          setListings(values);
          setCategories(references.categories);
          setLocalities(references.localities);
        }
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
    <section aria-labelledby="discovery-title" className="pb-16">
      <div className="grid gap-6 border-b border-line pb-8 sm:grid-cols-[1fr_auto] sm:items-end">
        <div className="market-page-header">
          <p className="market-kicker">{copy.marketplaceLabel}</p>
          <h1
            id="discovery-title"
            className="text-4xl font-bold tracking-[-0.055em] sm:text-5xl"
          >
            {copy.title}
          </h1>
          <p>{copy.description}</p>
        </div>
        <span className="text-sm font-semibold text-muted sm:pb-1">
          {copy.locationContextLabel}
        </span>
      </div>
      <form
        aria-label={copy.filtersLabel}
        className="market-toolbar mt-6 grid gap-3 p-4 sm:grid-cols-2 lg:grid-cols-6 lg:items-end"
        onSubmit={(event) => {
          event.preventDefault();
          void load();
        }}
      >
        <label className="grid gap-1 sm:col-span-2 lg:col-span-2">
          <span className="px-1 text-sm font-semibold text-ink">
            {copy.searchLabel}
          </span>
          <input
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            maxLength={80}
            placeholder={copy.searchLabel}
            className="market-control w-full"
          />
        </label>
        <FilterSelect
          label={copy.categoryLabel}
          value={filters.categoryId}
          onChange={(categoryId) => setFilters({ ...filters, categoryId })}
          options={categories.map((item) => ({
            value: item.id,
            label: item.name,
          }))}
          placeholder={copy.allCategories}
        />
        <FilterSelect
          label={copy.localityLabel}
          value={filters.nearLocalityId}
          onChange={(nearLocalityId) =>
            setFilters({ ...filters, nearLocalityId })
          }
          options={localities.map((item) => ({
            value: item.id,
            label: item.name,
          }))}
          placeholder={copy.allLocalities}
        />
        <FilterSelect
          disabled={!filters.nearLocalityId}
          label={copy.radiusLabel}
          value={filters.radiusKm}
          onChange={(radiusKm) => setFilters({ ...filters, radiusKm })}
          options={[5, 10, 25, 50, 100, 200].map((radius) => ({
            value: String(radius),
            label: `${radius} km`,
          }))}
          placeholder={copy.radiusLabel}
        />
        <FilterSelect
          label={copy.priceLabel}
          value={filters.priceType}
          onChange={(priceType) => setFilters({ ...filters, priceType })}
          options={[
            { value: "fixed", label: copy.priceFixed },
            { value: "hourly", label: copy.priceHourly },
            { value: "daily", label: copy.priceDaily },
            { value: "quote", label: copy.priceQuote },
            { value: "negotiable", label: copy.priceNegotiable },
          ]}
          placeholder={copy.anyPrice}
        />
        <FilterSelect
          label={copy.modeLabel}
          value={filters.serviceMode}
          onChange={(serviceMode) => setFilters({ ...filters, serviceMode })}
          options={[
            { value: "travels_to_customer", label: copy.modeTravels },
            { value: "receives_customer", label: copy.modeReceives },
            { value: "remote_services", label: copy.modeRemote },
          ]}
          placeholder={copy.anyMode}
        />
        <button type="submit" className="market-button w-full lg:col-start-6">
          {copy.applyFilters}
        </button>
      </form>
      {failed ? (
        <p role="alert" className="market-alert mt-4">
          {copy.error}
        </p>
      ) : null}

      <div className="mt-6 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        {listings?.length ? (
          listings.map((listing) => (
            <article
              key={listing.id}
              className="market-card group flex min-h-64 flex-col overflow-hidden p-0"
            >
              <div className="flex items-center justify-between gap-3 border-b border-line bg-control/70 px-5 py-3">
                <span className="text-xs font-bold tracking-[0.08em] text-muted uppercase">
                  {listing.categoryName}
                </span>
                {listing.promoted ? (
                  <span className="market-chip border-accent/20 bg-accent-soft text-accent">
                    {copy.promoted}
                  </span>
                ) : null}
              </div>
              <div className="flex flex-1 flex-col p-5">
                <p className="text-sm font-medium text-muted">
                  {listing.localityName}
                </p>
                <h2 className="mt-2 text-xl font-bold tracking-[-0.035em] group-hover:text-accent">
                  {listing.title}
                </h2>
                <p className="mt-2 line-clamp-2 text-sm leading-6 text-muted">
                  {listing.description}
                </p>
                <div className="mt-auto flex items-center justify-between gap-4 pt-6">
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
          <p className="market-empty md:col-span-2 xl:col-span-3">
            {copy.empty}
          </p>
        )}
      </div>
    </section>
  );
}

async function fetchListings(
  locale: "pt-PT" | "en" | "es",
  query: string,
  filters: DiscoveryFilters,
): Promise<Listing[]> {
  const params = new URLSearchParams({ locale });
  const normalized = query.trim().replace(/\s+/g, " ");
  if (normalized.length >= 2) params.set("q", normalized);
  if (filters.categoryId) params.set("categoryId", filters.categoryId);
  if (filters.nearLocalityId) {
    params.set("nearLocalityId", filters.nearLocalityId);
    params.set("radiusKm", filters.radiusKm);
  }
  if (filters.priceType) params.set("priceType", filters.priceType);
  if (filters.serviceMode) params.set("serviceMode", filters.serviceMode);
  const response = await fetch(`/api/v1/discovery/listings?${params}`);
  const value: unknown = await response.json();
  if (!response.ok || !validListings(value)) throw new Error();
  return value.listings;
}

async function fetchReferences(locale: "pt-PT" | "en" | "es") {
  const query = `?locale=${encodeURIComponent(locale)}`;
  const [categoryResponse, localityResponse] = await Promise.all([
    fetch(`/api/v1/catalog/categories${query}`),
    fetch(`/api/v1/reference/localities${query}`),
  ]);
  const categories: unknown = await categoryResponse.json();
  const localities: unknown = await localityResponse.json();
  if (
    !categoryResponse.ok ||
    !localityResponse.ok ||
    !validOptions(categories, "categories") ||
    !validOptions(localities, "localities")
  )
    throw new Error();
  return {
    categories: categories.categories,
    localities: localities.localities,
  };
}

function FilterSelect({
  disabled = false,
  label,
  onChange,
  options,
  placeholder,
  value,
}: {
  disabled?: boolean;
  label: string;
  onChange: (value: string) => void;
  options: Array<{ value: string; label: string }>;
  placeholder: string;
  value: string;
}) {
  return (
    <label className="grid gap-1">
      <span className="px-1 text-sm font-semibold text-ink">{label}</span>
      <select
        className="market-control w-full"
        disabled={disabled}
        value={value}
        onChange={(event) => onChange(event.target.value)}
      >
        <option value="">{placeholder}</option>
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
    </label>
  );
}

function validOptions(
  value: unknown,
  key: "categories" | "localities",
): value is Record<
  "categories" | "localities",
  Array<{ id: string; name: string }>
> {
  return (
    value !== null &&
    typeof value === "object" &&
    !Array.isArray(value) &&
    Array.isArray((value as Record<string, unknown>)[key]) &&
    ((value as Record<string, unknown>)[key] as unknown[]).every(
      (item) =>
        item !== null &&
        typeof item === "object" &&
        !Array.isArray(item) &&
        uuid((item as Record<string, unknown>).id) &&
        typeof (item as Record<string, unknown>).name === "string",
    )
  );
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
