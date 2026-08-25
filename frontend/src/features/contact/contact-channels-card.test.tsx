import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { ContactChannelsCard } from "./contact-channels-card";

const copy = {
  title: "Canais de contacto",
  description: "Controle como os clientes autenticados o podem contactar.",
  loading: "A carregar canais…",
  error: "Não foi possível carregar ou guardar os canais.",
  retry: "Tentar novamente",
  phone: "Telefone",
  whatsapp: "WhatsApp",
  contact: "Contacto",
  formatHint: "Use o formato internacional.",
  enabled: "Ativo",
  consent: "Autorizar revelação",
  save: "Guardar canal",
  saved: "Canal guardado.",
};

afterEach(() => vi.restoreAllMocks());

describe("ContactChannelsCard", () => {
  it("loads status-only channels and saves a private update without rendering the value", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
        if (!init?.method)
          return Response.json({
            channels: [
              {
                channel: "phone",
                configured: true,
                enabled: true,
                revealConsent: true,
              },
            ],
          });
        await expect(JSON.parse(String(init.body))).toEqual({
          channel: "phone",
          contact: "+12025550123",
          enabled: true,
          revealConsent: true,
        });
        return Response.json({
          channel: "phone",
          configured: true,
          enabled: true,
          revealConsent: true,
        });
      }),
    );
    render(<ContactChannelsCard copy={copy} />);
    await waitFor(() =>
      expect(
        screen.getByRole("combobox", { name: "Telefone" }),
      ).toBeInTheDocument(),
    );
    fireEvent.change(screen.getByRole("textbox"), {
      target: { value: "+12025550123" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Guardar canal" }));
    await waitFor(() =>
      expect(screen.getByText("Canal guardado.")).toBeInTheDocument(),
    );
    expect(screen.queryByDisplayValue("+12025550123")).not.toBeInTheDocument();
    expect(
      screen.queryByText(/ciphertext|nonce|keyVersion/i),
    ).not.toBeInTheDocument();
  });
});
