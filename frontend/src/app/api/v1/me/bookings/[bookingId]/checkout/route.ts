import {
  exact,
  fail,
  requestID,
  sessionToken,
  uuid,
} from "@/features/messaging/protected-bff";
import { validCheckoutResult } from "@/features/payments/payment-bff";
import { forwardPayment } from "@/features/payments/payment-proxy";

export const runtime = "nodejs";
type Context = { params: Promise<{ bookingId: string }> };

export async function POST(
  request: Request,
  context: Context,
): Promise<Response> {
  const id = requestID(request.headers);
  const token = await sessionToken();
  if (!token) return fail("UNAUTHORIZED", "Unauthorized", 401, id);
  const { bookingId } = await context.params;
  if (!uuid(bookingId))
    return fail("INVALID_REQUEST", "Invalid request", 400, id);
  const body = await requestJSON(request);
  if (
    !exact(body, ["idempotencyKey", "locale"]) ||
    typeof body.idempotencyKey !== "string" ||
    !/^[A-Za-z0-9._:-]{8,128}$/.test(body.idempotencyKey) ||
    !["pt-PT", "en", "es"].includes(String(body.locale))
  ) {
    return fail("INVALID_REQUEST", "Invalid request", 400, id);
  }
  return forwardPayment(
    id,
    token,
    `/api/v1/me/bookings/${bookingId}/checkout`,
    "POST",
    JSON.stringify(body),
    validCheckoutResult,
    201,
  );
}

async function requestJSON(request: Request): Promise<unknown> {
  try {
    return JSON.parse(await request.text());
  } catch {
    return null;
  }
}
