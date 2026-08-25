"use client";

import { useEffect, useRef, useState } from "react";

type AccountCapabilities = {
  customerEnabled: true;
  providerEnabled: boolean;
  onboardingCompletedAt: string;
};

export type AccountCapabilitiesCopy = {
  title: string;
  description: string;
  customerLabel: string;
  customerDescription: string;
  providerLabel: string;
  providerDescription: string;
  enabled: string;
  disabled: string;
  loading: string;
  saving: string;
  loadError: string;
  retry: string;
  manageProvider: string;
  manageListings: string;
};

type AccountCapabilitiesCardProps = {
  copy: AccountCapabilitiesCopy;
  providerProfileUrl?: string;
  listingsUrl?: string;
};

export function AccountCapabilitiesCard({
  copy,
  providerProfileUrl,
  listingsUrl,
}: AccountCapabilitiesCardProps) {
  const [account, setAccount] = useState<AccountCapabilities | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [failed, setFailed] = useState(false);
  const requestGeneration = useRef(0);

  async function loadAccount() {
    const generation = ++requestGeneration.current;
    setLoading(true);
    setFailed(false);

    try {
      const nextAccount = await fetchAccount();
      if (generation === requestGeneration.current) {
        setAccount(nextAccount);
      }
    } catch {
      if (generation === requestGeneration.current) {
        setFailed(true);
      }
    } finally {
      if (generation === requestGeneration.current) {
        setLoading(false);
      }
    }
  }

  useEffect(() => {
    const generation = ++requestGeneration.current;

    async function loadInitialAccount() {
      try {
        const nextAccount = await fetchAccount();
        if (generation === requestGeneration.current) {
          setAccount(nextAccount);
        }
      } catch {
        if (generation === requestGeneration.current) {
          setFailed(true);
        }
      } finally {
        if (generation === requestGeneration.current) {
          setLoading(false);
        }
      }
    }

    void loadInitialAccount();
    return () => {
      requestGeneration.current += 1;
    };
  }, []);

  async function updateProviderCapability() {
    if (!account || saving) {
      return;
    }

    const generation = ++requestGeneration.current;
    const providerEnabled = !account.providerEnabled;
    setSaving(true);
    setFailed(false);

    try {
      const response = await fetch("/api/v1/me/account", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ providerEnabled }),
      });
      const nextAccount = await parseAccountResponse(response);
      if (generation === requestGeneration.current) {
        setAccount(nextAccount);
      }
    } catch {
      if (generation === requestGeneration.current) {
        setFailed(true);
      }
    } finally {
      if (generation === requestGeneration.current) {
        setSaving(false);
      }
    }
  }

  if (loading && !account) {
    return (
      <section
        className="mt-8 rounded-2xl border border-line bg-canvas p-5 text-sm text-muted"
        aria-live="polite"
        aria-busy="true"
      >
        {copy.loading}
      </section>
    );
  }

  if (!account) {
    return (
      <section
        className="mt-8 rounded-2xl border border-line bg-canvas p-5"
        role="alert"
      >
        <p className="text-sm text-muted">{copy.loadError}</p>
        <button
          type="button"
          className="mt-4 min-h-11 rounded-full bg-ink px-5 text-sm font-semibold text-canvas outline-none focus-visible:ring-2 focus-visible:ring-focus focus-visible:ring-offset-4 focus-visible:ring-offset-canvas"
          onClick={() => void loadAccount()}
        >
          {copy.retry}
        </button>
      </section>
    );
  }

  return (
    <section
      className="mt-8 border-t border-line pt-8"
      aria-labelledby="account-capabilities-title"
    >
      <h2 id="account-capabilities-title" className="text-xl font-semibold">
        {copy.title}
      </h2>
      <p className="mt-2 leading-7 text-muted">{copy.description}</p>

      {failed ? (
        <p
          className="mt-5 rounded-xl border border-earth/40 bg-earth-soft p-4 text-sm text-ink"
          role="alert"
        >
          {copy.loadError}
        </p>
      ) : null}

      <div className="mt-6 grid gap-4">
        <div className="rounded-2xl border border-line bg-canvas p-5">
          <div className="flex items-start justify-between gap-4">
            <div>
              <h3 className="font-semibold">{copy.customerLabel}</h3>
              <p className="mt-1 text-sm leading-6 text-muted">
                {copy.customerDescription}
              </p>
            </div>
            <span className="rounded-full bg-accent-soft px-3 py-1 text-sm font-semibold text-accent">
              {copy.enabled}
            </span>
          </div>
        </div>

        <div className="rounded-2xl border border-line bg-canvas p-5">
          <div className="flex items-center justify-between gap-5">
            <div>
              <h3 id="provider-capability-label" className="font-semibold">
                {copy.providerLabel}
              </h3>
              <p
                id="provider-capability-description"
                className="mt-1 text-sm leading-6 text-muted"
              >
                {copy.providerDescription}
              </p>
              <p
                className="mt-2 text-sm font-semibold text-accent"
                aria-live="polite"
              >
                {saving
                  ? copy.saving
                  : account.providerEnabled
                    ? copy.enabled
                    : copy.disabled}
              </p>
              {account.providerEnabled && providerProfileUrl ? (
                <a
                  href={providerProfileUrl}
                  className="mt-3 inline-flex min-h-11 items-center rounded-full border border-line px-4 text-sm font-semibold text-ink outline-none focus-visible:ring-2 focus-visible:ring-focus"
                >
                  {copy.manageProvider}
                </a>
              ) : null}
              {account.providerEnabled && listingsUrl ? (
                <a
                  href={listingsUrl}
                  className="mt-3 inline-flex min-h-11 items-center rounded-full border border-line px-4 text-sm font-semibold text-ink outline-none focus-visible:ring-2 focus-visible:ring-focus"
                >
                  {copy.manageListings}
                </a>
              ) : null}
            </div>
            <button
              type="button"
              role="switch"
              aria-checked={account.providerEnabled}
              aria-labelledby="provider-capability-label"
              aria-describedby="provider-capability-description"
              disabled={saving}
              onClick={() => void updateProviderCapability()}
              className="relative min-h-11 min-w-16 shrink-0 rounded-full border border-line bg-surface p-1 transition-colors outline-none focus-visible:ring-2 focus-visible:ring-focus focus-visible:ring-offset-4 focus-visible:ring-offset-canvas disabled:cursor-wait disabled:opacity-60 data-[enabled=true]:bg-accent motion-reduce:transition-none"
              data-enabled={account.providerEnabled}
            >
              <span
                aria-hidden="true"
                className="block h-8 w-8 rounded-full bg-canvas shadow-sm transition-transform data-[enabled=true]:translate-x-5 motion-reduce:transition-none"
                data-enabled={account.providerEnabled}
              />
            </button>
          </div>
        </div>
      </div>
    </section>
  );
}

async function parseAccountResponse(
  response: Response,
): Promise<AccountCapabilities> {
  if (!response.ok) {
    throw new Error("account unavailable");
  }
  const value: unknown = await response.json();
  if (!isAccountCapabilities(value)) {
    throw new Error("invalid account response");
  }
  return value;
}

async function fetchAccount(): Promise<AccountCapabilities> {
  const response = await fetch("/api/v1/me/account", { method: "GET" });
  return parseAccountResponse(response);
}

function isAccountCapabilities(value: unknown): value is AccountCapabilities {
  if (value === null || typeof value !== "object" || Array.isArray(value)) {
    return false;
  }
  const keys = Object.keys(value).sort();
  const expected = [
    "customerEnabled",
    "onboardingCompletedAt",
    "providerEnabled",
  ];
  const candidate = value as Record<string, unknown>;
  return (
    keys.length === expected.length &&
    keys.every((key, index) => key === expected[index]) &&
    candidate.customerEnabled === true &&
    typeof candidate.providerEnabled === "boolean" &&
    typeof candidate.onboardingCompletedAt === "string" &&
    !Number.isNaN(Date.parse(candidate.onboardingCompletedAt))
  );
}
