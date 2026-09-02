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
  ownListing: string;
  manageListing: string;
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
  const [owned, setOwned] = useState(false);
  const [failed, setFailed] = useState(false);
  const generation = useRef(0);

  async function load() {
    const current = ++generation.current;
    setFailed(false);
    try {
      const [value, isOwned] = await Promise.all([
        fetchListing(locale, listingId),
        ownsListing(listingId),
      ]);
      if (current === generation.current) {
        setListing(value);
        setOwned(isOwned);
      }
    } catch {
      if (current === generation.current) setFailed(true);
    }
  }

  useEffect(() => {
    const current = ++generation.current;
    async function loadInitial() {
      try {
        const [value, isOwned] = await Promise.all([
          fetchListing(locale, listingId),
          ownsListing(listingId),
        ]);
        if (current === generation.current) {
          setListing(value);
          setOwned(isOwned);
        }
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
    return (
      <p className="market-empty" aria-live="polite">
        {copy.loading}
      </p>
    );
  if (failed || listing === null)
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
    <article className="grid items-start gap-5 lg:grid-cols-[minmax(0,1.35fr)_minmax(18rem,0.65fr)] lg:gap-6">
      <div className="market-panel overflow-hidden p-6 sm:p-8 lg:p-10">
        <div className="flex flex-wrap items-center gap-2">
          <span className="market-chip bg-control text-ink">
            {listing.categoryName}
          </span>
        </div>
        <div className="mt-8 max-w-3xl">
          <p className="text-sm font-medium text-muted">
            {listing.localityName}
          </p>
          <h1 className="mt-3 text-4xl font-bold tracking-[-0.055em] sm:text-5xl">
            {listing.title}
          </h1>
          <p className="mt-6 max-w-2xl text-lg leading-8 text-muted">
            {listing.description}
          </p>
        </div>
      </div>
      <aside
        className="market-panel h-fit overflow-hidden"
        aria-label={copy.provider}
      >
        <div className="border-b border-line bg-control/70 px-6 py-4">
          <p className="text-sm font-semibold text-muted">{copy.provider}</p>
        </div>
        <div className="p-6">
          <dl className="grid gap-5">
            <div>
              <dt className="text-xs font-bold tracking-[0.1em] text-muted uppercase">
                {copy.provider}
              </dt>
              <dd className="mt-2 text-lg font-bold">
                {listing.providerDisplayName}
              </dd>
            </div>
            <div className="border-t border-line pt-5">
              <dt className="text-xs font-bold tracking-[0.1em] text-muted uppercase">
                {copy.locality}
              </dt>
              <dd className="mt-2 font-semibold">{listing.localityName}</dd>
            </div>
          </dl>
          {owned ? (
            <div className="market-alert mt-6 grid justify-items-start gap-4">
              <p>{copy.ownListing}</p>
              <a
                className="market-button-secondary"
                href={`/${locale}/account/listings`}
              >
                {copy.manageListing}
              </a>
            </div>
          ) : (
            <>
              <div className="mt-6 border-t border-line pt-3">
                <ContactRevealControl
                  listingId={listing.id}
                  copy={{
                    phone: copy.phone,
                    whatsapp: copy.whatsapp,
                    error: copy.revealError,
                  }}
                />
              </div>
              <StartConversationControl
                listingId={listing.id}
                label={copy.message}
                error={copy.messageError}
                messagesUrl={`/${locale}/account/messages`}
              />
            </>
          )}
        </div>
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

async function ownsListing(listingId: string): Promise<boolean> {
  try {
    const response = await fetch("/api/v1/me/listings");
    if (!response.ok) return false;
    const value: unknown = await response.json();
    return (
      exact(value, ["listings"]) &&
      Array.isArray(value.listings) &&
      value.listings.some(
        (item) =>
          item !== null &&
          typeof item === "object" &&
          !Array.isArray(item) &&
          (item as Record<string, unknown>).id === listingId,
      )
    );
  } catch {
    return false;
  }
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
