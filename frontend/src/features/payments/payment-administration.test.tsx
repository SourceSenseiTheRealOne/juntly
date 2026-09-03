import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { PaymentAdministration } from "./payment-administration";

const copy = {
  title: "Pagamentos e disputas",
  description: "Acompanhe pagamentos protegidos.",
  loading: "A carregar pagamentos…",
  error: "Não foi possível carregar os pagamentos.",
  empty: "Não há pagamentos.",
  booking: "Reserva",
  gross: "Total",
  fee: "Taxa Vila",
  providerNet: "Prestador",
  refund: "Reembolsar",
  refundConfirm: "Confirmar reembolso total?",
  refundPending: "Reembolso enviado.",
  state: "Estado",
};
const paid = {
  id: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
  bookingId: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
  customerId: "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
  providerId: "dddddddd-dddd-4ddd-8ddd-dddddddddddd",
  state: "paid",
  grossMinor: 12500,
  platformFeeMinor: 1250,
  providerNetMinor: 11250,
  currency: "EUR",
  createdAt: "2026-09-02T12:00:00Z",
  updatedAt: "2026-09-02T12:00:00Z",
};

describe("PaymentAdministration", () => {
  afterEach(() => vi.restoreAllMocks());

  it("lists durable amounts and requires confirmation before a full refund", async () => {
    vi.stubGlobal(
      "confirm",
      vi.fn(() => true),
    );
    const fetch = vi
      .fn()
      .mockResolvedValueOnce(Response.json({ orders: [paid] }))
      .mockResolvedValueOnce(
        Response.json({ ...paid, state: "refund_pending" }),
      );
    vi.stubGlobal("fetch", fetch);
    render(<PaymentAdministration copy={copy} locale="pt-PT" />);

    expect(await screen.findByText(/125,00/)).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: copy.refund }));
    await waitFor(() => expect(fetch).toHaveBeenCalledTimes(2));
    expect(confirm).toHaveBeenCalledWith(copy.refundConfirm);
    expect(fetch).toHaveBeenLastCalledWith(
      `/api/v1/admin/payments/${paid.id}/refund`,
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ idempotencyKey: `refund-${paid.id}` }),
      }),
    );
    expect(await screen.findByText(copy.refundPending)).toBeInTheDocument();
  });
});
