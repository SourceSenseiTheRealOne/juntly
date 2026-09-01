import { getMyBooking } from "@/shared/api/generated";
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
import { validBooking } from "@/features/bookings/booking-bff";
export const runtime = "nodejs";
type Context = { params: Promise<{ bookingId: string }> };
export async function GET(
  request: Request,
  context: Context,
): Promise<Response> {
  const id = requestID(request.headers),
    token = await sessionToken(),
    { bookingId } = await context.params;
  if (!token) return fail("UNAUTHORIZED", "Unauthorized", 401, id);
  if (!uuid(bookingId))
    return fail("INVALID_REQUEST", "Invalid request", 400, id);
  const baseUrl = process.env.JUNTLY_API_ORIGIN;
  if (!baseUrl) return unavailable(id);
  try {
    const up = await getMyBooking({
      baseUrl,
      path: { bookingId },
      headers: authorizedHeaders(token, id),
    });
    if (
      up.error ||
      !up.response?.ok ||
      !correlated(up.response, id) ||
      !validBooking(up.data)
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
