"use client";

import { useEffect, useState } from "react";

import {
  type PayoutAccount,
  validPayoutAccount,
  validPayoutOnboarding,
} from "./payment-bff";

type Copy = {
  title: string;
  description: string;
  loading: string;
  unavailable: string;
  notStarted: string;
  ready: string;
  incomplete: string;
  charges: string;
  payouts: string;
  details: string;
  start: string;
  continue: string;
};

export function PayoutDashboard({
  copy,
  locale,
}: {
  copy: Copy;
  locale: "pt-PT" | "en" | "es";
}) {
  const [account, setAccount] = useState<PayoutAccount | null>(null);
  const [loading, setLoading] = useState(true);
  const [failed, setFailed] = useState(false);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    let active = true;
    void fetch("/api/v1/me/payout-account")
      .then(async (response) => {
        if (response.status === 404) return null;
        const value: unknown = await response.json();
        if (!response.ok || !validPayoutAccount(value)) throw new Error();
        return value;
      })
      .then((value) => {
        if (active) setAccount(value);
      })
      .catch(() => {
        if (active) setFailed(true);
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  async function onboard() {
    if (saving) return;
    setSaving(true);
    setFailed(false);
    try {
      const response = await fetch("/api/v1/me/payout-account", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ locale }),
      });
      const value: unknown = await response.json();
      if (!response.ok || !validPayoutOnboarding(value)) throw new Error();
      window.location.assign(value.url);
    } catch {
      setFailed(true);
      setSaving(false);
    }
  }

  return (
    <section aria-labelledby="payout-title">
      <div className="market-page-header">
        <h1
          id="payout-title"
          className="text-4xl font-bold tracking-[-0.055em] sm:text-5xl"
        >
          {copy.title}
        </h1>
        <p>{copy.description}</p>
      </div>
      {loading ? <p className="market-empty mt-8">{copy.loading}</p> : null}
      {failed ? (
        <p className="market-alert mt-6" role="alert">
          {copy.unavailable}
        </p>
      ) : null}
      {!loading ? (
        <div className="market-card mt-8 p-6">
          <p className="font-semibold">
            {!account
              ? copy.notStarted
              : account.detailsSubmitted &&
                  account.chargesEnabled &&
                  account.payoutsEnabled
                ? copy.ready
                : copy.incomplete}
          </p>
          {account ? (
            <dl className="mt-5 grid gap-3 sm:grid-cols-3">
              <Status label={copy.details} enabled={account.detailsSubmitted} />
              <Status label={copy.charges} enabled={account.chargesEnabled} />
              <Status label={copy.payouts} enabled={account.payoutsEnabled} />
            </dl>
          ) : null}
          {!account ||
          !account.detailsSubmitted ||
          !account.chargesEnabled ||
          !account.payoutsEnabled ? (
            <button
              className="market-button mt-6"
              type="button"
              disabled={saving}
              onClick={() => void onboard()}
            >
              {account ? copy.continue : copy.start}
            </button>
          ) : null}
        </div>
      ) : null}
    </section>
  );
}

function Status({ label, enabled }: { label: string; enabled: boolean }) {
  return (
    <div>
      <dt className="text-sm text-muted">{label}</dt>
      <dd className="mt-1 font-semibold">{enabled ? "✓" : "—"}</dd>
    </div>
  );
}
