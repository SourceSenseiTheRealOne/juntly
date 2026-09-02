import {
  listConversations,
  startListingConversation,
} from "@/shared/api/generated";
import type {
  Conversation,
  ConversationsResponse,
} from "@/shared/api/generated";
import {
  authorizedHeaders,
  correlated,
  correlatedError,
  dateTime,
  exact,
  fail,
  requestID,
  requestIDHeader,
  sessionToken,
  unavailable,
  uuid,
} from "@/features/messaging/protected-bff";

export const runtime = "nodejs";
export async function GET(request: Request): Promise<Response> {
  const id = requestID(request.headers),
    token = await sessionToken();
  if (!token) return fail("UNAUTHORIZED", "Unauthorized", 401, id);
  const baseUrl = process.env.JUNTLY_API_ORIGIN;
  if (!baseUrl) return unavailable(id);
  try {
    const upstream = await listConversations({
      baseUrl,
      headers: authorizedHeaders(token, id),
    });
    if (
      upstream.error ||
      !upstream.response?.ok ||
      !correlated(upstream.response, id) ||
      !validList(upstream.data)
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
  if (!exact(value, ["listingId"]) || !uuid(value.listingId))
    return fail("INVALID_REQUEST", "Invalid request", 400, id);
  const baseUrl = process.env.JUNTLY_API_ORIGIN;
  if (!baseUrl) return unavailable(id);
  try {
    const upstream = await startListingConversation({
      baseUrl,
      body: { listingId: value.listingId },
      headers: authorizedHeaders(token, id),
    });
    if (
      upstream.response?.status === 403 &&
      correlated(upstream.response, id) &&
      correlatedError(upstream.error, "FORBIDDEN", id)
    )
      return Response.json(upstream.error, {
        status: 403,
        headers: { [requestIDHeader]: id },
      });
    if (
      upstream.response?.status === 404 &&
      correlated(upstream.response, id) &&
      correlatedError(upstream.error, "NOT_FOUND", id)
    )
      return Response.json(upstream.error, {
        status: 404,
        headers: { [requestIDHeader]: id },
      });
    if (
      upstream.error ||
      upstream.response?.status !== 201 ||
      !correlated(upstream.response, id) ||
      !validConversation(upstream.data)
    )
      return unavailable(id);
    return Response.json(upstream.data, {
      status: 201,
      headers: { [requestIDHeader]: id },
    });
  } catch {
    return unavailable(id);
  }
}
function validList(
  value: ConversationsResponse | undefined,
): value is ConversationsResponse {
  return (
    !!value &&
    exact(value, ["conversations"]) &&
    Array.isArray(value.conversations) &&
    value.conversations.every(validConversation)
  );
}
function validConversation(
  value: Conversation | undefined,
): value is Conversation {
  return (
    !!value &&
    exact(value, [
      "id",
      "listingId",
      "customerId",
      "providerId",
      "blocked",
      "createdAt",
      "updatedAt",
    ]) &&
    uuid(value.id) &&
    (value.listingId === null || uuid(value.listingId)) &&
    uuid(value.customerId) &&
    uuid(value.providerId) &&
    typeof value.blocked === "boolean" &&
    dateTime(value.createdAt) &&
    dateTime(value.updatedAt)
  );
}
