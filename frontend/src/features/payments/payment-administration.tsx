"use client";

import { useEffect, useState } from "react";

import {
  type PaymentOrder,
  validOrder,
  validPaymentOrders,
} from "./payment-bff";

export type PaymentAdministrationCopy = {
  title: string;
  description: string;
  loading: string;
  error: string;
  empty: string;
  booking: string;
  gross: string;
  fee: string;
  providerNet: string;
  refund: string;
  refundConfirm: string;
  refundPending: string;
  state: string;
};

export function PaymentAdministration({
  copy,
  locale,
}: {
  copy: PaymentAdministrationCopy;
  locale: string;
}) {
  const [orders, setOrders] = useState<PaymentOrder[] | null>(null);
  const [failed, setFailed] = useState(false);
  const [pendingID, setPendingID] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);
  useEffect(() => {
    let active = true;
    void fetch("/api/v1/admin/payments")
      .then(async (response) => {
        const value: unknown = await response.json();
        if (!response.ok || !validPaymentOrders(value)) throw new Error();
        if (active) setOrders(value.orders);
      })
      .catch(() => {
        if (active) setFailed(true);
      });
    return () => {
      active = false;
    };
  }, []);

  async function refund(order: PaymentOrder) {
    if (pendingID !== null || !confirm(copy.refundConfirm)) return;
    setPendingID(order.id);
    setFailed(false);
    setSaved(false);
    try {
      const response = await fetch(
        `/api/v1/admin/payments/${order.id}/refund`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ idempotencyKey: `refund-${order.id}` }),
        },
      );
      const value: unknown = await response.json();
      if (!response.ok || !validOrder(value) || value.id !== order.id)
        throw new Error();
      setOrders(
        (current) =>
          current?.map((item) => (item.id === value.id ? value : item)) ?? null,
      );
      setSaved(true);
    } catch {
      setFailed(true);
    } finally {
      setPendingID(null);
    }
  }

  return (
    <section aria-labelledby="payment-administration-title">
      <div className="market-page-header">
        <h1
          id="payment-administration-title"
          className="text-4xl font-bold tracking-[-0.055em] sm:text-5xl"
        >
          {copy.title}
        </h1>
        <p>{copy.description}</p>
      </div>
      {failed ? (
        <p role="alert" className="market-alert mt-6">
          {copy.error}
        </p>
      ) : null}
      {saved ? (
        <p aria-live="polite" className="market-success mt-6 font-semibold">
          {copy.refundPending}
        </p>
      ) : null}
      {orders === null && !failed ? (
        <p className="market-empty mt-6">{copy.loading}</p>
      ) : null}
      {orders?.length === 0 ? (
        <p className="market-empty mt-6">{copy.empty}</p>
      ) : null}
      <div className="mt-8 grid gap-4">
        {orders?.map((order) => {
          const refundable = ["paid", "dispute_won"].includes(order.state);
          return (
            <article className="market-card p-5 sm:p-6" key={order.id}>
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <p className="text-xs font-bold tracking-[0.08em] text-muted uppercase">
                    {copy.booking}
                  </p>
                  <p className="mt-1 font-mono text-sm break-all">
                    {order.bookingId}
                  </p>
                </div>
                <span className="market-chip">
                  {copy.state}: {order.state}
                </span>
              </div>
              <dl className="mt-5 grid gap-4 sm:grid-cols-3">
                <Amount
                  label={copy.gross}
                  value={order.grossMinor}
                  locale={locale}
                />
                <Amount
                  label={copy.fee}
                  value={order.platformFeeMinor}
                  locale={locale}
                />
                <Amount
                  label={copy.providerNet}
                  value={order.providerNetMinor}
                  locale={locale}
                />
              </dl>
              {refundable ? (
                <button
                  className="market-button market-button-compact mt-5"
                  disabled={pendingID !== null}
                  onClick={() => void refund(order)}
                  type="button"
                >
                  {copy.refund}
                </button>
              ) : null}
            </article>
          );
        })}
      </div>
    </section>
  );
}

function Amount({
  label,
  value,
  locale,
}: {
  label: string;
  value: number;
  locale: string;
}) {
  return (
    <div>
      <dt className="text-sm text-muted">{label}</dt>
      <dd className="mt-1 text-xl font-bold">
        {new Intl.NumberFormat(locale, {
          style: "currency",
          currency: "EUR",
        }).format(value / 100)}
      </dd>
    </div>
  );
}
