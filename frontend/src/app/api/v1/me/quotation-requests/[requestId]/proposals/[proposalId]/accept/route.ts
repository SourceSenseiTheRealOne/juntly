import { acceptQuotationProposal } from "@/shared/api/generated";
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
import { validQuotationProposal } from "@/features/quotations/quotation-bff";
export const runtime = "nodejs";
type Context = { params: Promise<{ requestId: string; proposalId: string }> };
export async function POST(
  request: Request,
  context: Context,
): Promise<Response> {
  const id = requestID(request.headers),
    token = await sessionToken(),
    { requestId, proposalId } = await context.params;
  if (!token) return fail("UNAUTHORIZED", "Unauthorized", 401, id);
  if (!uuid(requestId) || !uuid(proposalId))
    return fail("INVALID_REQUEST", "Invalid request", 400, id);
  const baseUrl = process.env.JUNTLY_API_ORIGIN;
  if (!baseUrl) return unavailable(id);
  try {
    const upstream = await acceptQuotationProposal({
      baseUrl,
      path: { requestId, proposalId },
      headers: authorizedHeaders(token, id),
    });
    if (
      upstream.error ||
      !upstream.response?.ok ||
      !correlated(upstream.response, id) ||
      !validQuotationProposal(upstream.data)
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
