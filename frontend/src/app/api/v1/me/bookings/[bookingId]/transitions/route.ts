import {
  transitionBooking,
  type BookingState,
  type TransitionBooking,
} from "@/shared/api/generated";
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
import { validBooking } from "@/features/bookings/booking-bff";
const states: BookingState[] = [
  "draft",
  "pending_provider_confirmation",
  "confirmed",
  "scheduled",
  "in_progress",
  "completed",
  "cancelled",
  "disputed",
  "refunded",
];
export const runtime = "nodejs";
type Context = { params: Promise<{ bookingId: string }> };
export async function POST(
  request: Request,
  context: Context,
): Promise<Response> {
  const id = requestID(request.headers),
    token = await sessionToken(),
    { bookingId } = await context.params;
  if (!token) return fail("UNAUTHORIZED", "Unauthorized", 401, id);
  const body = await readTransition(request);
  if (!uuid(bookingId) || !body)
    return fail("INVALID_REQUEST", "Invalid request", 400, id);
  const baseUrl = process.env.JUNTLY_API_ORIGIN;
  if (!baseUrl) return unavailable(id);
  try {
    const up = await transitionBooking({
      baseUrl,
      path: { bookingId },
      body,
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
async function readTransition(
  request: Request,
): Promise<TransitionBooking | null> {
  let value: unknown;
  try {
    value = JSON.parse(await request.text());
  } catch {
    return null;
  }
  const valid =
    exact(value, ["expectedState", "targetState", "revision"]) ||
    exact(value, ["expectedState", "targetState", "revision", "reason"]);
  if (!valid) return null;
  const v = value as Record<string, unknown>;
  if (
    !states.includes(v.expectedState as BookingState) ||
    !states.includes(v.targetState as BookingState) ||
    !Number.isInteger(v.revision) ||
    Number(v.revision) < 1 ||
    ("reason" in v && v.reason !== null && typeof v.reason !== "string")
  )
    return null;
  return {
    expectedState: v.expectedState as BookingState,
    targetState: v.targetState as BookingState,
    revision: Number(v.revision),
    ...("reason" in v ? { reason: v.reason as string | null } : {}),
  };
}
