"use client";

import { useEffect, useRef, useState } from "react";
import { ContactRevealControl } from "@/features/contact/contact-reveal-control";
import { StartConversationControl } from "@/features/messaging/start-conversation-control";

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

export type PublicListingDetailCopy = {
  loading: string;
  error: string;
  retry: string;
  provider: string;
  locality: string;
  category: string;
  phone: string;
  whatsapp: string;
  revealError: string;
  message: string;
  messageError: string;
};

export function PublicListingDetail({
  copy,
  locale,
  listingId,
}: {
  copy: PublicListingDetailCopy;
  locale: "pt-PT" | "en" | "es";
  listingId: string;
}) {
  const [listing, setListing] = useState<Listing | null>(null);
  const [failed, setFailed] = useState(false);
  const generation = useRef(0);

  async function load() {
    const current = ++generation.current;
    setFailed(false);
    try {
      const value = await fetchListing(locale, listingId);
      if (current === generation.current) setListing(value);
    } catch {
      if (current === generation.current) setFailed(true);
    }
  }

  useEffect(() => {
    const current = ++generation.current;
    async function loadInitial() {
      try {
        const value = await fetchListing(locale, listingId);
        if (current === generation.current) setListing(value);
      } catch {
        if (current === generation.current) setFailed(true);
      }
    }
    void loadInitial();
    return () => {
      generation.current += 1;
    };
  }, [locale, listingId]);

  if (listing === null && !failed)
    return <p aria-live="polite">{copy.loading}</p>;
  if (failed || listing === null)
    return (
      <div role="alert">
        <p>{copy.error}</p>
        <button type="button" onClick={() => void load()}>
          {copy.retry}
        </button>
      </div>
    );

  return (
    <article className="grid gap-5 lg:grid-cols-[1.2fr_0.8fr]">
      <div className="market-panel overflow-hidden">
        <div className="min-h-44 bg-control p-6 sm:p-8">
          <span className="market-chip bg-surface text-ink">
            {listing.categoryName}
          </span>
        </div>
        <div className="p-6 sm:p-8">
          <p className="text-sm font-medium text-muted">{copy.category}</p>
          <h1 className="mt-2 text-4xl font-bold tracking-[-0.05em]">
            {listing.title}
          </h1>
          <p className="mt-5 max-w-2xl leading-7 text-muted">
            {listing.description}
          </p>
        </div>
      </div>
      <aside className="market-panel h-fit p-6" aria-label={copy.provider}>
        <dl className="grid gap-5">
          <div className="border-b border-line pb-5">
            <dt className="text-xs font-bold tracking-[0.12em] text-muted uppercase">
              {copy.provider}
            </dt>
            <dd className="mt-2 font-bold">{listing.providerDisplayName}</dd>
          </div>
          <div>
            <dt className="text-xs font-bold tracking-[0.12em] text-muted uppercase">
              {copy.locality}
            </dt>
            <dd className="mt-2 font-semibold">{listing.localityName}</dd>
          </div>
        </dl>
        <ContactRevealControl
          listingId={listing.id}
          copy={{
            phone: copy.phone,
            whatsapp: copy.whatsapp,
            error: copy.revealError,
          }}
        />
        <StartConversationControl
          listingId={listing.id}
          label={copy.message}
          error={copy.messageError}
          messagesUrl={`/${locale}/account/messages`}
        />
      </aside>
    </article>
  );
}

async function fetchListing(
  locale: "pt-PT" | "en" | "es",
  listingId: string,
): Promise<Listing> {
  if (!uuid(listingId)) throw new Error();
  const response = await fetch(
    `/api/v1/public/listings/${listingId}?locale=${encodeURIComponent(locale)}`,
  );
  const value: unknown = await response.json();
  if (!response.ok || !validListing(value)) throw new Error();
  return value;
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
    ["individual", "professional", "business"].includes(
      String(item.providerType),
    ) &&
    typeof item.promoted === "boolean" &&
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
