import { render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { AvailableListingSelect } from "./available-listing-select";

afterEach(() => vi.restoreAllMocks());

describe("AvailableListingSelect", () => {
  it("loads available services and submits the selected listing id", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        Response.json({
          listings: [
            {
              id: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
              title: "Consultoria React e Go",
              localityName: "Castelo Branco",
            },
          ],
        }),
      ),
    );

    render(
      <AvailableListingSelect
        emptyLabel="Sem serviços disponíveis"
        label="Serviço"
        loadingLabel="A carregar serviços…"
        locale="pt-PT"
        name="listingId"
        placeholder="Selecionar serviço"
        scope="public"
      />,
    );

    await waitFor(() =>
      expect(
        screen.getByRole("option", { name: /Consultoria React e Go/ }),
      ).toHaveValue("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"),
    );
    expect(screen.getByLabelText("Serviço")).toHaveAttribute(
      "name",
      "listingId",
    );
  });
});
