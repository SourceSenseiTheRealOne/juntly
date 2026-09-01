import { requestListingPromotion } from "@/shared/api/generated";
import {
  authorizedHeaders,
  correlated,
  exact,
  fail,
  requestID,
  requestIDHeader,
  sessionToken,
  unavailable,
  uuid,
} from "@/features/messaging/protected-bff";
import { validPromotion } from "@/features/entitlements/entitlement-bff";
export const runtime = "nodejs";
export async function POST(request: Request): Promise<Response> {
  const id = requestID(request.headers),
    token = await sessionToken();
  if (!token) return fail("UNAUTHORIZED", "Unauthorized", 401, id);
  let value: unknown;
  try {
    value = JSON.parse(await request.text());
  } catch {
    return fail("INVALID_REQUEST", "Invalid request", 400, id);
  }
  if (
    !exact(value, ["listingId", "periodId"]) ||
    !uuid(value.listingId) ||
    !uuid(value.periodId)
  )
    return fail("INVALID_REQUEST", "Invalid request", 400, id);
  const baseUrl = process.env.JUNTLY_API_ORIGIN;
  if (!baseUrl) return unavailable(id);
  try {
    const up = await requestListingPromotion({
      baseUrl,
      body: { listingId: value.listingId, periodId: value.periodId },
      headers: authorizedHeaders(token, id),
    });
    if (
      up.error ||
      up.response?.status !== 201 ||
      !correlated(up.response, id) ||
      !validPromotion(up.data)
    )
      return unavailable(id);
    return Response.json(up.data, {
      status: 201,
      headers: { [requestIDHeader]: id },
    });
  } catch {
    return unavailable(id);
  }
}
