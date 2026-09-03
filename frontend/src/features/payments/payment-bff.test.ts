import { describe, expect, it } from "vitest";

import {
  validCheckoutResult,
  validPaymentOrders,
  validPayoutAccount,
} from "./payment-bff";

const order = {
  id: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
  bookingId: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
  customerId: "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
  providerId: "dddddddd-dddd-4ddd-8ddd-dddddddddddd",
  state: "checkout_created",
  grossMinor: 12500,
  platformFeeMinor: 1250,
  providerNetMinor: 11250,
  currency: "EUR",
  createdAt: "2026-09-02T12:00:00Z",
  updatedAt: "2026-09-02T12:00:00Z",
};

describe("payment BFF contracts", () => {
  it("accepts bounded checkout, order-list, and payout projections", () => {
    expect(
      validCheckoutResult({
        order,
        url: "https://checkout.stripe.test/session",
      }),
    ).toBe(true);
    expect(validPaymentOrders({ orders: [order] })).toBe(true);
    expect(
      validPayoutAccount({
        detailsSubmitted: true,
        chargesEnabled: true,
        payoutsEnabled: true,
        updatedAt: "2026-09-02T12:00:00Z",
      }),
    ).toBe(true);
    expect(
      validPayoutAccount({
        internalUserId: order.providerId,
        stripeAccountId: "acct_test",
        detailsSubmitted: true,
        chargesEnabled: true,
        payoutsEnabled: true,
        updatedAt: "2026-09-02T12:00:00Z",
      }),
    ).toBe(false);
    expect(
      validPaymentOrders({
        orders: [{ ...order, paymentIntentId: "pi_test" }],
      }),
    ).toBe(false);
  });

  it("rejects browser-visible provider authority and inconsistent money", () => {
    expect(
      validCheckoutResult({
        order: { ...order, providerNetMinor: 12000 },
        url: "https://checkout.stripe.test/session",
      }),
    ).toBe(false);
    expect(
      validCheckoutResult({
        order,
        url: "javascript:alert(1)",
        stripeSecret: "no",
      }),
    ).toBe(false);
  });
});
