import { fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { ListingDashboard } from "./listing-dashboard";
const copy = {
  title: "Os meus anúncios",
  description: "Crie e acompanhe anúncios privados.",
  newListing: "Novo anúncio",
  create: "Guardar rascunho",
  submit: "Enviar para revisão",
  loading: "A carregar anúncios…",
  error: "Não foi possível carregar ou guardar os anúncios.",
  retry: "Tentar novamente",
  empty: "Ainda não tem anúncios.",
  saved: "Rascunho guardado.",
  titleLabel: "Título",
  descriptionLabel: "Descrição",
  categoryLabel: "Categoria",
  localityLabel: "Localidade",
  priceLabel: "Preço em cêntimos",
};
const listing = {
  id: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
  categoryId: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
  primaryLocalityId: "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
  title: "Listing test",
  description: "Synthetic listing dashboard response description.",
  priceType: "fixed",
  priceMinor: 5000,
  currency: "EUR",
  travelsToCustomer: true,
  receivesCustomer: false,
  remoteServices: false,
  state: "draft",
  revision: 1,
  createdAt: "2026-08-24T12:00:00Z",
  updatedAt: "2026-08-24T12:00:00Z",
};
afterEach(() => vi.restoreAllMocks());
describe("ListingDashboard", () => {
  it("loads private listings and submits a draft without rendering internals", async () => {
    const fetch = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.endsWith("/api/v1/me/listings"))
        return Response.json({ listings: [listing] });
      if (url.includes("/api/v1/catalog/categories"))
        return Response.json({
          categories: [
            {
              id: listing.categoryId,
              parentId: null,
              slug: "plumbing",
              name: "Canalização",
            },
          ],
        });
      if (url.includes("/api/v1/reference/localities"))
        return Response.json({
          localities: [
            {
              id: listing.primaryLocalityId,
              slug: "zebreira",
              name: "Zebreira",
              parishName: "Zebreira",
              municipalityName: "Idanha-a-Nova",
              districtName: "Castelo Branco",
            },
          ],
          attribution: {
            text: "© OpenStreetMap contributors",
            url: "https://www.openstreetmap.org/copyright",
          },
        });
      if (url.endsWith("/submit"))
        return Response.json({
          ...listing,
          state: "pending_review",
          revision: 2,
        });
      throw new Error("unexpected");
    });
    vi.stubGlobal("fetch", fetch);
    render(
      <ListingDashboard
        copy={copy}
        categories={[{ id: listing.categoryId, name: "Canalização" }]}
        localities={[{ id: listing.primaryLocalityId, name: "Zebreira" }]}
      />,
    );
    expect(await screen.findByText(listing.title)).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: copy.submit }));
    expect(await screen.findByText("pending_review")).toBeInTheDocument();
    expect(document.body.textContent).not.toContain("objectReference");
    expect(document.body.textContent).not.toContain("internalUserId");
  });
  it("shows safe retry on load failure", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockRejectedValue(new Error("http://internal-api private")),
    );
    render(<ListingDashboard copy={copy} categories={[]} localities={[]} />);
    expect(await screen.findByText(copy.error)).toBeInTheDocument();
    expect(document.body.textContent).not.toContain("internal-api");
  });
});
