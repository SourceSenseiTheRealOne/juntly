import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { ContactRevealControl } from "./contact-reveal-control";

const copy = {
  phone: "Revelar telefone",
  whatsapp: "Revelar WhatsApp",
  error: "Não foi possível revelar o contacto.",
};

afterEach(() => vi.restoreAllMocks());

describe("ContactRevealControl", () => {
  it("reveals contact only after an explicit authenticated BFF action", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
        expect(String(input)).toContain(
          "/api/v1/listings/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa/contact-reveals",
        );
        await expect(JSON.parse(String(init?.body))).toEqual({
          channel: "phone",
        });
        return Response.json({ channel: "phone", contact: "+12025550123" });
      }),
    );
    render(
      <ContactRevealControl
        copy={copy}
        listingId="aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
      />,
    );
    expect(screen.queryByText("+12025550123")).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Revelar telefone" }));
    await waitFor(() =>
      expect(screen.getByText("+12025550123")).toBeInTheDocument(),
    );
  });
});
