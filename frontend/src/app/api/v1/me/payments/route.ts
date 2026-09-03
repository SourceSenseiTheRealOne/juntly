import {
  fail,
  requestID,
  sessionToken,
} from "@/features/messaging/protected-bff";
import { validPaymentOrders } from "@/features/payments/payment-bff";
import { forwardPayment } from "@/features/payments/payment-proxy";

export const runtime = "nodejs";

export async function GET(request: Request): Promise<Response> {
  const id = requestID(request.headers);
  const token = await sessionToken();
  if (!token) return fail("UNAUTHORIZED", "Unauthorized", 401, id);
  return forwardPayment(
    id,
    token,
    "/api/v1/me/payments",
    "GET",
    undefined,
    validPaymentOrders,
    200,
  );
}
