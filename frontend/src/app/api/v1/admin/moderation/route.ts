import { moderateAdministrativeTarget } from "@/shared/api/generated";
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
export async function POST(request: Request): Promise<Response> {
  const id = requestID(request.headers),
    token = await sessionToken();
  if (!token) return fail("UNAUTHORIZED", "Unauthorized", 401, id);
  let v: unknown;
  try {
    v = JSON.parse(await request.text());
  } catch {
    return fail("INVALID_REQUEST", "Invalid request", 400, id);
  }
  if (
    !exact(v, ["kind", "targetId", "reason"]) ||
    !["hide_review", "publish_review", "resolve_report"].includes(
      String(v.kind),
    ) ||
    !uuid(v.targetId) ||
    typeof v.reason !== "string"
  )
    return fail("INVALID_REQUEST", "Invalid request", 400, id);
  const baseUrl = process.env.JUNTLY_API_ORIGIN;
  if (!baseUrl) return unavailable(id);
  try {
    const up = await moderateAdministrativeTarget({
      baseUrl,
      body: {
        kind: v.kind as "hide_review" | "publish_review" | "resolve_report",
        targetId: v.targetId,
        reason: v.reason,
      },
      headers: authorizedHeaders(token, id),
    });
    if (up.error || up.response?.status !== 204 || !correlated(up.response, id))
      return unavailable(id);
    return new Response(null, {
      status: 204,
      headers: { [requestIDHeader]: id },
    });
  } catch {
    return unavailable(id);
  }
}
