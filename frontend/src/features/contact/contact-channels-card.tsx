"use client";

import { useEffect, useRef, useState } from "react";

type Channel = "phone" | "whatsapp";
type Status = {
  channel: Channel;
  configured: boolean;
  enabled: boolean;
  revealConsent: boolean;
};

export type ContactChannelsCopy = {
  title: string;
  description: string;
  loading: string;
  error: string;
  retry: string;
  phone: string;
  whatsapp: string;
  contact: string;
  enabled: string;
  consent: string;
  save: string;
  saved: string;
};

export function ContactChannelsCard({ copy }: { copy: ContactChannelsCopy }) {
  const [statuses, setStatuses] = useState<Status[] | null>(null);
  const [failed, setFailed] = useState(false);
  const [saved, setSaved] = useState(false);
  const [channel, setChannel] = useState<Channel>("phone");
  const [contact, setContact] = useState("");
  const [enabled, setEnabled] = useState(true);
  const [consent, setConsent] = useState(true);
  const generation = useRef(0);

  async function load() {
    const current = ++generation.current;
    setFailed(false);
    try {
      const values = await fetchStatuses();
      if (current === generation.current) setStatuses(values);
    } catch {
      if (current === generation.current) setFailed(true);
    }
  }

  useEffect(() => {
    const current = ++generation.current;
    async function loadInitial() {
      try {
        const values = await fetchStatuses();
        if (current === generation.current) setStatuses(values);
      } catch {
        if (current === generation.current) setFailed(true);
      }
    }
    void loadInitial();
    return () => {
      generation.current += 1;
    };
  }, []);

  async function save(event: React.FormEvent) {
    event.preventDefault();
    setSaved(false);
    setFailed(false);
    try {
      const response = await fetch("/api/v1/me/contact-channels", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          channel,
          contact,
          enabled,
          revealConsent: consent,
        }),
      });
      const value: unknown = await response.json();
      if (!response.ok || !validStatus(value)) throw new Error();
      setStatuses((previous) => [
        ...(previous ?? []).filter((item) => item.channel !== value.channel),
        value,
      ]);
      setContact("");
      setSaved(true);
    } catch {
      setFailed(true);
    }
  }

  if (statuses === null && !failed)
    return <p aria-live="polite">{copy.loading}</p>;
  if (failed && statuses === null)
    return (
      <div role="alert">
        <p>{copy.error}</p>
        <button type="button" onClick={() => void load()}>
          {copy.retry}
        </button>
      </div>
    );
  return (
    <section aria-labelledby="contact-channels-title">
      <h1 id="contact-channels-title" className="text-3xl font-semibold">
        {copy.title}
      </h1>
      <p className="mt-3 text-muted">{copy.description}</p>
      {failed ? (
        <p role="alert" className="mt-4 text-earth">
          {copy.error}
        </p>
      ) : null}
      {saved ? (
        <p role="status" className="mt-4">
          {copy.saved}
        </p>
      ) : null}
      <ul className="mt-6 grid gap-2">
        {statuses?.map((status) => (
          <li key={status.channel}>
            {status.channel === "phone" ? copy.phone : copy.whatsapp}:{" "}
            {status.enabled && status.revealConsent ? copy.enabled : "—"}
          </li>
        ))}
      </ul>
      <form className="mt-6 grid gap-4" onSubmit={(event) => void save(event)}>
        <label>
          <span className="sr-only">{copy.phone}</span>
          <select
            value={channel}
            onChange={(event) => setChannel(event.target.value as Channel)}
          >
            <option value="phone">{copy.phone}</option>
            <option value="whatsapp">{copy.whatsapp}</option>
          </select>
        </label>
        <label>
          {copy.contact}
          <input
            value={contact}
            onChange={(event) => setContact(event.target.value)}
            autoComplete="tel"
          />
        </label>
        <label>
          <input
            type="checkbox"
            checked={enabled}
            onChange={(event) => setEnabled(event.target.checked)}
          />{" "}
          {copy.enabled}
        </label>
        <label>
          <input
            type="checkbox"
            checked={consent}
            onChange={(event) => setConsent(event.target.checked)}
          />{" "}
          {copy.consent}
        </label>
        <button type="submit">{copy.save}</button>
      </form>
    </section>
  );
}

async function fetchStatuses(): Promise<Status[]> {
  const response = await fetch("/api/v1/me/contact-channels");
  const value: unknown = await response.json();
  if (
    !response.ok ||
    !exact(value, ["channels"]) ||
    !Array.isArray((value as { channels: unknown }).channels) ||
    !(value as { channels: unknown[] }).channels.every(validStatus)
  )
    throw new Error();
  return (value as { channels: Status[] }).channels;
}
function validStatus(value: unknown): value is Status {
  return (
    exact(value, ["channel", "configured", "enabled", "revealConsent"]) &&
    ((value as Status).channel === "phone" ||
      (value as Status).channel === "whatsapp") &&
    typeof (value as Status).configured === "boolean" &&
    typeof (value as Status).enabled === "boolean" &&
    typeof (value as Status).revealConsent === "boolean"
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
