import { getProviderRating } from "@/shared/api/generated";
import {
  correlated,
  fail,
  requestID,
  requestIDHeader,
  unavailable,
  uuid,
} from "@/features/messaging/protected-bff";
import { validRating } from "@/features/reviews/review-bff";
export const runtime = "nodejs";
type Context = { params: Promise<{ providerId: string }> };
export async function GET(
  request: Request,
  context: Context,
): Promise<Response> {
  const id = requestID(request.headers),
    { providerId } = await context.params;
  if (!uuid(providerId))
    return fail("INVALID_REQUEST", "Invalid request", 400, id);
  const baseUrl = process.env.JUNTLY_API_ORIGIN;
  if (!baseUrl) return unavailable(id);
  try {
    const up = await getProviderRating({
      baseUrl,
      path: { providerId },
      headers: { [requestIDHeader]: id },
    });
    if (
      up.error ||
      !up.response?.ok ||
      !correlated(up.response, id) ||
      !validRating(up.data)
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
