"use client";
import { FormEvent, useEffect, useState } from "react";
import type { Review } from "@/shared/api/generated";
export type ReviewsCopy = {
  title: string;
  description: string;
  newReview: string;
  bookingId: string;
  rating: string;
  body: string;
  submit: string;
  received: string;
  empty: string;
  response: string;
  respond: string;
  loading: string;
  error: string;
  created: string;
  verified: string;
};
export function ReviewsDashboard({ copy }: { copy: ReviewsCopy }) {
  const [items, setItems] = useState<Review[]>([]),
    [loading, setLoading] = useState(true),
    [failed, setFailed] = useState(false),
    [notice, setNotice] = useState("");
  useEffect(() => {
    let active = true;
    void fetch("/api/v1/me/reviews/provider")
      .then(async (r) => {
        const v: unknown = await r.json();
        if (!r.ok || !validList(v)) throw new Error();
        if (active) setItems(v.reviews);
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
  async function create(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const d = new FormData(e.currentTarget);
    setFailed(false);
    setNotice("");
    try {
      const r = await fetch("/api/v1/me/reviews", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            bookingId: d.get("bookingId"),
            rating: Number(d.get("rating")),
            body: d.get("body"),
          }),
        }),
        v: unknown = await r.json();
      if (!r.ok || !validReview(v)) throw new Error();
      e.currentTarget.reset();
      setNotice(copy.created);
    } catch {
      setFailed(true);
    }
  }
  async function respond(e: FormEvent<HTMLFormElement>, id: string) {
    e.preventDefault();
    const d = new FormData(e.currentTarget);
    setFailed(false);
    try {
      const r = await fetch(`/api/v1/me/reviews/${id}/response`, {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ response: d.get("response") }),
        }),
        v: unknown = await r.json();
      if (!r.ok || !validReview(v)) throw new Error();
      setItems((current) => current.map((i) => (i.id === id ? v : i)));
    } catch {
      setFailed(true);
    }
  }
  return (
    <section aria-labelledby="reviews-title">
      <h1 id="reviews-title" className="text-4xl font-bold tracking-[-0.05em]">
        {copy.title}
      </h1>
      <p className="mt-3 text-lg leading-8 text-muted">{copy.description}</p>
      {failed ? (
        <p role="alert" className="mt-6 text-earth">
          {copy.error}
        </p>
      ) : null}
      {notice ? (
        <p aria-live="polite" className="mt-6 font-semibold text-accent">
          {notice}
        </p>
      ) : null}
      <form className="market-card mt-8 grid gap-4 p-5" onSubmit={create}>
        <h2 className="text-xl font-bold">{copy.newReview}</h2>
        <Field name="bookingId" label={copy.bookingId} />
        <label className="grid gap-2 font-semibold">
          {copy.rating}
          <select className="market-control" name="rating">
            {[5, 4, 3, 2, 1].map((n) => (
              <option key={n}>{n}</option>
            ))}
          </select>
        </label>
        <label className="grid gap-2 font-semibold">
          {copy.body}
          <textarea
            className="market-control min-h-28 py-3"
            name="body"
            required
            minLength={10}
            maxLength={2000}
          />
        </label>
        <button className="market-button justify-self-start" type="submit">
          {copy.submit}
        </button>
      </form>
      <h2 className="mt-8 text-2xl font-bold">{copy.received}</h2>
      {loading ? <p className="mt-4 text-muted">{copy.loading}</p> : null}
      <div className="mt-4 grid gap-4">
        {!loading && items.length === 0 ? (
          <p className="market-card p-5 text-muted">{copy.empty}</p>
        ) : (
          items.map((item) => (
            <article className="market-card p-5" key={item.id}>
              <div className="flex items-center justify-between gap-4">
                <p className="font-bold">
                  {"★".repeat(item.rating)}
                  {"☆".repeat(5 - item.rating)}
                </p>
                {item.verifiedBooking ? (
                  <span className="market-chip">{copy.verified}</span>
                ) : null}
              </div>
              <p className="mt-3 leading-7 text-muted">{item.body}</p>
              {item.providerResponse ? (
                <p className="mt-4 rounded-xl bg-control p-4">
                  {item.providerResponse}
                </p>
              ) : (
                <form
                  className="mt-4 grid gap-3 border-t border-line pt-4"
                  onSubmit={(e) => void respond(e, item.id)}
                >
                  <label className="grid gap-2 font-semibold">
                    {copy.response}
                    <textarea
                      className="market-control min-h-20 py-3"
                      name="response"
                      required
                      minLength={3}
                      maxLength={1000}
                    />
                  </label>
                  <button
                    className="market-button-secondary justify-self-start"
                    type="submit"
                  >
                    {copy.respond}
                  </button>
                </form>
              )}
            </article>
          ))
        )}
      </div>
    </section>
  );
}
function Field({ name, label }: { name: string; label: string }) {
  return (
    <label className="grid gap-2 font-semibold">
      {label}
      <input className="market-control" name={name} required />
    </label>
  );
}
function record(v: unknown): v is Record<string, unknown> {
  return v !== null && typeof v === "object" && !Array.isArray(v);
}
function validReview(v: unknown): v is Review {
  return (
    record(v) &&
    typeof v.id === "string" &&
    typeof v.bookingId === "string" &&
    Number.isInteger(v.rating) &&
    typeof v.body === "string" &&
    typeof v.providerResponse === "string" &&
    v.verifiedBooking === true
  );
}
function validList(v: unknown): v is { reviews: Review[] } {
  return record(v) && Array.isArray(v.reviews) && v.reviews.every(validReview);
}
