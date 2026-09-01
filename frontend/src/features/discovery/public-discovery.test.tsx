import { render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { PublicDiscovery } from "./public-discovery";

const copy = {
  title: "Encontrar serviços",
  description: "Encontre anúncios ativos perto de si.",
  loading: "A procurar serviços…",
  empty: "Não encontrámos serviços.",
  error: "Não foi possível procurar serviços.",
  retry: "Tentar novamente",
  searchLabel: "Procurar",
  searchButton: "Procurar",
  categoryLabel: "Categoria",
  localityLabel: "Localidade",
  radiusLabel: "Raio",
  priceLabel: "Tipo de preço",
  modeLabel: "Modo de serviço",
  details: "Ver anúncio",
  marketplaceLabel: "Marketplace Juntly",
  locationContextLabel: "Portugal",
  filtersLabel: "Filtros disponíveis",
};

afterEach(() => vi.restoreAllMocks());

describe("PublicDiscovery", () => {
  it("renders a text-first closed public card from the same-origin BFF", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL) => {
        expect(String(input)).toContain(
          "/api/v1/discovery/listings?locale=pt-PT",
        );
        return Response.json({
          listings: [
            {
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
            },
          ],
        });
      }),
    );
    render(<PublicDiscovery copy={copy} locale="pt-PT" />);
    await waitFor(() =>
      expect(
        screen.getByRole("heading", { name: "Canalização local" }),
      ).toBeInTheDocument(),
    );
    expect(screen.getByText("Marketplace Juntly")).toBeInTheDocument();
    expect(screen.getByText("Portugal")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Ver anúncio" })).toHaveAttribute(
      "href",
      "/pt-PT/listings/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
    );
    expect(
      screen.queryByText(/internalUserId|objectReference|clerkSubject/),
    ).not.toBeInTheDocument();
  });
});
