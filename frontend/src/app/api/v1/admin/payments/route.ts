import { resolveSoleAdministratorSession } from "@/features/auth/sole-administrator";
import { fail, requestID } from "@/features/messaging/protected-bff";
import { validPaymentOrders } from "@/features/payments/payment-bff";
import { forwardPayment } from "@/features/payments/payment-proxy";

export const runtime = "nodejs";

export async function GET(request: Request): Promise<Response> {
  const id = requestID(request.headers);
  const session = await resolveSoleAdministratorSession();
  if (session.status === "unauthenticated")
    return fail("UNAUTHORIZED", "Unauthorized", 401, id);
  if (session.status === "forbidden")
    return fail("FORBIDDEN", "Forbidden", 403, id);
  if (session.status !== "authorized")
    return fail("SERVICE_UNAVAILABLE", "Service unavailable", 503, id);
  return forwardPayment(
    id,
    session.token,
    "/api/v1/admin/payments",
    "GET",
    undefined,
    validPaymentOrders,
    200,
  );
}
