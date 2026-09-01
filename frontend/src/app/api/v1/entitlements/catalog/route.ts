import { getEntitlementCatalog } from "@/shared/api/generated";
import {
  correlated,
  requestID,
  requestIDHeader,
  unavailable,
} from "@/features/messaging/protected-bff";
import { validCatalog } from "@/features/entitlements/entitlement-bff";
export const runtime = "nodejs";
export async function GET(request: Request): Promise<Response> {
  const id = requestID(request.headers),
    baseUrl = process.env.JUNTLY_API_ORIGIN;
  if (!baseUrl) return unavailable(id);
  try {
    const up = await getEntitlementCatalog({
      baseUrl,
      headers: { [requestIDHeader]: id },
    });
    if (
      up.error ||
      !up.response?.ok ||
      !correlated(up.response, id) ||
      !validCatalog(up.data)
    )
      return unavailable(id);
    return Response.json(up.data, {
      status: 200,
      headers: { [requestIDHeader]: id },
    });
  } catch {
    return unavailable(id);
  }
}
