import { fireEvent, render, screen } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import { ModerationQueue } from "./moderation-queue";
const copy = {
  title: "Revisão de anúncios",
  loading: "A carregar revisão…",
  empty: "Não há anúncios pendentes.",
  error: "Não foi possível carregar a revisão.",
  retry: "Tentar novamente",
  approve: "Aprovar",
  reject: "Rejeitar",
};
const listing = {
  id: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
  title: "Listing test",
  description: "Synthetic moderation queue listing description.",
  state: "pending_review",
  revision: 1,
};
afterEach(() => vi.restoreAllMocks());
it("loads pending queue and performs approve without internals", async () => {
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.endsWith("/api/v1/moderation/listings"))
        return Response.json({
          listings: [
            {
              ...listing,
              categoryId: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
              primaryLocalityId: "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
              priceType: "fixed",
              priceMinor: 5000,
              currency: "EUR",
              travelsToCustomer: true,
              receivesCustomer: false,
              remoteServices: false,
              createdAt: "2026-08-24T12:00:00Z",
              updatedAt: "2026-08-24T12:00:00Z",
            },
          ],
        });
      if (url.endsWith("/approve"))
        return Response.json({
          ...listing,
          categoryId: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
          primaryLocalityId: "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
          priceType: "fixed",
          priceMinor: 5000,
          currency: "EUR",
          travelsToCustomer: true,
          receivesCustomer: false,
          remoteServices: false,
          createdAt: "2026-08-24T12:00:00Z",
          updatedAt: "2026-08-24T12:00:00Z",
          state: "active",
          revision: 2,
        });
      throw new Error("unexpected");
    }),
  );
  render(<ModerationQueue copy={copy} />);
  expect(await screen.findByText(listing.title)).toBeInTheDocument();
  expect(document.body.textContent).not.toContain("internalUserId");
});

it("sends a bounded reason through the reject action", async () => {
  const fetch = vi.fn(async (input: RequestInfo | URL) => {
    const url = String(input);
    const full = {
      ...listing,
      categoryId: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
      primaryLocalityId: "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
      priceType: "fixed",
      priceMinor: 5000,
      currency: "EUR",
      travelsToCustomer: true,
      receivesCustomer: false,
      remoteServices: false,
      createdAt: "2026-08-24T12:00:00Z",
      updatedAt: "2026-08-24T12:00:00Z",
    };
    if (url.endsWith("/api/v1/moderation/listings"))
      return Response.json({ listings: [full] });
    if (url.endsWith("/reject"))
      return Response.json({ ...full, state: "rejected", revision: 2 });
    throw new Error("unexpected");
  });
  vi.stubGlobal("fetch", fetch);
  render(<ModerationQueue copy={copy} />);
  await screen.findByText(listing.title);
  fireEvent.change(screen.getByLabelText(copy.reject), {
    target: { value: "Needs clearer scope" },
  });
  fireEvent.click(screen.getAllByRole("button", { name: copy.reject })[0]);
  expect(await screen.findByText(copy.empty)).toBeInTheDocument();
});
