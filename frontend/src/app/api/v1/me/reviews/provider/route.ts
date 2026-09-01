import { listMyProviderReviews } from "@/shared/api/generated";
import {
  authorizedHeaders,
  correlated,
  fail,
  requestID,
  requestIDHeader,
  sessionToken,
  unavailable,
} from "@/features/messaging/protected-bff";
import { validReviews } from "@/features/reviews/review-bff";
export const runtime = "nodejs";
export async function GET(request: Request): Promise<Response> {
  const id = requestID(request.headers),
    token = await sessionToken();
  if (!token) return fail("UNAUTHORIZED", "Unauthorized", 401, id);
  const baseUrl = process.env.JUNTLY_API_ORIGIN;
  if (!baseUrl) return unavailable(id);
  try {
    const up = await listMyProviderReviews({
      baseUrl,
      headers: authorizedHeaders(token, id),
    });
    if (
      up.error ||
      !up.response?.ok ||
      !correlated(up.response, id) ||
      !validReviews(up.data)
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
