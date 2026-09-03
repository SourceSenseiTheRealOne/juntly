export type PaymentState =
  | "pending_checkout"
  | "checkout_created"
  | "processing"
  | "paid"
  | "failed"
  | "refund_pending"
  | "refunded"
  | "disputed"
  | "dispute_won"
  | "dispute_lost"
  | "cancelled";

export type PaymentOrder = {
  id: string;
  bookingId: string;
  customerId: string;
  providerId: string;
  state: PaymentState;
  grossMinor: number;
  platformFeeMinor: number;
  providerNetMinor: number;
  currency: "EUR";

  createdAt: string;
  updatedAt: string;
};

export type PayoutAccount = {
  detailsSubmitted: boolean;
  chargesEnabled: boolean;
  payoutsEnabled: boolean;
  updatedAt: string;
};

const states = new Set<PaymentState>([
  "pending_checkout",
  "checkout_created",
  "processing",
  "paid",
  "failed",
  "refund_pending",
  "refunded",
  "disputed",
  "dispute_won",
  "dispute_lost",
  "cancelled",
]);

export function validCheckoutResult(
  value: unknown,
): value is { order: PaymentOrder; url: string } {
  return (
    exact(value, ["order", "url"]) &&
    validOrder(value.order) &&
    validHTTPS(value.url)
  );
}

export function validPaymentOrders(
  value: unknown,
): value is { orders: PaymentOrder[] } {
  return (
    exact(value, ["orders"]) &&
    Array.isArray(value.orders) &&
    value.orders.every(validOrder)
  );
}

export function validPayoutAccount(value: unknown): value is PayoutAccount {
  return (
    exact(value, [
      "detailsSubmitted",
      "chargesEnabled",
      "payoutsEnabled",
      "updatedAt",
    ]) &&
    typeof value.detailsSubmitted === "boolean" &&
    typeof value.chargesEnabled === "boolean" &&
    typeof value.payoutsEnabled === "boolean" &&
    validDate(value.updatedAt)
  );
}

export function validPayoutOnboarding(
  value: unknown,
): value is { account: PayoutAccount; url: string } {
  return (
    exact(value, ["account", "url"]) &&
    validPayoutAccount(value.account) &&
    validHTTPS(value.url)
  );
}

export function validOrder(value: unknown): value is PaymentOrder {
  if (!record(value)) return false;
  const required = [
    "id",
    "bookingId",
    "customerId",
    "providerId",
    "state",
    "grossMinor",
    "platformFeeMinor",
    "providerNetMinor",
    "currency",
    "createdAt",
    "updatedAt",
  ];
  if (!allowed(value, required, [])) return false;
  return (
    required.slice(0, 4).every((key) => uuid(value[key])) &&
    typeof value.state === "string" &&
    states.has(value.state as PaymentState) &&
    Number.isInteger(value.grossMinor) &&
    Number(value.grossMinor) > 0 &&
    Number.isInteger(value.platformFeeMinor) &&
    Number(value.platformFeeMinor) >= 0 &&
    Number.isInteger(value.providerNetMinor) &&
    Number(value.providerNetMinor) ===
      Number(value.grossMinor) - Number(value.platformFeeMinor) &&
    value.currency === "EUR" &&
    validDate(value.createdAt) &&
    validDate(value.updatedAt)
  );
}

function record(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function exact(
  value: unknown,
  keys: string[],
): value is Record<string, unknown> {
  return record(value) && allowed(value, keys, []);
}

function allowed(
  value: Record<string, unknown>,
  required: string[],
  optional: string[],
): boolean {
  const actual = Object.keys(value);
  return (
    required.every((key) => actual.includes(key)) &&
    actual.every((key) => required.includes(key) || optional.includes(key))
  );
}

function uuid(value: unknown): value is string {
  return (
    typeof value === "string" &&
    /^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$/i.test(value)
  );
}

function validDate(value: unknown): value is string {
  return typeof value === "string" && !Number.isNaN(Date.parse(value));
}

function validHTTPS(value: unknown): value is string {
  if (typeof value !== "string") return false;
  try {
    return new URL(value).protocol === "https:";
  } catch {
    return false;
  }
}
