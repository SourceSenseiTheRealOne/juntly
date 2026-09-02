"use client";

import { useEffect, useState } from "react";

type ListingOption = {
  id: string;
  title: string;
  localityName?: string;
  state?: string;
};

export function AvailableListingSelect({
  emptyLabel,
  label,
  loadingLabel,
  locale,
  name,
  placeholder,
  scope,
  required = true,
}: {
  emptyLabel: string;
  label: string;
  loadingLabel: string;
  locale: "pt-PT" | "en" | "es";
  name: string;
  placeholder: string;
  scope: "public" | "mine";
  required?: boolean;
}) {
  const [items, setItems] = useState<ListingOption[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let active = true;
    const url =
      scope === "public"
        ? `/api/v1/discovery/listings?locale=${encodeURIComponent(locale)}`
        : "/api/v1/me/listings";
    void fetch(url)
      .then(async (response) => {
        const value: unknown = await response.json();
        if (!response.ok || !validOptions(value)) throw new Error();
        if (active) {
          setItems(
            value.listings.filter(
              (item) => scope === "public" || item.state === "active",
            ),
          );
        }
      })
      .catch(() => {
        if (active) setItems([]);
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [locale, scope]);

  return (
    <label className="grid gap-2 font-semibold">
      {label}
      <select
        className="market-control"
        disabled={loading || items.length === 0}
        name={name}
        required={required}
        defaultValue=""
      >
        <option value="" disabled>
          {loading ? loadingLabel : items.length ? placeholder : emptyLabel}
        </option>
        {items.map((item) => (
          <option key={item.id} value={item.id}>
            {item.title}
            {item.localityName ? `, ${item.localityName}` : ""}
          </option>
        ))}
      </select>
    </label>
  );
}

function validOptions(value: unknown): value is { listings: ListingOption[] } {
  return (
    value !== null &&
    typeof value === "object" &&
    !Array.isArray(value) &&
    Array.isArray((value as Record<string, unknown>).listings) &&
    (value as { listings: unknown[] }).listings.every(
      (item) =>
        item !== null &&
        typeof item === "object" &&
        !Array.isArray(item) &&
        typeof (item as Record<string, unknown>).id === "string" &&
        typeof (item as Record<string, unknown>).title === "string" &&
        ((item as Record<string, unknown>).localityName === undefined ||
          typeof (item as Record<string, unknown>).localityName === "string") &&
        ((item as Record<string, unknown>).state === undefined ||
          typeof (item as Record<string, unknown>).state === "string"),
    )
  );
}
