import { reportConversation } from "@/shared/api/generated";
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
export const runtime = "nodejs";
type RouteContext = { params: Promise<{ conversationId: string }> };
export async function POST(
  request: Request,
  context: RouteContext,
): Promise<Response> {
  const id = requestID(request.headers),
    token = await sessionToken(),
    { conversationId } = await context.params;
  if (!token) return fail("UNAUTHORIZED", "Unauthorized", 401, id);
  let value: unknown;
  try {
    value = JSON.parse(await request.text());
  } catch {
    return fail("INVALID_REQUEST", "Invalid request", 400, id);
  }
  if (
    !uuid(conversationId) ||
    !exact(
      value,
      Object.hasOwn(value as object, "messageId")
        ? ["messageId", "reason"]
        : ["reason"],
    ) ||
    typeof value.reason !== "string" ||
    value.reason.trim().length < 5 ||
    value.reason.length > 500 ||
    ("messageId" in value && value.messageId !== null && !uuid(value.messageId))
  )
    return fail("INVALID_REQUEST", "Invalid request", 400, id);
  const baseUrl = process.env.JUNTLY_API_ORIGIN;
  if (!baseUrl) return unavailable(id);
  try {
    const upstream = await reportConversation({
      baseUrl,
      path: { conversationId },
      body: {
        reason: value.reason.trim(),
        ...("messageId" in value
          ? { messageId: value.messageId as string | null }
          : {}),
      },
      headers: authorizedHeaders(token, id),
    });
    if (
      upstream.error ||
      upstream.response?.status !== 201 ||
      !correlated(upstream.response, id)
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
