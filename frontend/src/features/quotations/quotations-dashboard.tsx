"use client";

import { FormEvent, useEffect, useState } from "react";
import type {
  Category,
  Locality,
  QuotationProposal,
  QuotationRequest,
} from "@/shared/api/generated";

export type QuotationsCopy = {
  title: string;
  description: string;
  customerRequests: string;
  opportunities: string;
  newRequest: string;
  requestTitle: string;
  requestDescription: string;
  category: string;
  locality: string;
  budget: string;
  deadline: string;
  publish: string;
  emptyRequests: string;
  emptyOpportunities: string;
  viewProposals: string;
  proposals: string;
  proposalPrice: string;
  proposalMessage: string;
  availableAt: string;
  estimatedMinutes: string;
  submitProposal: string;
  accept: string;
  loading: string;
  error: string;
  created: string;
  submitted: string;
  accepted: string;
};
export function QuotationsDashboard({
  locale,
  copy,
}: {
  locale: "pt-PT" | "en" | "es";
  copy: QuotationsCopy;
}) {
  const [requests, setRequests] = useState<QuotationRequest[]>([]),
    [opportunities, setOpportunities] = useState<QuotationRequest[]>([]),
    [categories, setCategories] = useState<Category[]>([]),
    [localities, setLocalities] = useState<Locality[]>([]),
    [proposals, setProposals] = useState<QuotationProposal[]>([]),
    [selected, setSelected] = useState<string | null>(null),
    [proposalTarget, setProposalTarget] = useState<string | null>(null),
    [loading, setLoading] = useState(true),
    [failed, setFailed] = useState(false),
    [notice, setNotice] = useState("");
  useEffect(() => {
    let active = true;
    const q = `?locale=${encodeURIComponent(locale)}`;
    void Promise.all([
      fetch("/api/v1/me/quotation-requests"),
      fetch("/api/v1/me/quotation-opportunities"),
      fetch(`/api/v1/catalog/categories${q}`),
      fetch(`/api/v1/reference/localities${q}`),
    ])
      .then(async (responses) => {
        const values: unknown[] = await Promise.all(
          responses.map((r) => r.json()),
        );
        if (
          responses.some((r) => !r.ok) ||
          !requestList(values[0]) ||
          !requestList(values[1]) ||
          !categoryList(values[2]) ||
          !localityList(values[3])
        )
          throw new Error();
        if (active) {
          setRequests(values[0].requests);
          setOpportunities(values[1].requests);
          setCategories(values[2].categories);
          setLocalities(values[3].localities);
        }
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
  }, [locale]);
  async function create(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const data = new FormData(event.currentTarget),
      budget = String(data.get("budgetMinor") ?? "").trim();
    setFailed(false);
    setNotice("");
    try {
      const response = await fetch("/api/v1/me/quotation-requests", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          title: data.get("title"),
          description: data.get("description"),
          categoryId: data.get("categoryId"),
          localityId: data.get("localityId"),
          proposalDeadline: new Date(
            String(data.get("proposalDeadline")),
          ).toISOString(),
          ...(budget ? { budgetMinor: Number(budget) } : {}),
        }),
      });
      const value: unknown = await response.json();
      if (!response.ok || !requestValue(value)) throw new Error();
      setRequests((current) => [value, ...current]);
      event.currentTarget.reset();
      setNotice(copy.created);
    } catch {
      setFailed(true);
    }
  }
  async function loadProposals(requestId: string) {
    setFailed(false);
    setSelected(requestId);
    try {
      const response = await fetch(
          `/api/v1/me/quotation-requests/${requestId}/proposals`,
        ),
        value: unknown = await response.json();
      if (!response.ok || !proposalList(value)) throw new Error();
      setProposals(value.proposals);
    } catch {
      setFailed(true);
    }
  }
  async function submitProposal(
    event: FormEvent<HTMLFormElement>,
    requestId: string,
  ) {
    event.preventDefault();
    const data = new FormData(event.currentTarget),
      estimate = String(data.get("estimatedMinutes") ?? "").trim();
    setFailed(false);
    setNotice("");
    try {
      const response = await fetch(
          `/api/v1/me/quotation-requests/${requestId}/proposals`,
          {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
              priceMinor: Number(data.get("priceMinor")),
              message: data.get("message"),
              availableAt: new Date(
                String(data.get("availableAt")),
              ).toISOString(),
              ...(estimate ? { estimatedMinutes: Number(estimate) } : {}),
            }),
          },
        ),
        value: unknown = await response.json();
      if (!response.ok || !proposalValue(value)) throw new Error();
      setProposalTarget(null);
      event.currentTarget.reset();
      setNotice(copy.submitted);
    } catch {
      setFailed(true);
    }
  }
  async function accept(requestId: string, proposalId: string) {
    setFailed(false);
    try {
      const response = await fetch(
          `/api/v1/me/quotation-requests/${requestId}/proposals/${proposalId}/accept`,
          { method: "POST" },
        ),
        value: unknown = await response.json();
      if (!response.ok || !proposalValue(value)) throw new Error();
      setProposals((current) =>
        current.map((item) =>
          item.id === proposalId
            ? value
            : {
                ...item,
                state: item.state === "submitted" ? "rejected" : item.state,
              },
        ),
      );
      setRequests((current) =>
        current.map((item) =>
          item.id === requestId ? { ...item, state: "accepted" } : item,
        ),
      );
      setNotice(copy.accepted);
    } catch {
      setFailed(true);
    }
  }
  return (
    <section aria-labelledby="quotations-title">
      <div className="market-page-header">
        <h1
          id="quotations-title"
          className="text-4xl font-bold tracking-[-0.055em] sm:text-5xl"
        >
          {copy.title}
        </h1>
        <p>{copy.description}</p>
      </div>
      {loading ? <p className="market-empty mt-8">{copy.loading}</p> : null}
      {failed ? (
        <p role="alert" className="market-alert mt-6">
          {copy.error}
        </p>
      ) : null}
      {notice ? (
        <p aria-live="polite" className="market-success mt-6 font-semibold">
          {notice}
        </p>
      ) : null}
      <div className="mt-8 grid gap-6 xl:grid-cols-2">
        <div>
          <h2 className="text-2xl font-bold">{copy.customerRequests}</h2>
          <form className="market-form-section mt-4" onSubmit={create}>
            <h3 className="font-bold">{copy.newRequest}</h3>
            <Field name="title" label={copy.requestTitle} />
            <label className="grid gap-2 font-semibold">
              {copy.requestDescription}
              <textarea
                className="market-control min-h-28 py-3"
                name="description"
                required
                minLength={20}
                maxLength={4000}
              />
            </label>
            <Select
              name="categoryId"
              label={copy.category}
              items={categories}
            />
            <Select
              name="localityId"
              label={copy.locality}
              items={localities}
            />
            <Field name="budgetMinor" label={copy.budget} type="number" />
            <Field
              name="proposalDeadline"
              label={copy.deadline}
              type="datetime-local"
            />
            <button className="market-button justify-self-start" type="submit">
              {copy.publish}
            </button>
          </form>
          <div className="mt-4 grid gap-3">
            {!loading && requests.length === 0 ? (
              <p className="market-empty">{copy.emptyRequests}</p>
            ) : (
              requests.map((item) => (
                <article className="market-card p-5" key={item.id}>
                  <h3 className="font-bold">{item.title}</h3>
                  <p className="mt-2 text-sm text-muted">{item.description}</p>
                  <button
                    className="market-button-secondary mt-4"
                    type="button"
                    onClick={() => void loadProposals(item.id)}
                  >
                    {copy.viewProposals}
                  </button>
                  {selected === item.id ? (
                    <div className="mt-4 grid gap-3">
                      <h4 className="font-bold">{copy.proposals}</h4>
                      {proposals.map((proposal) => (
                        <div
                          className="rounded-2xl border border-line bg-control p-4"
                          key={proposal.id}
                        >
                          <p className="font-semibold">
                            € {(proposal.priceMinor / 100).toFixed(2)}
                          </p>
                          <p className="mt-1 text-sm text-muted">
                            {proposal.message}
                          </p>
                          {proposal.state === "submitted" ? (
                            <button
                              className="market-button mt-3"
                              type="button"
                              onClick={() => void accept(item.id, proposal.id)}
                            >
                              {copy.accept}
                            </button>
                          ) : (
                            <p className="mt-2 text-sm font-semibold text-accent">
                              {proposal.state}
                            </p>
                          )}
                        </div>
                      ))}
                    </div>
                  ) : null}
                </article>
              ))
            )}
          </div>
        </div>
        <div>
          <h2 className="text-2xl font-bold">{copy.opportunities}</h2>
          <div className="mt-4 grid gap-3">
            {!loading && opportunities.length === 0 ? (
              <p className="market-empty">{copy.emptyOpportunities}</p>
            ) : (
              opportunities.map((item) => (
                <article className="market-card p-5" key={item.id}>
                  <h3 className="font-bold">{item.title}</h3>
                  <p className="mt-2 text-sm leading-6 text-muted">
                    {item.description}
                  </p>
                  <button
                    className="market-button-secondary mt-4"
                    type="button"
                    onClick={() => setProposalTarget(item.id)}
                  >
                    {copy.submitProposal}
                  </button>
                  {proposalTarget === item.id ? (
                    <form
                      className="market-form-section mt-4"
                      onSubmit={(event) => void submitProposal(event, item.id)}
                    >
                      <Field
                        name="priceMinor"
                        label={copy.proposalPrice}
                        type="number"
                      />
                      <label className="grid gap-2 font-semibold">
                        {copy.proposalMessage}
                        <textarea
                          className="market-control min-h-24 py-3"
                          name="message"
                          required
                          minLength={5}
                          maxLength={2000}
                        />
                      </label>
                      <Field
                        name="availableAt"
                        label={copy.availableAt}
                        type="datetime-local"
                      />
                      <Field
                        name="estimatedMinutes"
                        label={copy.estimatedMinutes}
                        type="number"
                      />
                      <button
                        className="market-button justify-self-start"
                        type="submit"
                      >
                        {copy.submitProposal}
                      </button>
                    </form>
                  ) : null}
                </article>
              ))
            )}
          </div>
        </div>
      </div>
    </section>
  );
}
function Field({
  name,
  label,
  type = "text",
}: {
  name: string;
  label: string;
  type?: string;
}) {
  return (
    <label className="grid gap-2 font-semibold">
      {label}
      <input
        className="market-control"
        name={name}
        type={type}
        required={name !== "budgetMinor" && name !== "estimatedMinutes"}
        min={type === "number" ? 1 : undefined}
      />
    </label>
  );
}
function Select({
  name,
  label,
  items,
}: {
  name: string;
  label: string;
  items: Array<{ id: string; name: string }>;
}) {
  return (
    <label className="grid gap-2 font-semibold">
      {label}
      <select className="market-control" name={name} required>
        <option value=""></option>
        {items.map((item) => (
          <option value={item.id} key={item.id}>
            {item.name}
          </option>
        ))}
      </select>
    </label>
  );
}
function record(v: unknown): v is Record<string, unknown> {
  return v !== null && typeof v === "object" && !Array.isArray(v);
}
function requestValue(v: unknown): v is QuotationRequest {
  return (
    record(v) &&
    typeof v.id === "string" &&
    typeof v.title === "string" &&
    typeof v.description === "string" &&
    typeof v.state === "string"
  );
}
function requestList(v: unknown): v is { requests: QuotationRequest[] } {
  return (
    record(v) && Array.isArray(v.requests) && v.requests.every(requestValue)
  );
}
function proposalValue(v: unknown): v is QuotationProposal {
  return (
    record(v) &&
    typeof v.id === "string" &&
    typeof v.requestId === "string" &&
    Number.isInteger(v.priceMinor) &&
    typeof v.message === "string" &&
    typeof v.state === "string"
  );
}
function proposalList(v: unknown): v is { proposals: QuotationProposal[] } {
  return (
    record(v) && Array.isArray(v.proposals) && v.proposals.every(proposalValue)
  );
}
function categoryList(v: unknown): v is { categories: Category[] } {
  return (
    record(v) &&
    Array.isArray(v.categories) &&
    v.categories.every(
      (i) =>
        record(i) && typeof i.id === "string" && typeof i.name === "string",
    )
  );
}
function localityList(v: unknown): v is { localities: Locality[] } {
  return (
    record(v) &&
    Array.isArray(v.localities) &&
    v.localities.every(
      (i) =>
        record(i) && typeof i.id === "string" && typeof i.name === "string",
    )
  );
}
