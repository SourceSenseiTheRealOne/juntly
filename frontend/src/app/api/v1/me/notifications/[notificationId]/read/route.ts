import { markNotificationRead } from "@/shared/api/generated";
import {
  authorizedHeaders,
  correlated,
  fail,
  requestID,
  requestIDHeader,
  sessionToken,
  unavailable,
  uuid,
} from "@/features/messaging/protected-bff";
export const runtime = "nodejs";
type RouteContext = { params: Promise<{ notificationId: string }> };
export async function POST(
  request: Request,
  context: RouteContext,
): Promise<Response> {
  const id = requestID(request.headers),
    token = await sessionToken(),
    { notificationId } = await context.params;
  if (!token) return fail("UNAUTHORIZED", "Unauthorized", 401, id);
  if (!uuid(notificationId))
    return fail("INVALID_REQUEST", "Invalid request", 400, id);
  const baseUrl = process.env.JUNTLY_API_ORIGIN;
  if (!baseUrl) return unavailable(id);
  try {
    const upstream = await markNotificationRead({
      baseUrl,
      path: { notificationId },
      headers: authorizedHeaders(token, id),
    });
    if (
      upstream.error ||
      !upstream.response?.ok ||
      !correlated(upstream.response, id)
    )
      return unavailable(id);
    return Response.json(upstream.data, {
      status: 200,
      headers: { [requestIDHeader]: id },
    });
  } catch {
    return unavailable(id);
  }
}
