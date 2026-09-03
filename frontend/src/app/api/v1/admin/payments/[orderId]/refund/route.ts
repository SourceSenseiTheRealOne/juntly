import { resolveSoleAdministratorSession } from "@/features/auth/sole-administrator";
import {
  exact,
  fail,
  requestID,
  uuid,
} from "@/features/messaging/protected-bff";
import { validOrder } from "@/features/payments/payment-bff";
import { forwardPayment } from "@/features/payments/payment-proxy";

export const runtime = "nodejs";
type Context = { params: Promise<{ orderId: string }> };

export async function POST(
  request: Request,
  context: Context,
): Promise<Response> {
  const id = requestID(request.headers);
  const session = await resolveSoleAdministratorSession();
  if (session.status === "unauthenticated")
    return fail("UNAUTHORIZED", "Unauthorized", 401, id);
  if (session.status === "forbidden")
    return fail("FORBIDDEN", "Forbidden", 403, id);
  if (session.status !== "authorized")
    return fail("SERVICE_UNAVAILABLE", "Service unavailable", 503, id);
  const { orderId } = await context.params;
  let body: unknown;
  try {
    body = JSON.parse(await request.text());
  } catch {
    body = null;
  }
  if (
    !uuid(orderId) ||
    !exact(body, ["idempotencyKey"]) ||
    typeof body.idempotencyKey !== "string" ||
    !/^[A-Za-z0-9._:-]{8,128}$/.test(body.idempotencyKey)
  ) {
    return fail("INVALID_REQUEST", "Invalid request", 400, id);
  }
  return forwardPayment(
    id,
    session.token,
    `/api/v1/admin/payments/${orderId}/refund`,
    "POST",
    JSON.stringify(body),
    validOrder,
    200,
  );
}
