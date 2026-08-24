import { render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { PublicListingDetail } from "./public-listing-detail";

const copy = {
  loading: "A carregar anúncio…",
  error: "Não foi possível carregar este anúncio.",
  retry: "Tentar novamente",
  provider: "Prestador",
  locality: "Localidade",
  category: "Categoria",
};

afterEach(() => vi.restoreAllMocks());

describe("PublicListingDetail", () => {
  it("renders only public detail fields from the same-origin BFF", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL) => {
        expect(String(input)).toContain(
          "/api/v1/public/listings/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa?locale=pt-PT",
        );
        return Response.json({
          id: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
          title: "Canalização local",
          description:
            "Reparações domésticas de canalização para pequenos trabalhos locais.",
          categoryId: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
          categorySlug: "plumbing",
          categoryName: "Canalização",
          primaryLocalityId: "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
          localitySlug: "zebreira",
          localityName: "Zebreira",
          priceType: "fixed",
          priceMinor: 5000,
          currency: "EUR",
          travelsToCustomer: true,
          receivesCustomer: false,
          remoteServices: false,
          providerDisplayName: "Prestador local",
          providerType: "professional",
          updatedAt: "2026-08-24T12:00:00Z",
        });
      }),
    );
    render(
      <PublicListingDetail
        copy={copy}
        locale="pt-PT"
        listingId="aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
      />,
    );
    await waitFor(() =>
      expect(
        screen.getByRole("heading", { name: "Canalização local" }),
      ).toBeInTheDocument(),
    );
    expect(screen.getByText("Prestador local")).toBeInTheDocument();
    expect(
      screen.queryByText(/internalUserId|phone|email|objectReference|bio/),
    ).not.toBeInTheDocument();
  });
});
