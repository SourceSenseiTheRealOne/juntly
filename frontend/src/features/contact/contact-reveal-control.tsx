"use client";

import { useState } from "react";

type Channel = "phone" | "whatsapp";

export type ContactRevealCopy = {
  phone: string;
  whatsapp: string;
  error: string;
};

export function ContactRevealControl({
  copy,
  listingId,
}: {
  copy: ContactRevealCopy;
  listingId: string;
}) {
  const [reveals, setReveals] = useState<Partial<Record<Channel, string>>>({});
  const [failed, setFailed] = useState(false);
  const [pending, setPending] = useState<Channel | null>(null);
  async function reveal(channel: Channel) {
    if (pending || reveals[channel]) return;
    setFailed(false);
    setPending(channel);
    try {
      const response = await fetch(
        `/api/v1/listings/${listingId}/contact-reveals`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ channel }),
        },
      );
      const value: unknown = await response.json();
      if (!response.ok || !validReveal(value, channel)) throw new Error();
      setReveals((previous) => ({ ...previous, [channel]: value.contact }));
    } catch {
      setFailed(true);
    } finally {
      setPending(null);
    }
  }
  return (
    <section className="mt-8" aria-live="polite">
      <div className="flex flex-wrap gap-3">
        <button
          type="button"
          disabled={pending !== null || !!reveals.phone}
          onClick={() => void reveal("phone")}
        >
          {copy.phone}
        </button>
        <button
          type="button"
          disabled={pending !== null || !!reveals.whatsapp}
          onClick={() => void reveal("whatsapp")}
        >
          {copy.whatsapp}
        </button>
      </div>
      {reveals.phone ? <p className="mt-3">{reveals.phone}</p> : null}
      {reveals.whatsapp ? <p className="mt-3">{reveals.whatsapp}</p> : null}
      {failed ? (
        <p role="alert" className="mt-3 text-earth">
          {copy.error}
        </p>
      ) : null}
    </section>
  );
}

function validReveal(
  value: unknown,
  channel: Channel,
): value is { channel: Channel; contact: string } {
  if (value === null || typeof value !== "object" || Array.isArray(value))
    return false;
  const record = value as Record<string, unknown>;
  const keys = Object.keys(record).sort();
  return (
    keys.length === 2 &&
    keys[0] === "channel" &&
    keys[1] === "contact" &&
    record.channel === channel &&
    typeof record.contact === "string" &&
    /^\+[1-9][0-9]{7,14}$/.test(record.contact)
  );
}
