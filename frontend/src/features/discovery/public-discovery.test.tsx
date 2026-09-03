import { fireEvent, render, screen, waitFor } from "@testing-library/react";
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
  marketplaceLabel: "Marketplace Vila",
  locationContextLabel: "Portugal",
  filtersLabel: "Filtros disponíveis",
  promoted: "Promovido",
  allCategories: "Todas as categorias",
  allLocalities: "Todas as localidades",
  anyPrice: "Qualquer preço",
  anyMode: "Qualquer modo",
  priceFixed: "Preço fixo",
  priceHourly: "À hora",
  priceDaily: "Ao dia",
  priceQuote: "Sob orçamento",
  priceNegotiable: "Negociável",
  modeTravels: "Desloca-se ao cliente",
  modeReceives: "Recebe clientes",
  modeRemote: "À distância",
  applyFilters: "Aplicar filtros",
};

afterEach(() => vi.restoreAllMocks());

describe("PublicDiscovery", () => {
  it("renders a text-first closed public card from the same-origin BFF", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL) => {
        const url = String(input);
        if (url.includes("/api/v1/catalog/categories")) {
          return Response.json({ categories: [] });
        }
        if (url.includes("/api/v1/reference/localities")) {
          return Response.json({
            attribution: "© OpenStreetMap contributors",
            localities: [],
          });
        }
        expect(url).toContain("/api/v1/discovery/listings?locale=pt-PT");
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
              promoted: true,
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
    expect(screen.getByText("Marketplace Vila")).toBeInTheDocument();
    expect(screen.getByText("Portugal")).toBeInTheDocument();
    expect(screen.getByText("Promovido")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Ver anúncio" })).toHaveAttribute(
      "href",
      "/pt-PT/listings/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
    );
    expect(
      screen.queryByText(/internalUserId|objectReference|clerkSubject/),
    ).not.toBeInTheDocument();
  });

  it("applies category, locality, radius, price, and mode filters", async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.includes("catalog/categories"))
        return Response.json({
          categories: [
            {
              id: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
              parentId: null,
              slug: "plumbing",
              name: "Canalização",
            },
          ],
        });
      if (url.includes("reference/localities"))
        return Response.json({
          attribution: "© OpenStreetMap contributors",
          localities: [
            {
              id: "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
              slug: "zebreira",
              name: "Zebreira",
              parishName: "Zebreira",
              municipalityName: "Idanha-a-Nova",
              districtName: "Castelo Branco",
            },
          ],
        });
      return Response.json({ listings: [] });
    });
    vi.stubGlobal("fetch", fetchMock);
    render(<PublicDiscovery copy={copy} locale="pt-PT" />);

    fireEvent.change(await screen.findByLabelText(copy.categoryLabel), {
      target: { value: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb" },
    });
    fireEvent.change(screen.getByLabelText(copy.localityLabel), {
      target: { value: "cccccccc-cccc-4ccc-8ccc-cccccccccccc" },
    });
    fireEvent.change(screen.getByLabelText(copy.radiusLabel), {
      target: { value: "25" },
    });
    fireEvent.change(screen.getByLabelText(copy.priceLabel), {
      target: { value: "fixed" },
    });
    fireEvent.change(screen.getByLabelText(copy.modeLabel), {
      target: { value: "travels_to_customer" },
    });
    const applyButton = screen.getByRole("button", { name: copy.applyFilters });
    expect(applyButton).toHaveClass("market-button-compact");
    expect(applyButton).not.toHaveClass("w-full");
    fireEvent.click(applyButton);

    await waitFor(() =>
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining(
          "categoryId=bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
        ),
      ),
    );
    const filteredUrl = String(fetchMock.mock.calls.at(-1)?.[0]);
    expect(filteredUrl).toContain(
      "nearLocalityId=cccccccc-cccc-4ccc-8ccc-cccccccccccc",
    );
    expect(filteredUrl).toContain("radiusKm=25");
    expect(filteredUrl).toContain("priceType=fixed");
    expect(filteredUrl).toContain("serviceMode=travels_to_customer");
  });
});
