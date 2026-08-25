"use client";

import { useEffect, useRef, useState } from "react";
import type { FormEvent } from "react";

type Locality = {
  id: string;
  slug: string;
  name: string;
  parishName: string;
  municipalityName: string;
  districtName: string;
};
type Language = { code: string; name: string };
type Draft = {
  displayName: string;
  providerType: "individual" | "professional" | "business";
  bio: string;
  primaryLocalityId: string;
  serviceLocalityIds: string[];
  maxTravelDistanceKm: number;
  travelsToCustomer: boolean;
  receivesCustomer: boolean;
  remoteServices: boolean;
  languageCodes: string[];
};
type Profile = Draft & { createdAt: string; updatedAt: string };
export type ProviderProfileCopy = {
  title: string;
  description: string;
  displayName: string;
  providerType: string;
  individual: string;
  professional: string;
  business: string;
  bio: string;
  primaryLocality: string;
  serviceLocalities: string;
  languages: string;
  travelRadius: string;
  travels: string;
  receives: string;
  remote: string;
  save: string;
  saving: string;
  loading: string;
  error: string;
  retry: string;
  saved: string;
};

export function ProviderProfileForm({
  locale,
  copy,
}: {
  locale: "pt-PT" | "en" | "es";
  copy: ProviderProfileCopy;
}) {
  const [localities, setLocalities] = useState<Locality[]>([]),
    [languages, setLanguages] = useState<Language[]>([]),
    [draft, setDraft] = useState<Draft | null>(null),
    [loading, setLoading] = useState(true),
    [saving, setSaving] = useState(false),
    [failed, setFailed] = useState(false),
    [saved, setSaved] = useState(false),
    [attribution, setAttribution] = useState<{
      text: string;
      url: string;
    } | null>(null);
  const generation = useRef(0);
  async function load() {
    const current = ++generation.current;
    setLoading(true);
    setFailed(false);
    try {
      const {
        localities: l,
        languages: g,
        profile: p,
      } = await fetchProfileState(locale);
      if (current !== generation.current) return;
      setLocalities(l.localities);
      setLanguages(g.languages);
      setAttribution(l.attribution);
      setDraft(p.profile ?? defaults(l.localities, g.languages));
    } catch {
      if (current === generation.current) setFailed(true);
    } finally {
      if (current === generation.current) setLoading(false);
    }
  }
  useEffect(() => {
    const current = ++generation.current;
    async function loadInitialProfile() {
      try {
        const {
          localities: l,
          languages: g,
          profile: p,
        } = await fetchProfileState(locale);
        if (current !== generation.current) return;
        setLocalities(l.localities);
        setLanguages(g.languages);
        setAttribution(l.attribution);
        setDraft(p.profile ?? defaults(l.localities, g.languages));
      } catch {
        if (current === generation.current) setFailed(true);
      } finally {
        if (current === generation.current) setLoading(false);
      }
    }
    void loadInitialProfile();
    return () => {
      generation.current += 1;
    };
  }, [locale]);
  async function submit(event: FormEvent) {
    event.preventDefault();
    if (!draft || saving) return;
    const current = ++generation.current;
    setSaving(true);
    setFailed(false);
    setSaved(false);
    try {
      const response = await fetch("/api/v1/me/provider-profile", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(draft),
      });
      const value = await response.json();
      if (!response.ok || !validEnvelope(value) || value.profile === null)
        throw new Error();
      if (current === generation.current) {
        setDraft(stripTimes(value.profile));
        setSaved(true);
      }
    } catch {
      if (current === generation.current) setFailed(true);
    } finally {
      if (current === generation.current) setSaving(false);
    }
  }
  if (loading && !draft) return <p aria-live="polite">{copy.loading}</p>;
  if (!draft)
    return (
      <div role="alert">
        <p>{copy.error}</p>
        <button
          type="button"
          className="min-h-11 rounded-full bg-ink px-5 text-canvas"
          onClick={() => void load()}
        >
          {copy.retry}
        </button>
      </div>
    );
  const update = <K extends keyof Draft>(key: K, value: Draft[K]) =>
    setDraft((current) => (current ? { ...current, [key]: value } : current));
  const toggle = (values: string[], value: string) =>
    values.includes(value)
      ? values.filter((item) => item !== value)
      : [...values, value];
  return (
    <section aria-labelledby="provider-profile-title">
      <h1 id="provider-profile-title" className="text-3xl font-semibold">
        {copy.title}
      </h1>
      <p className="mt-3 text-muted">{copy.description}</p>
      {failed ? (
        <p role="alert" className="mt-4 text-sm text-earth">
          {copy.error}
        </p>
      ) : null}
      {saved ? (
        <p role="status" className="mt-4 text-sm text-accent">
          {copy.saved}
        </p>
      ) : null}
      <form onSubmit={submit} className="mt-8 grid gap-6">
        <label className="grid gap-2 font-semibold">
          {copy.displayName}
          <input
            value={draft.displayName}
            onChange={(e) => update("displayName", e.target.value)}
            required
            minLength={2}
            maxLength={100}
            className="min-h-11 rounded-xl border border-line bg-surface px-4"
          />
        </label>
        <label className="grid gap-2 font-semibold">
          {copy.providerType}
          <select
            value={draft.providerType}
            onChange={(e) =>
              update("providerType", e.target.value as Draft["providerType"])
            }
            className="min-h-11 rounded-xl border border-line bg-surface px-4"
          >
            <option value="individual">{copy.individual}</option>
            <option value="professional">{copy.professional}</option>
            <option value="business">{copy.business}</option>
          </select>
        </label>
        <label className="grid gap-2 font-semibold">
          {copy.bio}
          <textarea
            value={draft.bio}
            onChange={(e) => update("bio", e.target.value)}
            maxLength={1000}
            className="min-h-32 rounded-xl border border-line bg-surface p-4"
          />
        </label>
        <label className="grid gap-2 font-semibold">
          {copy.primaryLocality}
          <select
            value={draft.primaryLocalityId}
            onChange={(e) => {
              const id = e.target.value;
              update("primaryLocalityId", id);
              if (!draft.serviceLocalityIds.includes(id))
                update("serviceLocalityIds", [...draft.serviceLocalityIds, id]);
            }}
            className="min-h-11 rounded-xl border border-line bg-surface px-4"
          >
            {localities.map((item) => (
              <option key={item.id} value={item.id}>
                {item.name}
              </option>
            ))}
          </select>
        </label>
        <fieldset className="grid gap-3">
          <legend className="font-semibold">{copy.serviceLocalities}</legend>
          {localities.map((item) => (
            <label key={item.id} className="flex min-h-11 items-center gap-3">
              <input
                type="checkbox"
                checked={draft.serviceLocalityIds.includes(item.id)}
                disabled={item.id === draft.primaryLocalityId || saving}
                onChange={() =>
                  update(
                    "serviceLocalityIds",
                    toggle(draft.serviceLocalityIds, item.id),
                  )
                }
              />
              {item.name}
            </label>
          ))}
        </fieldset>
        <fieldset className="grid gap-3">
          <legend className="font-semibold">{copy.languages}</legend>
          {languages.map((item) => (
            <label key={item.code} className="flex min-h-11 items-center gap-3">
              <input
                type="checkbox"
                checked={draft.languageCodes.includes(item.code)}
                disabled={saving}
                onChange={() =>
                  update(
                    "languageCodes",
                    toggle(draft.languageCodes, item.code),
                  )
                }
              />
              {item.name}
            </label>
          ))}
        </fieldset>
        <label className="grid gap-2 font-semibold">
          {copy.travelRadius}
          <input
            type="number"
            min={0}
            max={200}
            value={draft.maxTravelDistanceKm}
            onChange={(e) =>
              update("maxTravelDistanceKm", Number(e.target.value))
            }
            className="min-h-11 rounded-xl border border-line bg-surface px-4"
          />
        </label>
        <fieldset className="grid gap-3">
          <legend className="sr-only">Service modes</legend>
          {[
            ["travelsToCustomer", copy.travels],
            ["receivesCustomer", copy.receives],
            ["remoteServices", copy.remote],
          ].map(([key, label]) => (
            <label key={key} className="flex min-h-11 items-center gap-3">
              <input
                type="checkbox"
                checked={draft[key as keyof Draft] as boolean}
                disabled={saving}
                onChange={(e) =>
                  update(key as "travelsToCustomer", e.target.checked)
                }
              />
              {label}
            </label>
          ))}
        </fieldset>
        {attribution ? (
          <a
            href={attribution.url}
            rel="noreferrer"
            target="_blank"
            className="text-sm text-muted underline"
          >
            {attribution.text}
          </a>
        ) : null}
        <button
          type="submit"
          disabled={saving}
          className="min-h-11 rounded-full bg-ink px-6 font-semibold text-canvas disabled:opacity-60"
        >
          {saving ? copy.saving : copy.save}
        </button>
      </form>
    </section>
  );
}
function defaults(localities: Locality[], languages: Language[]): Draft {
  const primary = localities[0].id;
  const preferred =
    languages.find((x) => x.code === "pt-PT")?.code ?? languages[0].code;
  return {
    displayName: "",
    providerType: "individual",
    bio: "",
    primaryLocalityId: primary,
    serviceLocalityIds: [primary],
    maxTravelDistanceKm: 25,
    travelsToCustomer: true,
    receivesCustomer: false,
    remoteServices: false,
    languageCodes: [preferred],
  };
}
function stripTimes(profile: Profile): Draft {
  return {
    displayName: profile.displayName,
    providerType: profile.providerType,
    bio: profile.bio,
    primaryLocalityId: profile.primaryLocalityId,
    serviceLocalityIds: [...profile.serviceLocalityIds],
    maxTravelDistanceKm: profile.maxTravelDistanceKm,
    travelsToCustomer: profile.travelsToCustomer,
    receivesCustomer: profile.receivesCustomer,
    remoteServices: profile.remoteServices,
    languageCodes: [...profile.languageCodes],
  };
}
async function fetchProfileState(locale: "pt-PT" | "en" | "es") {
  const [localityResponse, languageResponse, profileResponse] =
    await Promise.all([
      fetch(
        `/api/v1/reference/localities?locale=${encodeURIComponent(locale)}`,
      ),
      fetch(`/api/v1/reference/languages?locale=${encodeURIComponent(locale)}`),
      fetch("/api/v1/me/provider-profile"),
    ]);
  if (!localityResponse.ok || !languageResponse.ok || !profileResponse.ok)
    throw new Error();
  const localities: unknown = await localityResponse.json();
  const languages: unknown = await languageResponse.json();
  const profile: unknown = await profileResponse.json();
  if (
    !validLocalities(localities) ||
    !validLanguages(languages) ||
    !validEnvelope(profile) ||
    localities.localities.length === 0 ||
    languages.languages.length === 0
  )
    throw new Error();
  return { localities, languages, profile };
}
function exact(v: unknown, e: string[]): v is Record<string, unknown> {
  if (v === null || typeof v !== "object" || Array.isArray(v)) return false;
  const k = Object.keys(v).sort(),
    w = [...e].sort();
  return k.length === w.length && k.every((x, i) => x === w[i]);
}
function validLocalities(
  v: unknown,
): v is { localities: Locality[]; attribution: { text: string; url: string } } {
  return (
    exact(v, ["attribution", "localities"]) &&
    Array.isArray(v.localities) &&
    exact(v.attribution, ["text", "url"]) &&
    v.attribution.text === "© OpenStreetMap contributors"
  );
}
function validLanguages(v: unknown): v is { languages: Language[] } {
  return exact(v, ["languages"]) && Array.isArray(v.languages);
}
function validEnvelope(v: unknown): v is { profile: Profile | null } {
  return (
    exact(v, ["profile"]) &&
    (v.profile === null ||
      (typeof v.profile === "object" && v.profile !== null))
  );
}
