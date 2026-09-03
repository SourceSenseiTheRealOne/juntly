import {
  exact,
  fail,
  requestID,
  sessionToken,
} from "@/features/messaging/protected-bff";
import {
  validPayoutAccount,
  validPayoutOnboarding,
} from "@/features/payments/payment-bff";
import { forwardPayment } from "@/features/payments/payment-proxy";

export const runtime = "nodejs";

export async function GET(request: Request): Promise<Response> {
  const id = requestID(request.headers);
  const token = await sessionToken();
  if (!token) return fail("UNAUTHORIZED", "Unauthorized", 401, id);
  return forwardPayment(
    id,
    token,
    "/api/v1/me/payout-account",
    "GET",
    undefined,
    validPayoutAccount,
    200,
  );
}

export async function POST(request: Request): Promise<Response> {
  const id = requestID(request.headers);
  const token = await sessionToken();
  if (!token) return fail("UNAUTHORIZED", "Unauthorized", 401, id);
  let body: unknown;
  try {
    body = JSON.parse(await request.text());
  } catch {
    body = null;
  }
  if (
    !exact(body, ["locale"]) ||
    !["pt-PT", "en", "es"].includes(String(body.locale))
  ) {
    return fail("INVALID_REQUEST", "Invalid request", 400, id);
  }
  return forwardPayment(
    id,
    token,
    "/api/v1/me/payout-account",
    "POST",
    JSON.stringify(body),
    validPayoutOnboarding,
    200,
  );
}
