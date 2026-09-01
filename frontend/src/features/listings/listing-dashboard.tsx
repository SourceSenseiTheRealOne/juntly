"use client";
import { useEffect, useRef, useState } from "react";
import type { FormEvent } from "react";
type Ref = { id: string; name: string };
type Listing = {
  id: string;
  categoryId: string;
  primaryLocalityId: string;
  title: string;
  description: string;
  priceType: "fixed" | "hourly" | "daily" | "quote" | "negotiable";
  priceMinor: number | null;
  currency: "EUR";
  travelsToCustomer: boolean;
  receivesCustomer: boolean;
  remoteServices: boolean;
  state: string;
  revision: number;
  createdAt: string;
  updatedAt: string;
};
export type ListingDashboardCopy = {
  title: string;
  description: string;
  newListing: string;
  create: string;
  submit: string;
  pause: string;
  archive: string;
  loading: string;
  error: string;
  retry: string;
  empty: string;
  saved: string;
  titleLabel: string;
  descriptionLabel: string;
  categoryLabel: string;
  localityLabel: string;
  priceLabel: string;
};
export function ListingDashboard({
  copy,
  categories,
  localities,
  locale = "pt-PT",
}: {
  copy: ListingDashboardCopy;
  categories: Ref[];
  localities: Ref[];
  locale?: "pt-PT" | "en" | "es";
}) {
  const [listings, setListings] = useState<Listing[] | null>(null),
    [categoryRefs, setCategoryRefs] = useState(categories),
    [localityRefs, setLocalityRefs] = useState(localities),
    [failed, setFailed] = useState(false),
    [saving, setSaving] = useState(false),
    [creating, setCreating] = useState(false);
  const generation = useRef(0);
  const [draft, setDraft] = useState(() => defaults(categories, localities));
  async function load() {
    const current = ++generation.current;
    setFailed(false);
    try {
      const values = await fetchDashboardState(locale);
      if (current === generation.current) {
        setListings(values.listings.listings);
        setCategoryRefs(values.categories.categories);
        setLocalityRefs(values.eligibleLocalities);
        setDraft((previous) =>
          previous.categoryId && previous.primaryLocalityId
            ? previous
            : defaults(values.categories.categories, values.eligibleLocalities),
        );
      }
    } catch {
      if (current === generation.current) setFailed(true);
    }
  }
  useEffect(() => {
    const current = ++generation.current;
    async function loadInitialDashboard() {
      try {
        const values = await fetchDashboardState(locale);
        if (current !== generation.current) return;
        setListings(values.listings.listings);
        setCategoryRefs(values.categories.categories);
        setLocalityRefs(values.eligibleLocalities);
        setDraft((previous) =>
          previous.categoryId && previous.primaryLocalityId
            ? previous
            : defaults(values.categories.categories, values.eligibleLocalities),
        );
      } catch {
        if (current === generation.current) setFailed(true);
      }
    }
    void loadInitialDashboard();
    return () => {
      generation.current += 1;
    };
  }, [locale]);
  async function submit(id: string, revision: number) {
    if (saving) return;
    const current = ++generation.current;
    setSaving(true);
    setFailed(false);
    try {
      const r = await fetch(`/api/v1/me/listings/${id}/submit`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ revision }),
      });
      const v: unknown = await r.json();
      if (!r.ok || !validListing(v)) throw new Error();
      if (current === generation.current)
        setListings(
          (prev) => prev?.map((item) => (item.id === id ? v : item)) ?? null,
        );
    } catch {
      if (current === generation.current) setFailed(true);
    } finally {
      if (current === generation.current) setSaving(false);
    }
  }
  async function transition(item: Listing, action: "pause" | "archive") {
    if (saving) return;
    const current = ++generation.current;
    setSaving(true);
    setFailed(false);
    try {
      const body =
        action === "pause"
          ? { revision: item.revision }
          : { revision: item.revision, state: item.state };
      const response = await fetch(`/api/v1/me/listings/${item.id}/${action}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });
      const value: unknown = await response.json();
      if (!response.ok || !validListing(value)) throw new Error();
      if (current === generation.current)
        setListings(
          (previous) =>
            previous?.map((listing) =>
              listing.id === item.id ? value : listing,
            ) ?? null,
        );
    } catch {
      if (current === generation.current) setFailed(true);
    } finally {
      if (current === generation.current) setSaving(false);
    }
  }
  async function create(e: FormEvent) {
    e.preventDefault();
    if (saving) return;
    const current = ++generation.current;
    setSaving(true);
    setFailed(false);
    try {
      const r = await fetch("/api/v1/me/listings", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(draft),
      });
      const v: unknown = await r.json();
      if (!r.ok || !validListing(v)) throw new Error();
      if (current === generation.current) {
        setListings((prev) => [...(prev ?? []), v]);
        setCreating(false);
      }
    } catch {
      if (current === generation.current) setFailed(true);
    } finally {
      if (current === generation.current) setSaving(false);
    }
  }
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
    <section aria-labelledby="listings-title">
      <h1 id="listings-title" className="text-4xl font-bold tracking-[-0.05em]">
        {copy.title}
      </h1>
      <p className="mt-3 text-muted">{copy.description}</p>
      {failed ? (
        <p role="alert" className="mt-4 text-earth">
          {copy.error}
        </p>
      ) : null}
      <button
        type="button"
        className="market-button mt-5"
        onClick={() => setCreating((v) => !v)}
      >
        {copy.newListing}
      </button>
      {creating ? (
        <form onSubmit={create} className="market-card mt-6 grid gap-4 p-5">
          <label>
            {copy.titleLabel}
            <input
              required
              minLength={2}
              value={draft.title}
              onChange={(e) => setDraft({ ...draft, title: e.target.value })}
            />
          </label>
          <label>
            {copy.descriptionLabel}
            <textarea
              required
              minLength={20}
              value={draft.description}
              onChange={(e) =>
                setDraft({ ...draft, description: e.target.value })
              }
            />
          </label>
          <label>
            {copy.categoryLabel}
            <select
              value={draft.categoryId}
              onChange={(e) =>
                setDraft({ ...draft, categoryId: e.target.value })
              }
            >
              {categoryRefs.map((v) => (
                <option key={v.id} value={v.id}>
                  {v.name}
                </option>
              ))}
            </select>
          </label>
          <label>
            {copy.localityLabel}
            <select
              value={draft.primaryLocalityId}
              onChange={(e) =>
                setDraft({ ...draft, primaryLocalityId: e.target.value })
              }
            >
              {localityRefs.map((v) => (
                <option key={v.id} value={v.id}>
                  {v.name}
                </option>
              ))}
            </select>
          </label>
          <label>
            {copy.priceLabel}
            <input
              type="number"
              min={1}
              value={draft.priceMinor ?? 0}
              onChange={(e) =>
                setDraft({ ...draft, priceMinor: Number(e.target.value) })
              }
            />
          </label>
          <button disabled={saving} type="submit">
            {copy.create}
          </button>
        </form>
      ) : null}
      <div className="mt-8 grid gap-4">
        {listings?.length ? (
          listings.map((item) => (
            <article key={item.id} className="market-card p-5">
              <h2 className="font-semibold">{item.title}</h2>
              <p className="mt-2 text-sm text-muted">{item.description}</p>
              <p className="mt-3 text-sm font-semibold">{item.state}</p>
              {item.state === "draft" ? (
                <button
                  disabled={saving}
                  type="button"
                  className="mt-3 min-h-11 rounded-full border border-line px-4"
                  onClick={() => void submit(item.id, item.revision)}
                >
                  {copy.submit}
                </button>
              ) : null}
              {item.state === "active" ? (
                <button
                  disabled={saving}
                  type="button"
                  className="mt-3 min-h-11 rounded-full border border-line px-4"
                  onClick={() => void transition(item, "pause")}
                >
                  {copy.pause}
                </button>
              ) : null}
              {["draft", "rejected", "active", "paused"].includes(
                item.state,
              ) ? (
                <button
                  disabled={saving}
                  type="button"
                  className="mt-3 min-h-11 rounded-full border border-line px-4"
                  onClick={() => void transition(item, "archive")}
                >
                  {copy.archive}
                </button>
              ) : null}
            </article>
          ))
        ) : (
          <p>{copy.empty}</p>
        )}
      </div>
      <p className="mt-6 text-sm text-muted">
        Media upload is unavailable until storage is configured.
      </p>
    </section>
  );
}
function defaults(categories: Ref[], localities: Ref[]) {
  return {
    categoryId: categories[0]?.id ?? "",
    primaryLocalityId: localities[0]?.id ?? "",
    title: "",
    description: "",
    priceType: "fixed" as const,
    priceMinor: 5000,
    currency: "EUR" as const,
    travelsToCustomer: true,
    receivesCustomer: false,
    remoteServices: false,
  };
}
async function fetchDashboardState(locale: "pt-PT" | "en" | "es") {
  const [listingResponse, profileResponse, categoryResponse, localityResponse] =
    await Promise.all([
      fetch("/api/v1/me/listings"),
      fetch("/api/v1/me/provider-profile"),
      fetch(`/api/v1/catalog/categories?locale=${encodeURIComponent(locale)}`),
      fetch(
        `/api/v1/reference/localities?locale=${encodeURIComponent(locale)}`,
      ),
    ]);
  const listings: unknown = await listingResponse.json(),
    profile: unknown = await profileResponse.json(),
    categories: unknown = await categoryResponse.json(),
    localities: unknown = await localityResponse.json();
  if (
    !listingResponse.ok ||
    !profileResponse.ok ||
    !categoryResponse.ok ||
    !localityResponse.ok ||
    !validList(listings) ||
    !validProfile(profile) ||
    !validCategories(categories) ||
    !validLocalities(localities)
  )
    throw new Error();
  const eligibleLocalities = localities.localities.filter((locality) =>
    profile.profile.serviceLocalityIds.includes(locality.id),
  );
  if (!eligibleLocalities.length) throw new Error();
  return { listings, categories, eligibleLocalities };
}
function exact(v: unknown, k: string[]): v is Record<string, unknown> {
  if (v === null || typeof v !== "object" || Array.isArray(v)) return false;
  const a = Object.keys(v).sort(),
    b = [...k].sort();
  return a.length === b.length && a.every((x, i) => x === b[i]);
}
function validList(v: unknown): v is { listings: Listing[] } {
  return (
    exact(v, ["listings"]) &&
    Array.isArray(v.listings) &&
    v.listings.every(validListing)
  );
}
function validCategories(v: unknown): v is { categories: Ref[] } {
  return (
    exact(v, ["categories"]) &&
    Array.isArray(v.categories) &&
    v.categories.every(
      (item) =>
        exact(item, ["id", "parentId", "slug", "name"]) &&
        typeof item.id === "string" &&
        typeof item.name === "string",
    )
  );
}
function validLocalities(v: unknown): v is { localities: Ref[] } {
  return (
    exact(v, ["localities", "attribution"]) &&
    Array.isArray(v.localities) &&
    v.localities.every(
      (item) =>
        exact(item, [
          "id",
          "slug",
          "name",
          "parishName",
          "municipalityName",
          "districtName",
        ]) &&
        typeof item.id === "string" &&
        typeof item.name === "string",
    )
  );
}
function validProfile(
  v: unknown,
): v is { profile: { serviceLocalityIds: string[] } } {
  return (
    exact(v, ["profile"]) &&
    exact(v.profile, [
      "displayName",
      "providerType",
      "bio",
      "primaryLocalityId",
      "serviceLocalityIds",
      "maxTravelDistanceKm",
      "travelsToCustomer",
      "receivesCustomer",
      "remoteServices",
      "languageCodes",
      "createdAt",
      "updatedAt",
    ]) &&
    Array.isArray(v.profile.serviceLocalityIds) &&
    v.profile.serviceLocalityIds.every((id) => typeof id === "string")
  );
}
function validListing(v: unknown): v is Listing {
  return (
    exact(v, [
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
    typeof (v as Record<string, unknown>).id === "string" &&
    typeof (v as Record<string, unknown>).title === "string" &&
    typeof (v as Record<string, unknown>).state === "string"
  );
}
